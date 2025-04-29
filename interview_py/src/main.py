import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

logger = logging.getLogger(__name__)

def add_numbers(a: float, b: float) -> float:
    result = a + b
    return result

if __name__ == "__main__":
    logger.info("Starting the add_numbers application")
    num1 = 5.0
    num2 = 3.0
    result = add_numbers(num1, num2)
    logger.info("Addition completed successfully")
