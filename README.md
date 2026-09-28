# whyslow

A lightweight Linux performance dashboard that explains likely bottlenecks from CPU, I/O wait, memory, swap, load, and process data.

## Run

```sh
go run ./cmd/whyslow
```

The terminal UI refreshes every second. Press `Ctrl-C` to exit cleanly.

```sh
# Refresh twice per second and show twenty processes.
go run ./cmd/whyslow --interval 500ms --top 20

# Disable ANSI colors (useful when capturing output).
go run ./cmd/whyslow --color=false
```

Use `go run ./cmd/whyslow --help` for the full command reference.

## What it reports

- CPU and I/O-wait utilization
- Memory and swap use
- 1-, 5-, and 15-minute load averages
- CPU-ranked processes with thread counts and resident memory
- Actionable diagnoses for CPU saturation, storage pressure, memory pressure, swap use, excess load, and CPU-heavy processes

`whyslow` reads Linux `/proc` files and is intended for Linux systems.
