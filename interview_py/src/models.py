from pydantic import BaseModel, Field
from sqlmodel import SQLModel, Field as SQLField
from typing import List, Optional, Dict, Any
from uuid import UUID, uuid4
from datetime import datetime


# API response models
class ValueItem(BaseModel):
    V: Optional[float] = None
    E: bool
    M: bool


class ValueRow(BaseModel):
    DF: str  # Date From
    TF: str  # Time From
    DT: str  # Date To
    TT: str  # Time To
    V: List[ValueItem]

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

    @property
    def timestamp(self) -> datetime:
        """Convert DF and TF to datetime"""
        return datetime.strptime(f"{self.DF} {self.TF}", "%d.%m.%Y %H:%M")


class ValueColumn(BaseModel):
    InternalName: str


class ResponseData(BaseModel):
    Description: str
    ValueColumns: List[ValueColumn]
    ValueRows: List[ValueRow]


class ImbalanceResponse(BaseModel):
    ResponseData: ResponseData


# Database models
class Metric(SQLModel, table=True):
    __tablename__ = "metric"
    __table_args__ = {"schema": "public"}

    metric_id: UUID = SQLField(default_factory=uuid4, primary_key=True)
    name: str = SQLField(index=True)
    description: Optional[str] = None


class Scope(SQLModel, table=True):
    __tablename__ = "scope"
    __table_args__ = {"schema": "public"}

    scope_id: UUID = SQLField(default_factory=uuid4, primary_key=True)
    name: str = SQLField(index=True)
    description: Optional[str] = None


class Actual(SQLModel, table=True):
    __tablename__ = "actual"
    __table_args__ = {"schema": "public"}

    time: datetime = SQLField(primary_key=True)
    data: float
    metric_id: UUID = SQLField(foreign_key="public.metric.metric_id", primary_key=True)
    scope_id: UUID = SQLField(foreign_key="public.scope.scope_id", primary_key=True)


# Restate service request/response models
class FetchDataRequest(BaseModel):
    start_date: str
    end_date: str


class ProcessDataRequest(BaseModel):
    metric_name: str = "apg_imbalance"
    scope_name: str = "austria"
    data_point: Dict[str, Any]


class ValidationResult(BaseModel):
    is_valid: bool
    reason: Optional[str] = None
    processed_data: Optional[Dict[str, Any]] = None
