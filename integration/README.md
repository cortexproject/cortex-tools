# Integration Tests

## Running (Legacy mode)

```bash
# Start Cortex
docker run -d --name cortex-test -p 9009:9009 \
  -v $(pwd)/integration/cortex-config.yaml:/etc/cortex/config.yaml \
  cortexproject/cortex:v1.21.0 \
  -config.file=/etc/cortex/config.yaml \
  -target=all,alertmanager

# Wait for ready
until curl -s http://localhost:9009/ready > /dev/null 2>&1; do sleep 2; done && echo "Ready"

# Run tests
go test -mod=vendor -tags=integration -v -count=1 ./integration/...

# Cleanup
docker stop cortex-test && docker rm cortex-test
```

## Running (UTF-8 mode)

```bash
# Start Cortex with UTF-8 validation
docker run -d --name cortex-utf8 -p 9009:9009 \
  -v $(pwd)/integration/cortex-config.yaml:/etc/cortex/config.yaml \
  cortexproject/cortex:v1.21.0 \
  -config.file=/etc/cortex/config.yaml \
  -target=all,alertmanager \
  -name-validation-scheme=utf8

# Wait for ready
until curl -s http://localhost:9009/ready > /dev/null 2>&1; do sleep 2; done && echo "Ready"

# Run tests
go test -mod=vendor -tags=integration_utf8 -v -count=1 ./integration/...

# Cleanup
docker stop cortex-utf8 && docker rm cortex-utf8
```

## Coverage

These integration tests require a running Cortex instance. Offline commands (lint, check, prepare, parse) are covered by unit tests in `pkg/rules/`.

### Legacy mode (`-tags=integration`)

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

### UTF-8 mode (`-tags=integration_utf8`)

| Test | Description |
|------|-------------|
| **TestUTF8RulesLoadListDelete** | Load/list/get/delete rules with dotted metric names against UTF-8 Cortex |
| **TestUTF8ParseBytes** | Verify `ParseBytes` accepts dotted names with `utf8` scheme and rejects with `legacy` |
