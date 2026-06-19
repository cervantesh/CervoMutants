# Adoption Analytics And Feedback Loops

Tracking issue: #190

This document defines how external adoption evidence becomes durable product
signal instead of fading into issue threads, release notes, or maintainer
memory.

It complements the intake path in
[docs/feedback-intake.md](feedback-intake.md) and the maintainer operating
baseline in
[docs/maintainer-operations-pack.md](maintainer-operations-pack.md).

## Canonical Unit Of Evidence

The default unit of adoption evidence is one GitHub issue created with the
[`Adoption feedback`](../.github/ISSUE_TEMPLATE/adoption-feedback.yml)
template, backed by report artifacts and environment details.

Each issue should capture enough structure to answer:

- what kind of repository was being rolled out
- which adoption stage the team was in
- which install path and environment were used
- what the primary friction or blocker actually was
- whether the outcome points to docs, workflow, or product work
- whether the finding ever reached an upstream maintainer and what the current
  response state is

That gives maintainers something they can aggregate later without inventing a
parallel spreadsheet or private tracker.

When maintainers want a dated rollup from existing GitHub issues, they can
export the issue set and feed it to the helper workflow:

```powershell
gh issue list --state all --search "\"Adoption Feedback\" in:title" --json number,title,state,url,closedAt,body > adoption-issues.json
go run ./cmd/actionhelper build-adoption-summary --issues-json adoption-issues.json --tracking-issue "#313" > adoption-summary.json
go run ./cmd/actionhelper render-adoption-summary-markdown --path adoption-summary.json > adoption-summary.md
```

That path is intentionally JSON-in and markdown-out so release reviews can keep
their issue-query step explicit while still reusing one tested summarizer.

## Structured Dimensions To Capture

Every adoption-feedback issue should preserve these dimensions explicitly:

- repository profile: compact library, medium service, large repo/monorepo, or
  another clearly named shape
- adoption stage: first dry run, first useful report, baseline setup, PR lane,
  nightly lane, history/governance, or comparison/benchmark work
- install path: `go install`, release archive, local source build, or GitHub
  Action
- environment: OS, Go version, local vs CI, and any important runtime limits
- rollout posture: policy, budget, workers, baseline/quarantine posture,
  actionable-only usage, and report surfaces
- primary blocker class: signal noise, review UX, runtime/resources,
  CI/install/platform, governance/history, or unsupported workflow
- observed outcome: docs gap, workflow friction, product defect, or evidence of
  a known limit
- upstream response linkage: linked upstream issue/discussion, current external
  response status, and last checked date when that state is being tracked

Those dimensions should come from the issue itself, not from later maintainer
guesswork.

## Derived Metrics That Matter

The goal is not vanity reporting. The goal is to make external adoption
friction measurable enough to prioritize product work.

The default metrics to derive from accumulated adoption-feedback issues are:

- time to first useful report: how long it takes from initial install to a run
  that the adopter considers reviewable
- first-run success rate: whether bounded dry run or first useful report worked
  without repo-specific surgery
- baseline progression rate: how often a team gets from first run to
  baseline-first governance
- primary blocker frequency: which blocker class repeats most often
- repeated noisy-signal patterns: which operator families, semantic groups, or
  report surfaces keep generating "review pain"
- docs vs product follow-up ratio: how much friction is solved by clearer
  guidance versus code changes
- unsupported-workflow rate: how often adopters are trying to use the product
  outside its public support boundary
- upstream-thread creation rate: how often real rollout findings are carried
  into an outward maintainer-facing thread
- direct-response rate: how often those outward threads receive any public
  maintainer response
- silence-pattern duration: how often maintainers revisit an outward thread and
  still find no response

These are release-planning metrics, not marketing metrics.

## Feedback Loops

### Per-Issue Loop

For every adoption-feedback issue:

1. confirm the artifact bundle and environment details
2. classify the repository profile, adoption stage, and primary blocker
3. decide whether the outcome is:
   - already explained by current docs
   - a documentation follow-up
   - a product or code follow-up
   - an unsupported workflow
