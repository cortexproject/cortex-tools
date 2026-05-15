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
