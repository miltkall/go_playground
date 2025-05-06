import requests
from datetime import datetime, timedelta
from models import ImbalanceResponse, ValueRow
from typing import List, Dict, Any, Optional

# Constants
APG_API_BASE_URL = "https://transparency.apg.at/api/v1/DRZ/Data/German/PT1M"


def format_date_for_api(dt: datetime) -> str:
    """Format datetime to APG API date format (YYYYMMDDThhmmss)"""
    return dt.strftime("%Y-%m-%dT%H%M%S")


def fetch_imbalance_data(start_date: str, end_date: str) -> ImbalanceResponse:
    """
    Fetch imbalance data from APG transparency API

    Args:
        start_date: Start date in format "YYYY-MM-DDTHHMMSS"
        end_date: End date in format "YYYY-MM-DDTHHMMSS"

    Returns:
        ImbalanceResponse object with parsed data
    """
    url = f"{APG_API_BASE_URL}/{start_date}/{end_date}"

    headers = {
        "Accept": "application/json, text/plain, */*",
        "Accept-Language": "en-US,en;q=0.9",
        "Cache-Control": "no-cache",
        "Pragma": "no-cache",
        "Referer": "https://transparency.apg.at/deltaregelzone/chart?p_drzMode=Operational&resolution=PT1M&language=German&embed=true",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
    }

    params = {"p_drzMode": "CurrentBalancingState", "resolution": "PT1M"}

    response = requests.get(url, headers=headers, params=params)
    response.raise_for_status()

    return ImbalanceResponse.model_validate(response.json())


def get_latest_data_window(window_minutes: int = 30) -> ImbalanceResponse:
    """
    Fetch the latest data window

    Args:
        window_minutes: Number of minutes to fetch (default: 30)

    Returns:
        ImbalanceResponse with latest data
    """
    end_time = datetime.now()
    start_time = end_time - timedelta(minutes=window_minutes)

    start_date = format_date_for_api(start_time)
    end_date = format_date_for_api(end_time)

    return fetch_imbalance_data(start_date, end_date)


def transform_row_to_data_point(row: ValueRow) -> Dict[str, Any]:
    """Transform a ValueRow to a standardized data point"""
    # Extract the first value (APG provides a single value per row)
    value = row.V[0].V if row.V and len(row.V) > 0 and row.V[0].V is not None else None

    return {"timestamp": row.timestamp, "value": value}


def extract_data_points(response: ImbalanceResponse) -> List[Dict[str, Any]]:
    """Extract data points from ImbalanceResponse"""
    data_points = []

    for row in response.ResponseData.ValueRows:
        if row.V and len(row.V) > 0 and row.V[0].V is not None:
            data_point = transform_row_to_data_point(row)
            data_points.append(data_point)

    return data_points
