# Release process

Releases are automated via the `Release` GitHub Actions workflow (`.github/workflows/release.yml`). Pushing a tag matching `v*` triggers it: it builds binaries with goreleaser, pushes Docker images to quay.io, and publishes the GitHub release with notes from `changelogs/${tag}.md`.

## Steps

1. **Open a PR** that adds:
   - An entry in `CHANGELOG.md` for the new version
   - A `changelogs/${tag}.md` file (used as the GitHub release body — see existing files for the format)

2. **Once the PR is merged**, create a signed tag and push it:

   ```bash
   tag=v0.3.0
   git checkout main && git pull origin main
   git tag -s "${tag}" -m "${tag}"
   git push origin "${tag}"
   ```

3. **Watch the workflow** at https://github.com/cortexproject/cortex-tools/actions

   On success:
   - Binaries for cortextool and benchtool (linux, mac-os, windows) are attached to the release
   - Docker images are pushed to `quay.io/cortexproject/cortex-tools:${tag}` and `quay.io/cortexproject/benchtool:${tag}`
   - The release notes include a Docker images section appended automatically

## Required secrets

The workflow uses two repository secrets:

- `QUAY_USERNAME`
- `QUAY_PASSWORD`

`GITHUB_TOKEN` is provided automatically by Actions.

`HOMEBREW_TAP_TOKEN` is **optional** — see below.

## Homebrew tap

`cortextool` is published through [`cortexproject/homebrew-tap`](https://github.com/cortexproject/homebrew-tap).
Nothing in this repository builds or pushes the formula, and `.goreleaser.yml`
does not know about Homebrew: **a release needs no Homebrew-specific steps.**

The tap updates itself. Its `Update cortextool` workflow reads our latest
release, rewrites the version and `sha256` with `brew bump-formula-pr`, then
builds the candidate from source and `brew test`s it on macOS arm64, macOS
x86_64, Linux x86_64 and Linux arm64, committing only if all four pass. It runs
on a 6-hourly schedule, so a new release is picked up within 6 hours with no
action from us.

The `notify-tap` job in `release.yml` just shortens that wait: it sends a
`repository_dispatch` so the bump starts within about a minute. The job is
`continue-on-error: true` and is skipped entirely when `HOMEBREW_TAP_TOKEN` is
unset, so **it can never fail a release** — if it is missing or its token has
expired, the tap's schedule still picks the release up.

`HOMEBREW_TAP_TOKEN` should be a fine-grained PAT scoped to
`cortexproject/homebrew-tap` only, with `Contents: write`.

To bump the tap by hand — after a retag, or if you would rather not wait:

```bash
gh workflow run "Update cortextool" --repo cortexproject/homebrew-tap
gh workflow run "Update cortextool" --repo cortexproject/homebrew-tap -f tag=v0.21.1
```

Note that retagging a release changes the checksums under a fixed version, which
users see as `SHA256 mismatch` until they `brew update` again. Prefer publishing
a new patch tag.
