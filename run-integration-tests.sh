#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/jij-service/tests/integration"

if [ ! -d .venv ]; then
    python -m venv .venv
fi

if [ -f .venv/Scripts/activate ]; then
    source .venv/Scripts/activate
else
    source .venv/bin/activate
fi

pip install -q -r requirements.txt
pytest "$@"
