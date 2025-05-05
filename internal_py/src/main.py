# internal_py/src/main.py
import logging
import uuid
import restate
from pydantic import BaseModel
from typing import List, Optional

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s] [%(process)d] [%(levelname)s] - %(message)s",
)
logger = logging.getLogger(__name__)


# Simple function for test compatibility
def add_numbers(a, b):
    """Simple function to add two numbers."""
    return a + b


# ----- Models -----
class SolarProductionData(BaseModel):
    timestamp: str
    plant_id: str
    value: float


class ValueValidationResult(BaseModel):
    is_valid: bool
    reason: Optional[str] = None
    original_value: float
    adjusted_value: Optional[float] = None


# ----- First Service: Solar Production ETL -----
solar_etl_service = restate.Service("SolarETLService")


@solar_etl_service.handler()
async def process_solar_data(
    ctx: restate.Context, data: SolarProductionData
) -> ValueValidationResult:
    """Process solar production data, validate it, and log the result."""
    # Constants for validation
    MIN_VALID_VALUE = 0.0
    MAX_VALID_VALUE = 5000.0  # Example max value for a solar plant in kW

    # Generate a stable ID for idempotency
    process_id = await ctx.run("process_id", lambda: str(uuid.uuid4()))

    # Validate the data
    validation_result = await ctx.run(
        "validate_solar_data",
        lambda: validate_solar_data(data, MIN_VALID_VALUE, MAX_VALID_VALUE),
    )

    # Log the result as if saving to a database
    await ctx.run(
        "log_results", lambda: log_db_operation(process_id, data, validation_result)
    )

    return validation_result


def validate_solar_data(
    data: SolarProductionData, min_val: float, max_val: float
) -> ValueValidationResult:
    """Validate solar data against min/max bounds."""
    if data.value < min_val:
        return ValueValidationResult(
            is_valid=False,
            reason=f"Value below minimum threshold ({min_val})",
            original_value=data.value,
            adjusted_value=min_val,
        )
    elif data.value > max_val:
        return ValueValidationResult(
            is_valid=False,
            reason=f"Value above maximum threshold ({max_val})",
            original_value=data.value,
            adjusted_value=max_val,
        )
    else:
        return ValueValidationResult(is_valid=True, original_value=data.value)


def log_db_operation(
    process_id: str, data: SolarProductionData, result: ValueValidationResult
):
    """Log as if saving to a database."""
    value_to_save = (
        result.adjusted_value
        if result.adjusted_value is not None
        else result.original_value
    )

    if result.is_valid:
        logger.info(
            f"[DB_SAVE] Process ID: {process_id} - Saved valid solar data: "
            f"Plant: {data.plant_id}, Time: {data.timestamp}, Value: {value_to_save} kW"
        )
    else:
        logger.info(
            f"[DB_SAVE] Process ID: {process_id} - Saved adjusted solar data: "
            f"Plant: {data.plant_id}, Time: {data.timestamp}, Value: {value_to_save} kW "
            f"(Original: {result.original_value} kW, Reason: {result.reason})"
        )


# ----- Second Service: Time Series Validation -----
time_series_validator = restate.VirtualObject("TimeSeriesValidator")


class TimeSeriesData(BaseModel):
    timestamp: str
    metric_name: str
    value: float


@time_series_validator.handler("addValue")
async def add_value(
    ctx: restate.ObjectContext, data: TimeSeriesData
) -> ValueValidationResult:
    """Add a value to the time series and validate it against previous values."""
    # Get historical values for this metric
    history = await ctx.get(f"history_{data.metric_name}") or []

    # Validate the new value
    result = validate_time_series_value(data.value, history)

    # If valid, add to history
    if result.is_valid:
        # Add the new value to history
        history.append(data.value)

        # Keep only the latest 100 values to prevent excessive state size
        if len(history) > 100:
            history = history[-100:]

        # Save updated history
        ctx.set(f"history_{data.metric_name}", history)

        logger.info(
            f"Added valid value {data.value} for metric {data.metric_name} at {data.timestamp}. "
            f"History now contains {len(history)} values."
        )
    else:
        logger.warning(
            f"Rejected implausible value {data.value} for metric {data.metric_name} at {data.timestamp}. "
            f"Reason: {result.reason}"
        )

    return result


# @time_series_validator.handler("getHistory", kind="shared")
# async def get_history(
#     ctx: restate.ObjectSharedContext, metric_name: str
# ) -> List[float]:
#     """Get the historical values for a metric."""
#     history = await ctx.get(f"history_{metric_name}") or []
#     print(history)
#     return history


def validate_time_series_value(
    value: float, history: List[float]
) -> ValueValidationResult:
    """Validate a time series value against its history."""
    # If no history, accept the value
    if not history:
        return ValueValidationResult(is_valid=True, original_value=value)

    # Get the most recent value
    last_value = history[-1]

    # Check for implausible jumps (more than 100)
    if abs(value - last_value) > 100:
        return ValueValidationResult(
            is_valid=False,
            reason=f"Value jump too large: {abs(value - last_value)} from previous value {last_value}",
            original_value=value,
        )

    return ValueValidationResult(is_valid=True, original_value=value)


# Create the application
app = restate.app([solar_etl_service, time_series_validator])

# Main entry point
if __name__ == "__main__":
    import hypercorn
    import asyncio

    conf = hypercorn.Config()
    conf.bind = ["0.0.0.0:9080"]
    asyncio.run(hypercorn.asyncio.serve(app, conf))
