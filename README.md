# gofss
### Go Filesystem Session Store
A multithreaded, optimized, ultra-performant session store that is faster, safer, and more robust than any other persistent solution available in the market today. Additionally, it saves money as storage costs continue to decline, while using minimal compute resources.

### Google style sessions
The store was designed to enable year-long Google-style sessions with configurable high-level cryptography.

### Benchmark
```markdown
$ go test -bench=. -benchmem

goos: linux
goarch: amd64
pkg: github.com/divilla/gofss
cpu: 13th Gen Intel(R) Core(TM) i7-13650HX

BenchmarkSessionCreate-20        87756      11413 ns/op        5600 B/op      115 allocs/op
BenchmarkSessionRead-20        1426465        941.8 ns/op      1331 B/op        8 allocs/op
BenchmarkSessionUpdate-20       123658      14473 ns/op         500 B/op        6 allocs/op
BenchmarkSessionDelete-20        56817      17765 ns/op         393 B/op        4 allocs/op
```