4. if the finding is taken upstream or discussed with an external maintainer,
   record the upstream thread link, current response state, and last checked
   date in the issue body
5. link repeated findings into
   [docs/evaluations/follow-up-ledger.md](evaluations/follow-up-ledger.md)

### Release Loop

Before each release, review adoption-feedback issues opened or updated since
the previous release and summarize:

- repeated blocker classes
- repeated noisy operator or semantic-group complaints
- rollout steps that still create avoidable setup friction
- docs clarifications that should ship with the release
- product issues that should remain explicitly narrowed rather than silently
  implied as fixed
- how many issues gained an upstream thread
- how many outward threads received maintainer response versus remaining silent

That review belongs in release preparation, not as an afterthought.

### Validation-Wave Loop

When running a public validation wave, compare the wave findings with the
historical adoption-feedback issues:

- did the same blockers repeat across multiple repository profiles?
- did the closest rollout playbook materially reduce friction?
- are the same unsupported expectations still appearing?
- are maintained examples still sufficient for the current release?

If the answer is yes, link the evidence into the follow-up ledger and open or
refresh the corresponding tracked work.

### Outward-Response Loop

When a structured adoption-feedback issue captures a real external rollout
finding, maintainers should decide deliberately whether that evidence needs an
outward maintainer-facing thread instead of leaving the issue as internal proxy
evidence only.

Use this loop:

1. confirm the issue has enough evidence to survive outside this repository:
   clear target repository, artifact links, blocker class, and the specific
   docs or product question being raised
2. choose the outward channel on purpose:
   upstream issue for actionable bug, docs, or workflow friction;
   public discussion when the question is interpretation, rollout posture, or
   support-boundary clarity
3. open at most one outward thread for one finding family, then record that URL
   in the adoption-feedback issue immediately
4. update the issue body to `Upstream thread opened, no maintainer response
   yet` and set `External response last checked` on the same day the outward
   thread is created
5. recheck on a bounded cadence instead of informal memory:
   first check within 3 business days, second check before the next release
   review, and any later check when a maintainer responds or the thread is
   closed
6. classify the response using the existing structured states:
   no upstream thread, thread opened/no response, maintainer replied or asked
   follow-up, maintainer accepted, maintainer rejected or narrowed, or response
   state not checked recently
7. if multiple adoption-feedback issues map to the same outward thread, link
   them all to the same URL and promote the shared pattern into
   [docs/evaluations/follow-up-ledger.md](evaluations/follow-up-ledger.md)
8. stop treating silence as "unknown" after repeated checks:
   preserve it as explicit evidence of non-engagement rather than repeatedly
   resetting the same issue without new outward movement

The goal of the outward loop is not to maximize issue volume. It is to make
direct engagement, explicit rejection, and durable silence all auditable from
the same structured issue trail.

## Promotion Rules For Repeated Findings

Move a finding from a single issue into the explicit ledger when any of these
are true:

- the same blocker appears in at least two adoption-feedback issues
- the same docs gap appears across different repository profiles
- the same noisy operator or review surface keeps surfacing in real rollouts
- a known limit keeps getting interpreted as supported behavior

When promoted, record:

- the evidence source
- the repeated finding in one sentence
- the priority
- the current status
- the linked issue, doc, or narrowing decision

## What Not To Do

Do not treat one anecdote as broad adoption proof.

Do not collapse every complaint into "noise" without preserving the repository
shape, adoption stage, and artifact context.

Do not rely on maintainers to remember which complaints repeated last month.

Do not create a separate private analytics system before the public issue and
artifact path is working.

## Related Guides

- [docs/feedback-intake.md](feedback-intake.md)
- [docs/maintainer-operations-pack.md](maintainer-operations-pack.md)
- [docs/adoption-guide.md](adoption-guide.md)
- [docs/rollout-playbooks.md](rollout-playbooks.md)
- [docs/evaluations/follow-up-ledger.md](evaluations/follow-up-ledger.md)
