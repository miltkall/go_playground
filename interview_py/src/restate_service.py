import uuid
from restate import Service, Context
from restate.object import VirtualObject
from datetime import datetime, timedelta
from typing import Dict, Any, List
import logging

import db_service
from models import FetchDataRequest, ProcessDataRequest, ValidationResult
import apg_data_service
import data_validator

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize Restate services
apg_etl_service = Service("APGEtlService")
data_processor = Service("DataProcessorService")
time_series_object = VirtualObject("TimeSeriesObject")
health_service = Service("HealthService")


@apg_etl_service.handler()
async def schedule_data_collection(ctx: Context, interval_minutes: int = 1):
    """Schedule periodic data collection"""
    # Schedule the first collection
    await fetch_and_process_data(ctx)

    # Schedule next run
    ctx.service_send(
        schedule_data_collection,
        arg=interval_minutes,
        send_delay=timedelta(minutes=interval_minutes),
    )


@apg_etl_service.handler()
async def fetch_and_process_data(ctx: Context):
    """Fetch data from APG and send for processing"""
    # Generate request ID for tracing
    request_id = await ctx.run("generate_id", lambda: str(uuid.uuid4()))
    logger.info(f"Starting data fetch with request ID: {request_id}")

    # Calculate time window (last 30 minutes)
    end_time = datetime.now()
    start_time = end_time - timedelta(minutes=30)

    start_date = await ctx.run(
        "format_start_date", lambda: apg_data_service.format_date_for_api(start_time)
    )

    end_date = await ctx.run(
        "format_end_date", lambda: apg_data_service.format_date_for_api(end_time)
    )

    # Fetch data
    response = await ctx.run(
        "fetch_imbalance_data",
        lambda: apg_data_service.fetch_imbalance_data(start_date, end_date),
    )

    # Extract data points
    data_points = await ctx.run(
        "extract_data_points", lambda: apg_data_service.extract_data_points(response)
    )

    logger.info(f"Fetched {len(data_points)} data points")

    # Process each data point
    for data_point in data_points:
        ctx.service_send(
            process_data_point,
            arg=ProcessDataRequest(
                metric_name="apg_imbalance", scope_name="austria", data_point=data_point
            ),
        )

    return {"request_id": request_id, "data_points_count": len(data_points)}


@data_processor.handler()
async def process_data_point(ctx: Context, request: ProcessDataRequest):
    """Process a single data point"""
    logger.info(f"Processing data point: {request.data_point}")

    # Get metric and scope
    metric = await ctx.run(
        "get_or_create_metric",
        lambda: db_service.get_or_create_metric(
            request.metric_name, f"APG imbalance data for {request.metric_name}"
        ),
    )

    scope = await ctx.run(
        "get_or_create_scope",
        lambda: db_service.get_or_create_scope(
            request.scope_name, f"Geographic scope for {request.scope_name}"
        ),
    )

    # Get recent data for validation context
    recent_data = await ctx.run(
        "get_recent_data",
        lambda: [
            {"timestamp": item.time, "value": item.data}
            for item in db_service.get_recent_data(metric.metric_id, scope.scope_id)
        ],
    )

    # Send to TimeSeriesObject for validation and storage
    await ctx.object_call(
        validate_and_store,
        key=f"{request.metric_name}_{request.scope_name}",
        arg={
            "data_point": request.data_point,
            "metric_id": str(metric.metric_id),
            "scope_id": str(scope.scope_id),
            "recent_data": recent_data,
        },
    )


@time_series_object.handler()
async def validate_and_store(ctx: Context, data: Dict[str, Any]):
    """Validate and store a data point with history context"""
    data_point = data["data_point"]
    metric_id = data["metric_id"]
    scope_id = data["scope_id"]
    recent_data = data.get("recent_data", [])

    # Get historical data from state if needed
    stored_history = await ctx.get("history") or []

    # Combine with recent data from database for better context
    validation_history = recent_data + stored_history

    # Validate data
    validation_result = await ctx.run(
        "validate_data",
        lambda: data_validator.validate_imbalance_data(data_point, validation_history),
    )

    if not validation_result.is_valid:
        logger.warning(f"Data validation failed: {validation_result.reason}")
        return {"success": False, "reason": validation_result.reason}

    # Store in database
    success = await ctx.run(
        "save_data",
        lambda: db_service.save_actual_data(
            data_point["timestamp"],
            data_point["value"],
            uuid.UUID(metric_id),
            uuid.UUID(scope_id),
        ),
    )

    if success:
        # Update history in object state
        history = await ctx.get("history") or []
        history.append(data_point)

        # Keep only the last 100 entries
        if len(history) > 100:
            history = history[-100:]

        ctx.set("history", history)

        logger.info(f"Data point saved successfully: {data_point}")
        return {"success": True, "data_point": data_point}
    else:
        logger.error(f"Failed to save data point: {data_point}")
        return {"success": False, "reason": "Database error"}


@health_service.handler()
async def health(ctx: Context):
    """Health check endpoint for Docker healthcheck"""
    return {"status": "healthy", "timestamp": datetime.now().isoformat()}


@time_series_object.handler()
async def get_history(ctx: Context) -> List[Dict[str, Any]]:
    """Get historical data points from the time series"""
    history = await ctx.get("history") or []
    return history
