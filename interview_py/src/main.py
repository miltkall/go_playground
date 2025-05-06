import restate
import logging
from db_service import init_db
from restate_service import (
    apg_etl_service,
    data_processor,
    time_series_object,
    health_service,
)

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Create Restate app with all services
app = restate.app(
    services=[apg_etl_service, data_processor, time_series_object, health_service]
)

if __name__ == "__main__":
    # Initialize database
    logger.info("Initializing database...")
    init_db()
    logger.info(
        "Database initialized. Run server with: hypercorn --reload --bind 0.0.0.0:9080 main:app"
    )
