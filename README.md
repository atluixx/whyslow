# whyslow

`whyslow` is a lightweight Linux performance diagnostic tool. It is not an
`htop` replacement: it samples system telemetry over time and answers the more
useful question, **“what evidence suggests my machine is slow?”**

It uses only local Linux `/proc` and `/sys` data. It sends no telemetry and
keeps diagnostic history only in memory while it runs.

## Run

```sh
go run ./cmd/whyslow
```

Useful options:

```sh
go run ./cmd/whyslow --interval 500ms --top 20
go run ./cmd/whyslow --color=false
go run ./cmd/whyslow --help
```

The dashboard refreshes every second by default. Press `Ctrl-C` to exit.

## Command and service lifecycle

```text
whyslow                 Run the interactive dashboard (same as `whyslow run`)
whyslow run             Explicitly run the dashboard in the foreground
whyslow status          Show the installed systemd service status
whyslow start|stop|restart
whyslow install         Install and enable the systemd service (root required)
whyslow uninstall       Stop and remove files created by install (root required)
whyslow version         Print the build version
whyslow help            Show command help
```

`install` copies the current executable to `/usr/local/bin/whyslow`, installs
only the managed `/etc/systemd/system/whyslow.service` unit, reloads systemd,
and enables the service. It does **not** start it; use `whyslow start` when
ready. Installation never invokes `sudo` itself and refuses to overwrite an
unmanaged binary or unit.

The service runs `whyslow service` as a normal foreground process. systemd—not
whyslow—handles supervision, restarts, journald logging, and background
execution. The service uses `DynamicUser` and systemd hardening settings; it
does not need a graphical session or elevated privileges while monitoring.

Build a release with an injected version:

```sh
make build VERSION=0.1.0
# or:
go build -ldflags "-X main.Version=0.1.0" -o whyslow ./cmd/whyslow
```

## Architecture

The main loop follows a small, explicit pipeline:

```text
collectors.CollectSnapshot → analyzers.Analyze(history) → ui.Render
```

- Collectors read one coherent `Snapshot` from `/proc` and `/sys`.
- Analyzers calculate counter deltas, rates, sustained conditions, likely
  contributors, and bounded in-memory events.
- The UI presents those results without collecting or interpreting telemetry.

## Metrics

Each snapshot includes:

- global CPU use and I/O wait from `/proc/stat`;
- memory availability, swap total/free, and derived swap used/percentage from
  `/proc/meminfo`;
- swap page-in/page-out activity from `/proc/vmstat`;
- 1-, 5-, and 15-minute load averages from `/proc/loadavg`;
- CPU-ranked processes, RSS, and thread counts from `/proc/<pid>`;
- aggregate whole-device read/write counters and I/O time from
  `/proc/diskstats` (partition entries and loop/ram devices are ignored);
- optional temperatures from readable thermal-zone and hwmon sensors.

Disk sectors are converted with Linux’s `/proc/diskstats` 512-byte sector unit.
Temperature sensors are optional: unavailable or unreadable sensors are simply
omitted.

## Diagnosis and events

Rather than warning on one spike, `whyslow` waits for three consecutive samples
before flagging sustained CPU, memory, load, swap, disk, or temperature
conditions. Current checks include sustained CPU pressure, elevated memory with
active swapping, load at or above the logical CPU count, high disk throughput,
and high reported temperatures.

Diagnosis messages include the most CPU-intensive processes as *contributors*
when that evidence is relevant. They do not assert that a process or disk is
the definitive cause. Events are emitted only when a condition starts (or a
new process becomes a significant top CPU contributor), preventing the event
list from repeating every refresh.

Example:

```text
SYSTEM HEALTH
  CPU        94.2% [███████████·]   MEMORY     82.1% [██████████··]
  I/O WAIT    2.1% [············]   SWAP        4.0% [············]

DISK
  Read 124.3 MB/s  Write 38.2 MB/s  I/O busy 61.0%

DIAGNOSIS
  CRITICAL CPU pressure
    CPU utilization has remained above 90% for several samples.
    contributor firefox                    82.31%

RECENT EVENTS
  21:43:17  CPU pressure
```

## Limitations

`whyslow` can identify evidence consistent with resource pressure, but cannot
prove a single root cause. In particular, high load is not CPU percentage,
high disk activity is not proof of a disk bottleneck, configured swap is not
itself a problem, and a high temperature does not prove throttling. It is
Linux-only and currently reports aggregate disk I/O rather than per-device or
per-process I/O attribution.
