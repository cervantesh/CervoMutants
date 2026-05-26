# CervoMutant Evaluation Framework

Tracking issues:

- https://gitea.cervbox.synology.me/CervoSoft/cervo-mutant/issues/5
- https://gitea.cervbox.synology.me/CervoSoft/cervo-mutant/issues/6

This framework evaluates mutation testing tools for CervoSoft and decides
whether `cervo-mutant` should be the default for CervoClaw and related Go
libraries. It combines published mutation-testing tool comparison work with
engineering criteria needed for CI, large Go modules, and agent-driven test
improvement.

The decisive question is:

> Does the tool produce actionable survivors, with low noise, in acceptable
> time, and do those survivors lead to tests that detect real or realistic
> faults?

## Research Basis

Use these studies as the methodological baseline:

- **How Do Java Mutation Tools Differ?**  
  CACM tool-comparison framework. Use its five dimensions: tool version,
  deployment, mutation process, user-centric features, and mutation operators.
- **How effective are mutation testing tools?**  
  Empirical Software Engineering study using real faults and manual mutant
  analysis. Use it to judge fault-revealing effectiveness instead of relying on
  mutation score alone.
- **On the use of commit-relevant mutants**  
  Empirical Software Engineering study on commit-scoped mutation. Use it to
  judge CI and pull-request usefulness.
- **Selecting fault revealing mutants**  
  Empirical Software Engineering study on reducing mutant sets while preserving
  fault-revealing power. Use it for mutator profiles, sampling, and selection.
- **Mutation Testing Advances: An Analysis and Survey**  
  Use as a threats-to-validity checklist for equivalent mutants, flaky tests,
  sampling bias, operator bias, and over-reliance on mutation score.

Use these recent studies as extensions to the baseline:

- **Equivalent Mutants in the Wild**  
  ISSTA 2024 study on identifying and suppressing equivalent mutants in real
  Java projects. Use it to evaluate equivalent-mutant suppression, false
  suppression risk, and noisy-operator behavior.
- **An Empirical Evaluation of Manually Created Equivalent Mutants**  
  Mutation 2024 / ICST 2024 study on human-created and human-identified
  equivalent mutants. Use it to avoid over-trusting single-reviewer equivalent
  classifications.
- **On Comparing Mutation Testing Tools through Learning-based Mutant
  Selection**  
  2023 study on comparing mutation tools through selected mutants. Use it to
  measure complementarity and unique useful survivors instead of rewarding raw
  mutant volume.
- **An Exploratory Study on Using Large Language Models for Mutation Testing**  
  2024 study on LLM-generated mutations over real-bug benchmarks. Use it to
  evaluate agent-facing explanations, AI-generated test value, and uncompilable
  mutant rates.
- **Latent Mutants: A large-scale study on the Interplay between mutation
  testing and software evolution**  
  2025 study on mutation testing across project releases. Use it to track
  long-standing survivors and determine whether survivors predict future
  changes, bugs, or persistent test debt.
- **A fine-grained evaluation of mutation operators to boost mutation testing
  for deep learning systems**  
  2025 study introducing operator quality and redundancy-style measures. Use
  the metric pattern, not the DL domain, to evaluate whether each Go mutator
  belongs in `conservative`, `default`, or `aggressive`.
- **METFORD: Mutation Testing Framework for Android** and **Mutta: a novel tool
  for E2E web mutation testing**  
  Recent framework/tool studies outside Go. Use them only for operational
  evaluation ideas: reproducible scripts, execution-strategy comparisons, and
  high-cost integration/E2E mutation tradeoffs.

These studies are technology-independent at the evaluation level. They should
not be copied as Java/JVM implementation details. Translate their concepts into
Go terms: classes become packages/files/functions, JUnit tests become `go test`
commands, bytecode mutants become AST/source patches, and build modules become
Go modules.

## Scorecard

Score each tool out of 100. Keep per-layer scores visible; do not hide a weak
layer behind a strong total score.

| Layer | Weight |
| --- | ---: |
| Tool capability | 20 |
| Fault-revealing effectiveness | 25 |
| CI and commit relevance | 15 |
| Actionability and agent utility | 15 |
| Cost and scalability | 8 |
| Noise and equivalent-mutant burden | 10 |
| Longitudinal and evolution relevance | 4 |
| Validity controls | 3 |

### Tool Capability - 20

Based on the CACM five-dimension framework.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Versioning and reproducibility | 3 | Version, config, schema, and toolchain are recorded in reports. |
| Deployment and CI setup | 4 | Install command, CI command, exit codes, and no hidden services required. |
| Mutation process | 5 | Clear discovery, mutation, isolation, selection, execution, and classification flow. |
| User-centric features | 4 | `affected`, `show`, `explain`, `report survivors`, HTML, JSON, JUnit. |
| Mutation operators | 4 | Operators are documented by profile, risk, node type, and example diff. |

