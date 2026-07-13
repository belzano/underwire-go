# underwire-go
Game simple backend services and utilities

## Integration tests

`jij-service` has Python integration tests that spin up its own isolated docker-compose stack (`jij-integration` project, ports 5021/27018/6380) and tear it down afterwards. Run them with:

```
./run-integration-tests.sh
```

This creates/reuses a venv in `jij-service/tests/integration/.venv`, installs `requirements.txt`, and runs `pytest`. Requires Docker to be running.
