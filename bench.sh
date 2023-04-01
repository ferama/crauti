# !/bin/sh

go test -benchmem -run=^$ -bench ^BenchmarkRequest1$ github.com/ferama/crauti/pkg/gateway