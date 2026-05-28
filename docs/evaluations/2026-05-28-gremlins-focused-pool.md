# Gremlins-Focused Comparison Pool

Tracking issue: https://gitea.cervbox.synology.me/CervoSoft/cervo-mutant/issues/13

Date: 2026-05-28

This campaign narrows external comparison to:

```text
CervoMutant vs Gremlins
```

`gomu` and `go-mutesting` remain historical findings only. They repeatedly
failed to produce stable, comparable metrics under Windows and WSL2 resource
limits, so they are no longer useful as primary benchmarks for CervoMutant.

## Current Anchors

The useful prior cross-tool anchors are:

| Repository | Target | Size bucket | Why it stays |
| --- | --- | --- | --- |
| `cobra` | `./doc` | small | Existing apples-to-apples CLI benchmark; CervoMutant is faster, Gremlins has stronger score. |
| `grpc-go` | `./metadata` | medium | CervoMutant completed cleanly in WSL2 and beat Gremlins on score. |
| `hugo` | `./helpers` | large | Stress target where Gremlins currently has better effective coverage/stability. |

`pflag` is kept in the pool but is not a primary anchor because previous runs
often exited without useful metrics for one or both tools.

## Size Buckets

The buckets are pragmatic, not prestige-based:

- `small`: library-sized targets expected to run locally without setup and give
  useful metrics inside a short CI budget.
- `medium`: framework or infrastructure libraries with broader dependency graphs
  or more package/test variance, but still practical as bounded CI targets.
- `large`: apps, platforms, or expensive framework targets that test scheduler,
  timeout, memory, coverage, and partial-report behavior.

## Small Pool: First Run

These are the first 10 to run. They intentionally cover CLI, logging, parsing,
numeric, validation, and utility domains while staying small enough for a local
Gremlins-vs-CervoMutant comparison.

| # | Repo | Target | Lane | Domain | Reason |
| ---: | --- | --- | --- | --- | --- |
| 1 | `cobra` | `./doc` | tuning | cli | Existing anchor; validates continuity with prior results. |
| 2 | `pflag` | `./...` | validation | cli | Cobra-adjacent but smaller; useful to detect CLI overfitting. |
| 3 | `logrus` | `./...` | tuning | logging | Mature small logging library with branch/string behavior. |
| 4 | `uuid` | `./...` | tuning | utility | Small correctness-focused package. |
| 5 | `decimal` | `./...` | tuning | numeric | Numeric/operator-sensitive code; previous timeout makes it a useful stress-small case. |
| 6 | `gjson` | `./...` | validation | parser | Fast parser/string-heavy library with useful survivor signal. |
| 7 | `sjson` | `./...` | validation | parser | JSON update logic, separate from `gjson` read path. |
| 8 | `jsonparser` | `./...` | validation | parser | Independent parser implementation. |
| 9 | `burntsushi-toml` | `./...` | holdout | parser | Independent TOML parser holdout. |
| 10 | `urfave-cli` | `./...` | validation | cli | CLI library independent from Cobra/pflag. |

Run command:

```powershell
.\scripts\compare-tools-pool.ps1 `
  -Names cobra,pflag,logrus,uuid,decimal,gjson,sjson,jsonparser,burntsushi-toml,urfave-cli `
  -Tools cervomut,gremlins `
  -OutputRoot "$env:TEMP\cervomut-gremlins-small-10" `
  -TimeoutSeconds 600 `
  -Workers 2 `
  -Resume `
  -MaxProcessTreeMemoryMB 6144 `
  -GoMemoryLimit 3GiB `
  -GoMaxProcs 2 `
  -GoFlags "-p=2"
```

Artifacts:

```text
C:\Users\c___h\AppData\Local\Temp\cervomut-gremlins-small-10
```

## Medium Pool

These should run after small-pool parser/stability issues are resolved.

