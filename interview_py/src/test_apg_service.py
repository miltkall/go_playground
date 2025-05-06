import pytest
from datetime import datetime
import apg_data_service
import data_validator
from models import ValidationResult


class TestAPGDataService:
    def test_format_date_for_api(self):
        dt = datetime(2025, 5, 6, 12, 30, 0)
        formatted = apg_data_service.format_date_for_api(dt)
        assert formatted == "2025-05-06T123000"

    def test_transform_row_to_data_point(self, mocker):
        # Create a mock ValueRow with test data
        mock_row = mocker.MagicMock()
        mock_row.timestamp = datetime(2025, 5, 6, 12, 30, 0)
        mock_row.V = [mocker.MagicMock()]
        mock_row.V[0].V = 123.45

        data_point = apg_data_service.transform_row_to_data_point(mock_row)

        assert data_point["timestamp"] == datetime(2025, 5, 6, 12, 30, 0)
        assert data_point["value"] == 123.45


class TestDataValidator:
    def test_valid_data_point(self):
        data_point = {"timestamp": datetime(2025, 5, 6, 12, 30, 0), "value": 150.0}

        result = data_validator.validate_imbalance_data(data_point)

        assert result.is_valid == True
        assert result.processed_data is not None

    def test_missing_value(self):
        data_point = {"timestamp": datetime(2025, 5, 6, 12, 30, 0), "value": None}

        result = data_validator.validate_imbalance_data(data_point)

        assert result.is_valid == False
        assert "Missing value" in result.reason

    def test_out_of_range_value(self):
        data_point = {
            "timestamp": datetime(2025, 5, 6, 12, 30, 0),
            "value": 1500.0,  # Over the 1000 limit
        }

        result = data_validator.validate_imbalance_data(data_point)

        assert result.is_valid == False
        assert "outside acceptable range" in result.reason

    def test_suspicious_jump(self):
        data_point = {"timestamp": datetime(2025, 5, 6, 12, 30, 0), "value": 500.0}

        history = [{"timestamp": datetime(2025, 5, 6, 12, 25, 0), "value": 100.0}]

        result = data_validator.validate_imbalance_data(data_point, history)

        assert result.is_valid == False
        assert "Suspicious jump" in result.reason