### Fault-Revealing Effectiveness - 25

Based on real-fault and manual-analysis studies.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Real or realistic fault detection | 8 | Tests added from survivors detect historical or seeded faults. |
| Survivor-to-test yield | 5 | Percentage of survivors that lead to useful new assertions. |
| Oracle-strength improvement | 4 | New tests fail on the mutant and pass on the original code. |
| Manual mutant analysis | 4 | Sampled mutants classified as useful, redundant, equivalent, or invalid. |
| Complementarity | 4 | Tool finds useful gaps missed by coverage, existing tests, or other mutation tools. |

### CI And Commit Relevance - 15

Based on commit-relevant mutant research.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Changed-scope correctness | 4 | `affected --scope changed` matches changed modules/packages/files. |
| Cost reduction versus full run | 4 | PR run time and mutant count are materially lower than full mutation. |
| Signal preservation | 3 | Commit-scoped runs still identify relevant survivors. |
| Feedback latency | 2 | PR feedback arrives within the chosen CI budget. |
| False-negative control | 2 | Periodic full runs or sampling detect what changed-scope may miss. |

### Actionability And Agent Utility - 15

This layer is CervoSoft-specific because `cervo-mutant` is intended for both
humans and coding agents.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Stable JSON contract | 3 | `schema_version`, documented enums, additive-only v1 fields, golden tests. |
| Survivor context | 3 | File, line, function, package, operator, diff, selected tests, output, reason. |
| `show` and `explain` usefulness | 3 | A survivor can be turned into a concrete test task without extra discovery. |
| Determinism | 2 | Same seed, config, and inputs produce stable mutant IDs and ordering. |
| Low-context reports | 2 | Reports are compact enough for agents while retaining necessary evidence. |
| Grouping and prioritization | 2 | Survivors can be sorted by package, operator, changed code, or prior history. |

### Cost And Scalability - 8

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Time budget support | 2 | `--budget` stops work predictably and reports skipped mutants. |
| Incremental cache value | 2 | Cache hit rate and time saved are reported and reproducible. |
| Test selection value | 2 | Package/coverage selection reduces test work without losing signal. |
| Workspace isolation cost | 1 | Temp-workdir overhead is measured, especially on Windows/OneDrive. |
| Large-project threshold behavior | 1 | Behavior is measured at 100+ packages, 5+ minute tests, or 1,000+ mutants. |

### Noise And Equivalent-Mutant Burden - 10

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Equivalent rate | 2 | Manual sample estimates equivalent survivors by mutator/profile. |
| Suppression precision | 2 | Automatically suppressed or ignored mutants are sampled for false suppression. |
| Triage time | 2 | Median time to classify or act on a survivor. |
| Redundancy | 1 | Mutants killed by the same test or producing same signal are measured. |
| Quarantine discipline | 2 | Quarantine requires reason, owner, issue, expiry, and does not inflate score. |
| Repeat offenders | 1 | Noisy operators are identified and moved to stricter profiles or disabled. |

Equivalent-mutant classifications must be graded by evidence quality:

| Evidence level | Meaning |
| --- | --- |
| Preliminary | One reviewer classified the mutant. |
| Accepted | Two reviewers agree, or a reproducible rule proves equivalence. |
| Disputed | Reviewers disagree or the classification depends on undocumented assumptions. |
| Suppressed | The tool automatically suppressed it; false-suppression risk must be sampled. |

Do not treat a single human classification as definitive evidence. Record
reviewer disagreement rate when manual classification influences adoption.

### Longitudinal And Evolution Relevance - 4

Based on latent-mutant and software-evolution studies.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Long-standing survivor tracking | 1 | Survivors persisting across releases are tracked by stable identity. |
| Survivor aging policy | 1 | Old survivors are resolved, quarantined with expiry, or intentionally rejected. |
| Evolution alignment | 1 | Survivors are checked against later changes, bugs, or contract updates. |
| Debt and release trend | 1 | Active survivors, quarantine count, and scores are comparable across releases. |

### Validity Controls - 3

Use this layer to keep evaluation claims defensible.

| Criterion | Points | Evidence |
| --- | ---: | --- |
| Baseline stability | 1 | Baseline tests pass repeatedly before mutation. |
| Flaky-test control | 1 | Timeouts/flaky failures are rerun or classified separately. |
| Sampling, operator bias, and threats | 1 | Seed, operator/profile breakdown, and known limitations are recorded. |

## Required Metrics

Collect these metrics for every evaluation run:

- mutation score and effective score excluding active quarantine
- total, killed, survived, timed out, compile error, skipped, ignored,
  quarantined, and cached mutants
