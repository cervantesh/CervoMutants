# Upgrade Notes

Each public release should have a matching file in this directory:

- `docs/upgrade-notes/vX.Y.Z.md`

The release workflow uses that file together with the matching `CHANGELOG.md`
section to build the GitHub release body.

Keep upgrade notes focused on:

- behavior changes operators should care about
- config migrations
- compatibility shifts
- policy or report changes that can affect CI

Each release note file should use this structure:

```markdown
# Upgrade Notes for vX.Y.Z

## Summary

- What changed at a high level.

## Operator Action

- What maintainers, CI owners, or consumers should check or update.

## Rollback

- How to return to the previous known-good version if the rollout fails.
```

The release workflow verifies that the matching file for a tagged version
contains these sections before publishing.
