from pydantic import UUID7, BaseModel
from typing import List, Optional
import requests
from datetime import datetime
from sqlmodel import SQLModel
from uuid import UUID, uuid4


# CREATE TABLE public.actual (
#     "time" timestamp without time zone NOT NULL,
#     data double precision NOT NULL,
#     metric_id uuid NOT NULL REFERENCES public.metric (metric_id),
#     scope_id uuid NOT NULL REFERENCES public.scope (scope_id),


#     PRIMARY KEY ("time", metric_id, scope_id)
# );


class ValueItem(BaseModel):
    V: Optional[float] = None
    E: bool
    M: bool


class ValueRow(BaseModel):
    DF: str  # Date From ++
    TF: str  # Time From ++
    DT: str  # Date To
    TT: str  # Time To
    V: ValueItem

    @property
    def minutes_from(self) -> int:
        """Convert TF time to minutes of the day"""
        hours, minutes = map(int, self.TF.split(":"))
        return hours * 60 + minutes

    @property
    def minutes_to(self) -> int:
        """Convert TT time to minutes of the day"""
        hours, minutes = map(int, self.TT.split(":"))
        return hours * 60 + minutes


class ValueColumn(BaseModel):
    InternalName: str


class ResponseData(BaseModel):
    Description: str
    ValueColumns: List[ValueColumn]
    ValueRows: List[ValueRow]


class ImbalanceResponse(BaseModel):
    ResponseData: ResponseData


class Actual(SQLModel, table=True):
    __tablename__ = "actual"
    __table_args__ = {"schema": "public"}

    time: datetime
    data: float
    metric_id: UUID
    scope_id: UUID


def new_actual_from_valuerow(row: ValueRow) -> Actual:
    return Actual(
        time=datetime.strptime(f"{row.DF} {row.TF}", "%d.%m.%Y %H:%M"),
        data=row.V[0].V,
        metric_id=uuid4(),
        scope_id=uuid4(),
    )


def fetch_imbalance_data(start_date: str, end_date: str) -> ImbalanceResponse:
    """
    Fetch imbalance data from APG transparency API

    Args:
        start_date: Start date in format "YYYY-MM-DDTHHMMSS"
        end_date: End date in format "YYYY-MM-DDTHHMMSS"

    Returns:
        ImbalanceResponse object with parsed data
    """
    url = f"https://transparency.apg.at/api/v1/DRZ/Data/German/PT1M/{start_date}/{end_date}"

    headers = {
        "Accept": "application/json, text/plain, */*",
        "Accept-Language": "en-US,en;q=0.9,el;q=0.8",
        "Cache-Control": "no-cache",
        "Pragma": "no-cache",
        "Referer": "https://transparency.apg.at/deltaregelzone/chart?p_drzMode=Operational&resolution=PT1M&language=German&embed=true",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
    }

    params = {"p_drzMode": "CurrentBalancingState", "resolution": "PT1M"}

    response = requests.get(url, headers=headers, params=params)
    response.raise_for_status()

    # print(response.json())
    # exit(0)
    return ImbalanceResponse.model_validate(response.json())


# Example usage
if __name__ == "__main__":
    # Fetch data for May 5, 2025
    start_date = "2025-05-05T000000"
    end_date = "2025-05-06T000000"

    try:
        data = fetch_imbalance_data(start_date, end_date)

        # Print some example data points with their minutes of the day
        print(f"Total rows: {len(data.ResponseData.ValueRows)}")

        for i, row in enumerate(data.ResponseData.ValueRows[:5]):
            value = row.V[0].V if row.V else None
            print(f"Time: {row.TF} (minutes: {row.minutes_from}), Value: {value}")

    except Exception as e:
        print(f"Error fetching data: {e}")