| # | Repo | Target | Lane | Domain | Reason |
| ---: | --- | --- | --- | --- | --- |
| 1 | `grpc-go` | `./metadata` | validation | networking | Existing anchor; strong CervoMutant result in WSL2. |
| 2 | `echo` | `./...` | validation | web | Web framework with prior CervoMutant signal. |
| 3 | `chi` | `./...` | tuning | web | Smaller router, good fast web target. |
| 4 | `gin` | `./...` | validation | web | Popular web framework; setup needs review. |
| 5 | `fiber` | `./...` | validation | web | Alternative web framework with different dependency shape. |
| 6 | `validator` | `./...` | tuning | validation | Tag/rule validation; prior not-covered result needs Gremlins comparison. |
| 7 | `testify` | `./assert` | tuning | testing | Assertion logic; survivors should be human-actionable. |
| 8 | `zap` | `./zapcore` | validation | logging | Performance-oriented logging core. |
| 9 | `go-toml` | `./...` | validation | parser | Config parser with many edge cases. |
| 10 | `go-redis` | `./internal/...` | validation | database | Protocol/client internals without live Redis target. |

## Large Pool

These are for scheduler, budget, memory, and partial-report validation. They
should be run one at a time or in WSL2/cgroup-protected mode when possible.

| # | Repo | Target | Lane | Domain | Reason |
| ---: | --- | --- | --- | --- | --- |
| 1 | `hugo` | `./helpers` | holdout | static-site | Existing large anchor; Gremlins currently wins effective coverage. |
| 2 | `moby` | `./pkg/...` | holdout | containers | Large production codebase; previous tools struggled. |
| 3 | `prometheus` | `./model/...` | holdout | observability | Metrics/parser-heavy production Go. |
| 4 | `terraform` | `./internal/addrs/...` | holdout | iac | Config/state semantics with bounded package target. |
| 5 | `gitea` | `./modules/...` | validation | devtools | Large Go app relevant to Cervo infrastructure. |
| 6 | `rclone` | `./fs/...` | holdout | storage | IO-heavy behavior and broad package graph. |
| 7 | `etcd` | `./client/v3/...` | holdout | distributed-systems | Distributed systems client code with concurrency. |
| 8 | `kubernetes` | `./pkg/scheduler/cache` | special | orchestration | Empirical-study-grade large repo; scoped target only. |
| 9 | `go` | `./src/cmd/compile/...` | special | language | Special layout; useful only after runner setup adapts. |
| 10 | `go-ethereum` | `./common/...` | special | blockchain | Large correctness-heavy codebase; scoped package target. |

## Evaluation Rules

- Compare only CervoMutant and Gremlins.
- Report denominators explicitly: generated, covered, executed, killed,
  survived, not-covered, timed-out, compile errors.
- Do not compare scores without noting not-covered and timed-out counts.
- Track wall time and seconds per executed mutant.
- Treat controlled watchdog exits as a useful CervoMutant/runner property, but
  not as a quality win unless partial metrics are preserved.
- Do not tune from holdout results until a candidate change is frozen.

## Small-Pool Result Table

First run completed on Windows/OneDrive.

Command:

```powershell
.\scripts\compare-tools-pool.ps1 `
  -Names cobra,pflag,logrus,uuid,decimal,gjson,sjson,jsonparser,burntsushi-toml,urfave-cli `
  -Tools cervomut,gremlins `
  -OutputRoot "$env:TEMP\cervomut-gremlins-small-10" `
  -TimeoutSeconds 600 `
  -Workers 2 `
  -Resume `
  -MaxProcessTreeMemoryMB 6144 `
  -GoMemoryLimit 3GiB `
  -GoMaxProcs 2 `
  -GoFlags "-p=2"
```

Artifacts:

```text
C:\Users\c___h\AppData\Local\Temp\cervomut-gremlins-small-10\summary.json
```

