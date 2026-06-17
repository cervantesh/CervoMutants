# Benchmark Corpus Guide

Tracking issue: https://github.com/cervantesh/cervo-mutants/issues/80

## Purpose

`cervomut pool benchmark` turns performance claims into an auditable lane.

The goal is not to chase micro-benchmarks. The goal is to keep a small pinned
corpus that makes these regressions visible before release:

- baseline test runtime regressions;
- dry-run discovery regressions;
- mutation execution runtime regressions;
- peak memory regressions;
- mutation throughput regressions.

The starter corpus lives in
`docs/evaluations/benchmark-corpus.json`.

## What The Harness Runs

Each corpus entry records:

- a pinned repo URL and `ref`;
- a scoped mutation target;
- a size and resource class label;
- a policy, worker count, and mutant cap;
- explicit thresholds for runtime, peak memory, throughput, and minimum
  executed-mutant volume.

For each entry, the harness runs:

1. `go test <target>` for a baseline runtime check.
2. `cervomut run <target> --dry-run` with `--report summary,json`.
3. `cervomut run <target>` with the same bounded policy and report set.

The mutation run then reads `mutation-report.json`, or
`partial-mutation-report.json` if a final report is absent, and derives:

- `generated_mutants`
- `executed_mutants`
- `effective_mutants`
- `score_denominator`
- `max_peak_memory_mb`
- `mutants_per_second`

## Command

Example:

```powershell
cervomut pool benchmark `
  --corpus docs/evaluations/benchmark-corpus.json `
  --work-root $env:TEMP/cervomut-benchmark-corpus `
  --output-root $env:TEMP/cervomut-benchmark-results `
  --resume
```

Target a subset while calibrating thresholds:

```powershell
cervomut pool benchmark `
  --corpus docs/evaluations/benchmark-corpus.json `
  --names cobra-doc,logrus-root `
  --output-root $env:TEMP/cervomut-benchmark-results
```

The command exits non-zero when any entry:

- fails a configured threshold, or
- fails to complete clone, checkout, baseline, dry-run, or mutation execution.

## Output

The harness writes `summary.json` under the chosen output root.

Top-level summary fields include:

- corpus metadata
- generated timestamp
- totals for passed, failed, errored, and resumed entries
- per-entry metrics, thresholds, and check results

Each entry records:

- command exits for baseline, dry-run, and mutation runs
- measured seconds for each phase
- report path and whether a partial report was used
- all threshold checks with `pass` or `fail`
- any operational notes such as resume reuse or checkout failure

## Threshold Strategy

Keep thresholds explicit and conservative.

Recommended process:

1. Pin a repo `ref` before trusting the entry as a benchmark.
2. Run the corpus 3-5 times on the intended CI lane or workstation class.
3. Set thresholds slightly above the 95th percentile for runtime and memory.
4. Set throughput floors slightly below the 5th percentile.
5. Require a minimum executed-mutant count so an apparently fast run with weak
   denominator health does not pass as a benchmark success.

Do not mix laptop thresholds, local WSL thresholds, and GitHub Actions
thresholds in the same corpus without saying so. If the hardware class changes,
fork the corpus or update the thresholds deliberately.

## Scope Notes

This harness is intentionally narrower than the external comparison harness:

- `pool benchmark` tracks CervoMutants performance regressions.
- `pool compare` tracks apples-to-apples multi-tool comparisons.

Use `pool benchmark` to protect runtime and memory behavior before release.
Use `pool compare` when making claims against Gremlins, gomu, or
go-mutesting.
