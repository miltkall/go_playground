from models import ValidationResult
from typing import Dict, Any, List
from datetime import datetime


def validate_imbalance_data(
    data_point: Dict[str, Any], history: List[Dict[str, Any]] = None
) -> ValidationResult:
    """
    Validate imbalance data point based on business rules

    Args:
        data_point: The data point to validate
        history: Optional list of recent data points for trend-based validation

    Returns:
        ValidationResult with validation status and details
    """
    # Extract key values
    timestamp = data_point.get("timestamp")
    value = data_point.get("value")

    # Basic validation
    if value is None:
        return ValidationResult(is_valid=False, reason="Missing value")

    if timestamp is None:
        return ValidationResult(is_valid=False, reason="Missing timestamp")

    # Range validation (imbalance data should be within realistic ranges)
    # For APG imbalance data, typical ranges are approximately -500 to 500 MW
    # but can go higher in exceptional cases
    if abs(value) > 1000:
        return ValidationResult(
            is_valid=False,
            reason=f"Value {value} outside acceptable range (-1000 to 1000)",
        )

    # Trend validation (if history provided)
    if history and len(history) > 0:
        # Sort history by timestamp to ensure proper order
        sorted_history = sorted(history, key=lambda x: x["timestamp"])
        last_value = sorted_history[-1]["value"]

        # Check for sudden extreme changes (more than 200 MW change in 5 minutes)
        # This threshold can be adjusted based on domain knowledge
        if abs(value - last_value) > 200:
            return ValidationResult(
                is_valid=False,
                reason=f"Suspicious jump from {last_value} to {value} (change: {value - last_value})",
            )

    # All validations passed
    return ValidationResult(
        is_valid=True, processed_data={"timestamp": timestamp, "value": value}
    )