| Repo | Tool | Exit | Seconds | Total | Killed | Survived | Not covered | Timed out | Errors | Score | Notes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `cobra` | `cervomut` | 0 | 28.49 | 69 | 51 | 18 | 0 | 0 | 0 | 73.91 | Complete metrics. |
| `cobra` | `gremlins` | 0 | 49.59 | 85 | 56 | 29 | 5 | 2 | 0 | 65.88 | Complete metrics. |
| `pflag` | `cervomut` | 0 | 223.33 | 214 | 185 | 27 | 0 | 2 | 0 | 86.45 | Complete metrics. |
| `pflag` | `gremlins` | 0 | 8.20 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `logrus` | `cervomut` | 0 | 206.62 | 103 | 73 | 29 | 0 | 1 | 0 | 70.87 | Complete metrics. |
| `logrus` | `gremlins` | 0 | 9.64 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `uuid` | `cervomut` | 0 | 182.87 | 89 | 70 | 16 | 0 | 3 | 0 | 78.65 | Complete metrics. |
| `uuid` | `gremlins` | 0 | 7.97 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `decimal` | `cervomut` | 124 | 611.43 |  |  |  |  |  |  |  | Timeout before final metrics. |
| `decimal` | `gremlins` | 0 | 6.99 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `gjson` | `cervomut` | 124 | 608.83 |  |  |  |  |  |  |  | Timeout before final metrics. |
| `gjson` | `gremlins` | 0 | 6.36 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `sjson` | `cervomut` | 126 | 47.16 |  |  |  |  |  |  |  | Process-tree watchdog kill. |
| `sjson` | `gremlins` | 0 | 6.64 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `jsonparser` | `cervomut` | 0 | 512.67 | 874 | 827 | 40 | 0 | 7 | 0 | 94.62 | Complete metrics, but near 10-minute budget. |
| `jsonparser` | `gremlins` | 0 | 6.82 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `burntsushi-toml` | `cervomut` | 124 | 605.28 |  |  |  |  |  |  |  | Timeout before final metrics. |
| `burntsushi-toml` | `gremlins` | 0 | 7.80 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |
| `urfave-cli` | `cervomut` | 124 | 606.75 |  |  |  |  |  |  |  | Timeout before final metrics. |
| `urfave-cli` | `gremlins` | 0 | 7.52 |  |  |  |  |  |  |  | Panic after coverage; no JSON report. |

## Small-Pool Findings

1. `cobra` is the only small-pool repo where both tools produced complete
   comparable metrics in this Windows run.
2. Gremlins panicked after coverage collection on 9/10 small-pool targets. The
   process returned exit 0 in the harness, so the comparison runner must classify
   `panic:` in the log as an execution failure even when the process exit code is
   misleading.
3. CervoMutant produced complete metrics on 5/10 targets: `cobra`, `pflag`,
   `logrus`, `uuid`, and `jsonparser`.
4. CervoMutant timed out on 4/10 targets: `decimal`, `gjson`,
   `burntsushi-toml`, and `urfave-cli`.
5. CervoMutant watchdog-killed `sjson` quickly. This is safer than exhausting the
   host, but the result still needs partial metrics to be useful.
6. Full mutation is too expensive for this "small" pool. `jsonparser` generated
   874 mutants and took 512.67s. The next comparison should be bounded by
   deterministic sample or budget on both tools where possible.
7. CervoMutant has a reporting advantage when it completes: the JSON captures
   denominator counts and timeout counts consistently. The weakness is that
   timeout/watchdog exits currently lose partial per-mutant metrics.

## Immediate Improvements From This Run

- Update the comparison harness to detect Gremlins panics in logs and mark them
  as `panic` or `tool_error`, not successful runs with null metrics.
- Add CervoMutant checkpoint/partial-summary preservation for timeout and
  watchdog exits.
- Add a bounded comparison mode for CervoMutant vs Gremlins:
  - deterministic sample when the tool supports it;
  - otherwise fixed budget and explicit "not comparable denominator" flag.
- Re-run the small pool in WSL2 to distinguish Windows-specific Gremlins panic
  behavior from general Gremlins behavior.
