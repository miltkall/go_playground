from sqlmodel import SQLModel, create_engine, Session, select
from models import Metric, Scope, Actual
from datetime import datetime
from uuid import UUID
from typing import Optional, List
import os

# Database connection settings
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASSWORD = os.getenv("DB_PASSWORD", "postgres")
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = os.getenv("DB_PORT", "5433")
DB_NAME = os.getenv("DB_NAME", "postgres")

DATABASE_URL = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

# Create engine
engine = create_engine(DATABASE_URL)


def init_db():
    """Create tables if they don't exist"""
    SQLModel.metadata.create_all(engine)


def get_or_create_metric(name: str, description: Optional[str] = None) -> Metric:
    """Get a metric by name or create it if it doesn't exist"""
    with Session(engine) as session:
        statement = select(Metric).where(Metric.name == name)
        metric = session.exec(statement).first()

        if metric is None:
            metric = Metric(name=name, description=description)
            session.add(metric)
            session.commit()
            session.refresh(metric)

        return metric


def get_or_create_scope(name: str, description: Optional[str] = None) -> Scope:
    """Get a scope by name or create it if it doesn't exist"""
    with Session(engine) as session:
        statement = select(Scope).where(Scope.name == name)
        scope = session.exec(statement).first()

        if scope is None:
            scope = Scope(name=name, description=description)
            session.add(scope)
            session.commit()
            session.refresh(scope)

        return scope


def save_actual_data(
    time: datetime, data: float, metric_id: UUID, scope_id: UUID
) -> bool:
    """Save actual data to database, update if already exists"""
    try:
        with Session(engine) as session:
            # Check if record exists
            statement = select(Actual).where(
                Actual.time == time,
                Actual.metric_id == metric_id,
                Actual.scope_id == scope_id,
            )
            existing = session.exec(statement).first()

            if existing:
                # Update existing record
                existing.data = data
            else:
                # Create new record
                actual = Actual(
                    time=time, data=data, metric_id=metric_id, scope_id=scope_id
                )
                session.add(actual)

            session.commit()
            return True
    except Exception as e:
        print(f"Error saving data: {e}")
        return False


def get_recent_data(metric_id: UUID, scope_id: UUID, limit: int = 5) -> List[Actual]:
    """Get recent data points for a metric and scope"""
    with Session(engine) as session:
        statement = (
            select(Actual)
            .where(Actual.metric_id == metric_id, Actual.scope_id == scope_id)
            .order_by(Actual.time.desc())
            .limit(limit)
        )
        return session.exec(statement).all()
