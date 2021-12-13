# For what
To delete promethues series in batch.
# Usage
1. Change these const variables
```go
const (
	prometheusURL = "http://localhost:9090"
	concurrentNum = 10 // must less than your series count
	seriesPrefix  = "envoy_"
)
```
2. Exec
```bash
go run main.go
```
