# Prometheus Medium-Service Prewarm Validation

Tracking issue: [#331](https://github.com/cervantesh/cervo-mutants/issues/331)

Date: 2026-06-20

This note records the hosted validation for the new optional
`prewarm_modules` path in `external-action-wave`.

The goal was narrower than lane retargeting or survivor-yield improvement:

1. prove that a manifest can opt specific jobs into module prewarm
2. make that choice visible in wave artifacts and summaries
3. verify the hosted workflow still completes end to end with mixed cold and
   prewarmed jobs

## Inputs

Manifest under test:

- [external-github-action-wave-prometheus-medium-service-prewarm.json](external-github-action-wave-prometheus-medium-service-prewarm.json)

Hosted validation runs:

- initial failed validation: run `27854409001`
- corrected successful validation: run `27854593611`
- successful run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27854593611`

Committed hosted summary artifact:

- [2026-06-20-prometheus-medium-service-prewarm-summary.json](2026-06-20-prometheus-medium-service-prewarm-summary.json)

## What Failed First

The first validation run did **not** fail because module prewarm itself was
broken.

It failed because the manifest pinned:

- `action_ref: v0.4.2`
- `install_path: github-action@v0.4.2`

while the branch workflow had already started passing the new
`--prewarm-modules` flag into `build-wave-result`.

That produced a helper compatibility failure in `Extract wave result`:

- `flag provided but not defined: -prewarm-modules`

So the first correction for `#331` was to make the demonstration manifest use
the current branch action source on purpose, instead of mixing a new workflow
with an older checked-out helper surface.

That compatibility bug is now tracked separately in
[#332](https://github.com/cervantesh/cervo-mutants/issues/332).

## Corrected Hosted Validation

The successful run kept the same three Prometheus jobs and only changed the
action source selection so the helper and workflow stayed aligned.

Hosted result summary:

| Job | Target | prewarm_modules | Generated | Effective | Failure kind | Outcome |
| --- | --- | --- | ---: | ---: | --- | --- |
| `prometheus-labels-cold` | `./model/labels` | `false` | `8` | `5` | none | workflow stayed healthy |
| `prometheus-rules-prewarmed` | `./rules` | `true` | `0` | `0` | `runner_error` | baseline timed out before mutation |
| `prometheus-storage-remote-prewarmed` | `./storage/remote` | `true` | `0` | `0` | `runner_error` | baseline timed out before mutation |

Two facts matter here:

1. The hosted workflow now finishes successfully with a mixed cold/prewarmed
   matrix, and the artifact summary records `prewarm_modules=true` for the two
   selected jobs.
2. Prewarming modules did **not** by itself make `./rules` or
   `./storage/remote` reach mutation within this bounded hosted lane.

## What The Successful Run Proves

The run is enough to validate the `#331` feature itself:

- manifest-driven opt-in works
- prewarm execution steps complete on hosted runners
- wave artifacts and aggregate summary retain the `prewarm_modules` bit
- mixed matrices with and without prewarm remain reportable

In other words, the optional prewarm path is now a real, auditable surface of
`external-action-wave`, not just a planned knob.

## What It Does Not Prove

This run does **not** show that prewarm is sufficient to recover every
dependency-heavy target.

For both prewarmed Prometheus candidates, the failure shape remained:

- `runner_error: baseline tests failed before mutation`
- underlying runner detail: `go test -coverprofile ... ./...`
- status reason: `test command timed out`

So the feature is validated, but the underlying hosted timeout shape for those
two packages still exists.

## Decision

Consider `#331` functionally validated once this evidence lands:

- optional module prewarm exists
- the choice is visible in wave artifacts
- the hosted workflow completes successfully with the new metadata path

Do **not** overstate the outcome:

- this was a feature validation success
- it was **not** a signal that prewarm alone fixes the current Prometheus
  medium-service timeout candidates

## Follow-Up

Keep the follow-up split clear:

- `#331`: add and validate optional module prewarm in the hosted wave workflow
- [#332](https://github.com/cervantesh/cervo-mutants/issues/332): decouple
  workflow helper execution from the pinned `action_ref` under test so older
  released action refs can still be summarized by newer workflow logic
