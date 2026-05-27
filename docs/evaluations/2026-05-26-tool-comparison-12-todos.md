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

