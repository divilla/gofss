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
cpu: AMD Ryzen 7 7700X 8-Core Processor             

BenchmarkSessionCreate-16     77972    14400 ns/op    5481 B/op    115 allocs/op
BenchmarkSessionRead-16      908493     1244 ns/op    1344 B/op      8 allocs/op
BenchmarkSessionUpdate-16    101702    10107 ns/op     502 B/op      6 allocs/op
BenchmarkSessionDelete-16     86104    12010 ns/op     414 B/op      5 allocs/op
```
