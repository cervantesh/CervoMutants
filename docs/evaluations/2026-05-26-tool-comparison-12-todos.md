# Tool Comparison 12 TODOs

Tracking issue: https://gitea.cervbox.synology.me/CervoSoft/cervo-mutant/issues/13

This file intentionally records the paused 12-repository comparison against the
three external tools: Gremlins, gomu, and go-mutesting.

## Scope

Repositories:

```text
cobra, pflag, moby, hugo, prometheus, terraform, grpc-go, echo, logrus, validator, decimal, gjson
```

Tools:

```text
cervomut
gremlins
gomu
go-mutesting
```

Timeout:

```text
600 seconds per tool/repository
```

Output root:

```text
C:\Users\c___h\AppData\Local\Temp\cervomut-tool-comparison-12
```

## Current Partial Results

Completed before pausing:

| Repo | Tool | Exit | Seconds | Notes |
| --- | --- | ---: | ---: | --- |
| `cobra` | `cervomut` | 0 | 8.63 | Report parsed, but metrics need review because current parser saw killed/survived as zero for this cached run. |
| `cobra` | `gremlins` | 0 | 41.02 | Parsed. |
| `cobra` | `gomu` | 0 | 76.75 | Parsed. |
| `cobra` | `go-mutesting` | 0 | 104.08 | Parsed. |
| `pflag` | `cervomut` | 0 | 2.43 | Report parsed, but metrics need review. |
| `pflag` | `gremlins` | 2 | 1.26 | Failed; inspect log. |
| `pflag` | `gomu` | 1 | 0.09 | Failed; inspect log. |
| `pflag` | `go-mutesting` | 124 | 600.21 | Timed out. |
| `moby` | `cervomut` | 0 | 338.12 | Parsed. |
| `moby` | `gremlins` | 0 | 53.95 | Completed but parser did not extract metrics; inspect report/log shape. |
| `moby` | `gomu` | 124 | 600.21 | Timed out. |
| `moby` | `go-mutesting` | 124 | 600.36 | Timed out. |
| `hugo` | `cervomut` | 0 | 355.46 | Parsed. |

Partial summary file:

```text
C:\Users\c___h\AppData\Local\Temp\cervomut-tool-comparison-12\summary.json
```

## TODO

1. Fix `scripts/compare-tools-pool.ps1` CervoMutant parser so cached/fast reports
   do not show `killed=0` and `survived=0` when the report contains real results.
2. Fix Gremlins parser for repos where Gremlins exits 0 but metrics are null.
3. Re-run with `-Resume` after parser fixes:

   ```powershell
   .\scripts\compare-tools-pool.ps1 -TimeoutSeconds 600 -Workers 2 -Resume
   ```

4. If the full 12-repo comparison remains too long, split by repo groups:

   ```powershell
   .\scripts\compare-tools-pool.ps1 -Names cobra,pflag,moby,hugo -TimeoutSeconds 600 -Workers 2 -Resume
   .\scripts\compare-tools-pool.ps1 -Names prometheus,terraform,grpc-go,echo -TimeoutSeconds 600 -Workers 2 -Resume
   .\scripts\compare-tools-pool.ps1 -Names logrus,validator,decimal,gjson -TimeoutSeconds 600 -Workers 2 -Resume
   ```

5. Convert `summary.json` into a Markdown comparison table with:
   - completion rate;
   - timeouts;
   - killed/survived/not-covered/errors/not-viable;
   - score/test efficacy;
   - denominator caveats per tool.

6. Update issue #13 with the completed comparison.

## 2026-05-27 Resource-Bounded Retry Findings

Issue #13 later expanded the comparison from the paused 12-repository mixed run
into separated tool phases over 20 repositories, followed by one-at-a-time
retries for `hugo` and `grpc-go`.

The retry target was to recover metrics for the two non-CervoMutant reference
tools that failed to produce usable metrics for `hugo` and `grpc-go`:

```text
gomu
go-mutesting
```

The one-at-a-time retry used process-tree resource controls instead of passive
global memory waiting:

```text
MaxProcessTreeMemoryMB: 6144
GOMEMLIMIT: 3GiB
GOMAXPROCS: 1
GOFLAGS: -p=1
workers: 1
timeout: 1800s per repo/tool
```

Results:

| Repo | Tool | Exit | Seconds | Outcome |
| --- | --- | ---: | ---: | --- |
| `hugo` | `gomu` | 124 | 1802.67 | Timed out without usable metrics. |
| `grpc-go` | `gomu` | 126 | 75.81 | Killed by process-tree watchdog at ~6953MB working set / ~6971MB private. |
| `hugo` | `go-mutesting` | 124 | 1802.12 | Timed out without usable metrics. |
| `grpc-go` | `go-mutesting` | 126 | 30.51 | Killed by process-tree watchdog at ~8464MB working set / ~8801MB private. |

Finding:

- Both reference tools can fail to degrade gracefully under resource limits on
  larger Go targets. Even with one repository, one package target, one worker,
  `GOMAXPROCS=1`, `GOFLAGS=-p=1`, and `GOMEMLIMIT=3GiB`, the `grpc-go` runs
  exceeded the 6GB process-tree limit through tool and `go test` child-process
  activity.
- `GOMEMLIMIT` is useful but insufficient as a hard memory boundary. It does not
  bound the whole process tree, compiler/linker subprocesses, or all native
  allocations. A CI-safe mutation tool needs an explicit process-tree watchdog.
- Timeout-only failure is not enough. A useful tool should write partial,
  comparable metrics before budget exhaustion or watchdog termination. These
  retries produced controlled exits but no additional metrics, which limits
  their value for large-project CI comparison.

CervoMutant design implications:

- Keep process-tree memory accounting in the comparison runner and move the same
  concept into CervoMutant's own execution model where possible.
- Prefer incremental result checkpoints after each mutant, not only at the end
  of a package/tool run.
- Treat timeout, memory-watchdog, and skipped-for-resources as first-class
  statuses in JSON reports.
- Budget-aware scheduling should stop before resource exhaustion and still
  report `attempted`, `killed`, `survived`, `pending`, and `stopped_reason`.
- Large-project CI profiles should support smaller package slices, maximum
  mutants per package, and early partial summaries so a failed run is still
  diagnostically useful.
