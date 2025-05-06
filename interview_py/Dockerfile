# Use a Python image with uv pre-installed
FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim

RUN apt-get update && apt-get install -y \
    build-essential \
    ca-certificates \
    curl \
    git \
    iproute2 \
    yq \     
    zsh \     
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install the project's dependencies using pyproject.toml
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=/app/pyproject.toml \
    uv sync --frozen --no-install-project --no-dev

# Copy application code
ADD ./src /app

# EXPOSE 8501

# --reload for hot reloading in development
ENTRYPOINT ["uv", "run", "-m", "hypercorn", "--reload", "--bind", "0.0.0.0:9080", "main:app"]
