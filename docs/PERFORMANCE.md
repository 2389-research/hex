# Performance Baseline and Optimization Guide

This document tracks performance benchmarks for Hex and provides guidelines for optimization.

**Last Updated:** 2025-11-28
**Platform:** Darwin ARM64 (Apple M4 Max)
**Go Version:** go1.23+

---

## Table of Contents

1. [Performance Targets](#performance-targets)
2. [Baseline Benchmarks](#baseline-benchmarks)
3. [Optimization Opportunities](#optimization-opportunities)
4. [Running Benchmarks](#running-benchmarks)
5. [Performance Guidelines](#performance-guidelines)

---

## Performance Targets

| Category | Target | Status |
|----------|--------|--------|
| Startup time (--help) | < 100ms | ✅ PASS |
| API request marshaling | < 1ms | ✅ PASS |
| HTTP round trip (local) | < 50ms | ✅ PASS |
| Database conversation create | < 100ms | ✅ PASS |
| Database message insert | < 100ms | ✅ PASS |
| Tool execution (Read) | < 50ms | ✅ PASS |
| Tool execution (Write) | < 100ms | ✅ PASS |

---

## Baseline Benchmarks

### 1. API Client Performance

**Test Environment:** Mock HTTP server, local loopback

| Benchmark | ops/sec | ns/op | allocs/op | bytes/op |
|-----------|---------|-------|-----------|----------|
| Client Creation | 102.6M | 23.55 ns | 2 | 96 B |
| Request Marshaling | 5.2M | 456.3 ns | 6 | 440 B |
| Response Unmarshaling | 1.5M | 1,554 ns | 16 | 624 B |
| HTTP Round Trip | 61.3K | 38,739 ns | 129 | 10,595 B |
| Large Payload (100KB) | 6.4K | 393,493 ns | 142 | 341,377 B |
| With Tool Definitions | 57.2K | 41,367 ns | 166 | 12,360 B |

**Key Findings:**
- ✅ Client creation is extremely fast (23ns)
- ✅ JSON marshaling is efficient (~450ns)
- ✅ HTTP round trip overhead is acceptable (~39µs)
- ⚠️ Large payloads show linear scaling (expected)
- ⚠️ Tool definitions add ~3µs overhead

**Optimization Opportunities:**
- None critical - performance is excellent
- Could pre-marshal tool definitions if they don't change

---

### 2. Database Performance

**Test Environment:** SQLite with WAL mode, temporary in-memory database

| Benchmark | ops/sec | ns/op | allocs/op | bytes/op |
|-----------|---------|-------|-----------|----------|
| Conversation Create | 51.2K | 58,468 ns | 22 | 865 B |
| Message Insert | 35.9K | 75,731 ns | 42 | 1,541 B |
| Message Get (by ID) | 382.9K | 6,215 ns | 45 | 1,360 B |
| Conversation List (20) | 76.5K | 31,527 ns | 423 | 14,472 B |
| Transaction Overhead | 46.4K | 61,239 ns | 27 | 1,154 B |
| Prepared Statement | 53.0K | 50,845 ns | 19 | 687 B |
| Large Message (1MB) | 1.1K | 1,975,081 ns | 41 | 1,050,141 B |
| Concurrent Reads | 483.8K | 4,812 ns | 45 | 1,379 B |

**Key Findings:**
- ✅ Message retrieval is very fast (6µs)
- ✅ Concurrent reads scale well (4.8µs)
- ✅ Prepared statements are ~20% faster than ad-hoc queries
- ⚠️ Message insertion is slower than ideal (75µs)
- ⚠️ Transaction overhead is significant (~61µs)
- ⚠️ Large message content shows expected scaling

**Optimization Opportunities:**
1. **Use prepared statements** - Already showing 20% improvement
2. **Batch message inserts** - Could reduce per-message overhead
3. **Add missing indexes** - Check query plans for conversation list
4. **Consider connection pooling** - For concurrent writes

---

### 3. Tool Execution Performance

**Test Environment:** Local filesystem, temporary directories

| Benchmark | ops/sec | ns/op | allocs/op | bytes/op |
|-----------|---------|-------|-----------|----------|
| Read (1KB file) | 93.4K | 12,880 ns | 15 | 3,368 B |
| Write (60B content) | 16.4K | 63,882 ns | 19 | 1,598 B |
| Edit (simple replace) | 20.0K | 44,600 ns | 15 | 1,587 B |
| Grep (100 files) | 141 | 7,092,272 ns | 220 | 70,509 B |
| Glob (100 files) | 1.3K | 786,344 ns | 1,407 | 209,254 B |
| Bash (echo) | 392 | 2,550,503 ns | 178 | 25,950 B |
| Registry Lookup | ~instant | <100 ns | - | - |
| Registry List All | 12.3M | 197.8 ns | 1 | 160 B |
| Executor Approval | ~instant | - | - | - |
| Concurrent Execution | 185.2M | 12.79 ns | 1 | 64 B |

**Key Findings:**
- ✅ Read tool is very fast (~13µs for 1KB)
- ✅ Registry operations are near-instant
- ✅ Concurrent execution scales well
- ⚠️ Write operations involve disk I/O (~64µs)
- ⚠️ Grep is expensive for large directory scans (~7ms)
- ⚠️ Bash tool has high overhead (~2.5ms for echo)

**Optimization Opportunities:**
1. **Grep tool** - Consider using ripgrep binary or optimize search patterns
2. **Bash tool** - Process pool for repeated commands
3. **Write tool** - Batch writes when possible
4. **Glob tool** - Cache results for repeated patterns

---

### 4. Startup Performance

**Test Environment:** Binary execution from command line

| Benchmark | Target | Actual | Status |
|-----------|--------|--------|--------|
| `hex --help` | < 100ms | ~50ms | ✅ PASS |
| `hex version` | < 100ms | ~45ms | ✅ PASS |

**Key Findings:**
- ✅ Startup is well under target
- ✅ No lazy loading needed

---

## Optimization Opportunities

### High Priority

None identified - all targets met.

### Medium Priority

1. **Database Message Insertion**
   - Current: 75µs per message
   - Target: < 50µs
   - Approach: Batch inserts, reduce transaction overhead

2. **Grep Tool Performance**
   - Current: 7ms for 100 files
   - Target: < 5ms
   - Approach: Use ripgrep binary, optimize pattern matching

### Low Priority

1. **Large Message Handling**
   - Current: Linear scaling with size (expected)
   - Approach: Streaming for very large messages
   - Note: Not critical for typical use

2. **Bash Tool Overhead**
   - Current: 2.5ms per command
   - Approach: Process pool, reuse shells
   - Note: Acceptable for current use cases

---

## Running Benchmarks

### Run All Benchmarks

```bash
# API client benchmarks
go test -bench=. -benchmem -benchtime=2s ./internal/core/

# Storage benchmarks
go test -bench=. -benchmem -benchtime=2s ./internal/storage/

# Tool benchmarks
go test -bench=. -benchmem -benchtime=2s ./internal/tools/

# Startup benchmarks
go test -bench=. -benchmem ./cmd/hex/
```

### Run Specific Benchmark

```bash
# Just HTTP round trip
go test -bench=BenchmarkHTTPRoundTrip -benchmem ./internal/core/

# Just message insertion
go test -bench=BenchmarkMessageInsert -benchmem ./internal/storage/
```

### Compare Before/After

```bash
# Save baseline
go test -bench=. -benchmem ./... > bench_before.txt

# Make changes...

# Compare
go test -bench=. -benchmem ./... > bench_after.txt
benchcmp bench_before.txt bench_after.txt  # or use benchstat
```

### Profiling

```bash
# CPU profile
go test -bench=BenchmarkHTTPRoundTrip -cpuprofile=cpu.prof ./internal/core/
go tool pprof cpu.prof

# Memory profile
go test -bench=BenchmarkMessageInsert -memprofile=mem.prof ./internal/storage/
go tool pprof -alloc_space mem.prof

# Block profile (for concurrency)
go test -bench=BenchmarkConcurrentReads -blockprofile=block.prof ./internal/storage/
go tool pprof block.prof
```

---

## Performance Guidelines

### When Writing New Code

1. **Avoid premature optimization**
   - Write clear code first
   - Profile before optimizing
   - Optimize hot paths only

2. **Database operations**
   - Use prepared statements for repeated queries
   - Batch operations when possible
   - Use transactions appropriately
   - Add indexes for common queries

3. **File I/O**
   - Buffer reads and writes
   - Avoid repeated filesystem operations
   - Cache file metadata when safe

4. **API calls**
   - Reuse HTTP clients
   - Don't pre-allocate large buffers unnecessarily
   - Stream large responses

5. **Tool execution**
   - Validate parameters early
   - Fail fast on errors
   - Clean up resources promptly

### Red Flags

🚨 **Alert if you see:**
- Any operation > 100ms (except external API calls)
- Memory allocations > 100KB for simple operations
- More than 1000 allocations for a single operation
- Startup time > 100ms

### Benchmarking Best Practices

1. **Always use `-benchmem`** to track allocations
2. **Use `-benchtime=2s`** or higher for stable results
3. **Run benchmarks multiple times** to account for variance
4. **Benchmark on the same hardware** for comparisons
5. **Use `b.ResetTimer()`** after setup code
6. **Use `b.StopTimer()/b.StartTimer()`** to exclude cleanup

---

## Historical Performance Tracking

| Date | Version | Key Metric | Value | Change |
|------|---------|------------|-------|--------|
| 2025-11-28 | v0.9.0 | API Round Trip | 38.7µs | Baseline |
| 2025-11-28 | v0.9.0 | Message Insert | 75.7µs | Baseline |
| 2025-11-28 | v0.9.0 | Startup Time | ~50ms | Baseline |

---

## Next Steps

1. ✅ Establish baseline (this document)
2. ⏭️ Identify optimization opportunities (see above)
3. ⏭️ Implement high-priority optimizations
4. ⏭️ Re-benchmark and track improvements
5. ⏭️ Add continuous performance testing to CI

---

## See Also

- [Architecture Documentation](ARCHITECTURE.md)
- [Developer Guide](../CONTRIBUTING.md)
- [Testing Guide](../README.md#testing)
