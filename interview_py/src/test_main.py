import pytest
from main import add_numbers


class TestAddNumbers:
    def test_positive_numbers(self):
        assert add_numbers(2, 3) == 5

    def test_negative_numbers(self):
        assert add_numbers(-2, -3) == -5

    def test_mixed_numbers(self):
        assert add_numbers(-2, 5) == 3

    def test_zero(self):
        assert add_numbers(7, 0) == 7
        assert add_numbers(0, 7) == 7

    def test_float_numbers(self):
        assert add_numbers(1.5, 2.5) == 4.0

    def test_large_numbers(self):
        assert add_numbers(1000000, 2000000) == 3000000


if __name__ == "__main__":
    pytest.main(["-v"])
