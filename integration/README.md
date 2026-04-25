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

| Command | Needs Cortex? | Tested? |
|---------|:---:|:---:|
| **rules load/list/print/get** | Yes | Yes |
| **rules delete (group)** | Yes | Yes |
| **rules delete-namespace** | Yes | Yes |
| **rules lint** | No | Yes |
| **rules check** | No | Yes |
| **rules prepare** | No | Yes |
| **rules parse/validate** | No | Yes |
| **alertmanager load** | Yes | Yes |
| **alertmanager load (with templates)** | Yes | Yes |
| **alertmanager get** | Yes | Yes |
| **alertmanager delete** | Yes | Yes |
| **remote-read dump/stats** | Yes | Yes |
| **remote-read export** | Yes | Yes |
| **loadgen (write workload)** | Yes | Yes |
| **analyse** (grafana/ruler/dashboard/rule-file) | No | No |
| **acl** | No | No |
| **bucket-validation** | Needs object store | No |
| **config** (use-context) | No | No |
| **alert verify** | Needs Cortex + alerting | No |
| **push-gateway** | Needs push gateway | No |
