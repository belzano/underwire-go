import os
import subprocess
import time
import uuid

import pytest
import requests

COMPOSE_DIR = os.path.normpath(
    os.path.join(os.path.dirname(__file__), "..", "..", "..", "jij-compose")
)
COMPOSE_PROJECT = "jij-integration"
COMPOSE_FILES = ["-f", "docker-compose.yml", "-f", "docker-compose.integration.yml"]
BASE_URL = "http://localhost:5021"


def _compose(*args):
    subprocess.run(
        ["docker", "compose", "-p", COMPOSE_PROJECT, *COMPOSE_FILES, *args],
        cwd=COMPOSE_DIR,
        check=True,
    )


def _wait_until_ready(timeout=90):
    deadline = time.time() + timeout
    last_error = None
    while time.time() < deadline:
        try:
            response = requests.get(f"{BASE_URL}/ping", timeout=2)
            if response.status_code == 200:
                return
        except requests.RequestException as error:
            last_error = error
        time.sleep(1)
    raise RuntimeError(f"jij-service did not become ready in time: {last_error}")


@pytest.fixture(scope="session", autouse=True)
def integration_stack():
    _compose("up", "-d", "--build")
    try:
        _wait_until_ready()
        yield
    finally:
        _compose("down", "-v")


@pytest.fixture(scope="session")
def base_url(integration_stack):
    return BASE_URL


@pytest.fixture
def unique_id():
    return uuid.uuid4().hex[:8]
