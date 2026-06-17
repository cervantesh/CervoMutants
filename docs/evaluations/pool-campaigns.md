# Pool Campaigns

`cervomut pool campaign` turns the existing pool runners into one resumable
workflow.

Use it when a study or release lane needs more than one step, for example:

- smoke a repository pool to verify checkout and dry-run viability
- compare a smaller subset against Gremlins or other tools
- run the pinned benchmark corpus as the final regression gate

## Command

```powershell
cervomut pool campaign `
  --file docs/evaluations/pool-campaign-example.json `
  --resume `
  --cervomutants $env:TEMP/cervomut-pool.exe
```

The command writes `campaign-summary.json` at the campaign output root. That
summary records:

- each job name and kind
- status: `ok`, `failed`, or `skipped`
- elapsed time
- nested summary path
- generated artifact paths
- whether the job was resumed from an earlier campaign summary

If one job fails, the campaign keeps going and records the failure in the
summary. The CLI exits non-zero afterward so CI can still fail correctly.

## Manifest Shape

The manifest is JSON:

```json
{
  "schema_version": "1",
  "tracking_issue": "#89",
  "description": "example mixed pool campaign",
  "jobs": [
    {
      "name": "pool-smoke-shared",
      "kind": "smoke",
      "manifest_path": "go-repo-pool-40.json"
    }
  ]
}
```

Supported job kinds:

- `smoke`
- `compare`
- `benchmark`

Common behavior:

- `manifest_path`, `corpus_path`, `work_root`, and `output_root` are resolved
  relative to the campaign file when they are relative paths.
- `enabled: false` skips a job without deleting it from the manifest.
- `resume: true` on `compare` and `benchmark` also enables the nested runner's
  own resume support.

## Root Strategy

By default each job gets its own work/output directory under the campaign
defaults. That avoids collisions across separate jobs.

When you want clone reuse across steps, set the same explicit `work_root` on
those jobs. This is the usual pattern for `smoke` followed by `compare`.

## Job Fields

`smoke` jobs support the same practical controls as `pool smoke`, including:

- `names`
- `limit`
- `run_mutation`
- `max_mutants`
- `workers`
- clone/test/dry-run/mutation timeout fields

`compare` jobs support the important pool-compare controls, including:

- `names`
- `tools`
- `workers`
- `compare_target_mode`
- `gremlins_target_mode`
- timeout and memory guard fields
- `go_memory_limit`
- `go_max_procs`
- `go_flags`

`benchmark` jobs support:

- `names`
- `limit`
- `resume`

Benchmark execution details still come from the pinned benchmark corpus.

## Example

See [pool-campaign-example.json](pool-campaign-example.json) for a complete
mixed campaign manifest.