- runtime, budget usage, cache hits, cache misses, and test-selection mode
- number of survivors converted into tests
- number of survivors classified as equivalent, redundant, invalid, or useful
- equivalent suppression precision and sampled false-suppression count
- unique actionable survivors not found by other tools or profiles
- reviewer disagreement rate for equivalent-mutant classifications
- long-standing survivor rate across releases
- survivor age distribution
- operator quality score and operator redundancy score
- cost per killed mutant
- cost per actionable survivor
- cost per real or realistic fault revealed
- quarantine active count, expired count, and growth over time
- changed-scope mutant count versus full-run mutant count

Do not use global mutation score as the primary success criterion. It is a
supporting metric, not the decision rule.

## Empirical Protocol

Use two evaluation levels.

### Level 1: Tool Comparison

Compare `cervo-mutant` against available Go mutation tools such as Gremlins,
go-mutesting variants, and gomu when they can run on the same target.

For each tool:

1. Record version, command, config, Go version, OS, CPU count, and commit SHA.
2. Run baseline tests repeatedly until stable.
3. Run mutation on the same packages with comparable timeout and budget.
4. Export raw reports and normalize results into the scorecard.
5. Manually classify a representative survivor sample.
6. Record setup friction, failures, unsupported cases, and workspace artifacts.

### Level 2: CervoClaw Adoption Study

Use CervoClaw and at least one CervoSoft library as target projects.

For each target:

1. Run `cervomut affected ./... --scope changed --since origin/main`.
2. Run `cervomut run ./... --scope changed --since origin/main --budget <N>`.
3. Run a full or sampled baseline mutation run.
4. Select survivors from changed-scope and full/sampled runs.
5. Ask a human or agent to write tests from survivor reports.
6. Verify each new test fails against the mutant and passes against original
   code.
7. Check whether those tests detect historical bugs, seeded realistic bugs, or
   realistic contract violations.
8. Record time spent, false starts, equivalent mutants, and test value.

### Level 3: Longitudinal Default-Readiness Study

Use this level before declaring `cervo-mutant` the default across CervoSoft.

1. Keep stable mutant identities across releases when source locations shift
   only minimally.
2. Track survivors that persist for more than one release or evaluation cycle.
3. Classify persistent survivors as useful debt, equivalent, redundant,
   flaky-test related, or intentionally accepted risk.
4. Check whether later bug fixes, contract changes, or production incidents
   touch code associated with long-standing survivors.
5. Report survivor age, quarantine age, expired quarantine, and resolved
   survivor counts.
6. Use the trend to decide whether the tool is improving test quality or merely
   accumulating mutation debt.

## Operator Quality And Redundancy

Evaluate mutators independently before promoting them between profiles.

For each mutator, measure:

- generated mutant count
- killed count
- survivor count
- equivalent rate
- compile-error rate
- timeout rate
- unique actionable survivor count
- redundancy with other mutators
- median triage time
- real or realistic fault detection yield

Use these derived scores:

```text
operator_quality =
  unique_actionable_survivors / generated_mutants

operator_redundancy =
  redundant_or_duplicate_signal_mutants / generated_mutants
```

Promotion rules:

- `conservative`: low compile-error rate, low equivalent rate, high
  actionability, and low triage cost.
- `default`: useful but noisier operators, including domain-sensitive
  mutations such as controlled error-return changes.
- `aggressive`: exploratory operators with high cost, high equivalence risk, or
  high redundancy.

## AI And Agent-Facing Evaluation

When evaluating agent utility, distinguish three cases:

1. The tool generated a useful mutant.
2. The agent understood the survivor and wrote a meaningful test.
3. The new test detected a real or realistic fault.

All three are required for strong evidence. A good explanation that leads to a
test that only kills an artificial mutant is useful, but weaker than a test that
also detects historical or seeded realistic faults.

Track:

- uncompilable mutant rate for any AI-generated or AI-suggested mutations
- prompt/explanation length needed for an agent to act
- number of agent attempts per useful test
- tests that kill the mutant but assert implementation details
- tests that kill the mutant and improve public behavior coverage

## Acceptance Guidance

`cervo-mutant` can be considered the default only if:

- it scores at least 80 overall,
- it scores at least 18/25 on fault-revealing effectiveness,
- it scores at least 11/15 on actionability and agent utility,
- it has an accepted equivalent-classification process for sampled survivors,
- it tracks long-standing survivors and quarantine growth across releases,
- it does not create untracked artifacts in target workspaces when `--out` is
  used,
- quarantine does not grow without issue-backed cleanup,
- and survivors regularly lead to useful tests rather than mostly equivalent or
  redundant findings.

Reject or defer default adoption if the tool is fast but not actionable, has a
high equivalent-mutant burden, produces unstable CI results, or improves
mutation score without improving real-fault detection.
