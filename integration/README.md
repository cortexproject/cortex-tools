# Integration Tests

## Running

```bash
# Start Cortex
docker run -d --name cortex-test -p 9009:9009 \
  -v $(pwd)/integration/cortex-config.yaml:/etc/cortex/config.yaml \
  cortexproject/cortex:v1.18.1 \
  -config.file=/etc/cortex/config.yaml \
  -target=all,alertmanager

# Wait for ready
until curl -s http://localhost:9009/ready > /dev/null 2>&1; do sleep 2; done && echo "Ready"

# Run tests
go test -mod=vendor -tags=integration -v -count=1 ./integration/...

# Cleanup
docker stop cortex-test && docker rm cortex-test
```

## Coverage

These integration tests require a running Cortex instance. Offline commands (lint, check, prepare, parse) are covered by unit tests in `pkg/rules/`.

| Command | Tested? |
|---------|:---:|
| **rules load/list/print/get** | Yes |
| **rules delete (group)** | Yes |
| **rules delete-namespace** | Yes |
| **alertmanager load** | Yes |
| **alertmanager load (with templates)** | Yes |
| **alertmanager get** | Yes |
| **alertmanager delete** | Yes |
| **remote-read dump/stats** | Yes |
| **remote-read export** | Yes |
| **loadgen (write workload)** | Yes |
| **analyse** (grafana/ruler/dashboard/rule-file) | No |
| **bucket-validation** | No |
| **config** (use-context) | No |
| **alert verify** | No |
| **push-gateway** | No |
