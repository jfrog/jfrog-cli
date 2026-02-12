# prepare-cli-release

Helper script that locally mirrors the GitHub Actions **Prepare Release** workflow. It bumps the CLI version, updates deps, builds, pushes a branch, and opens a PR.

## Prerequisites
- Python 3.8+ (runs the script).
- Go toolchain available (or provide `--go-home` to a GOROOT).
- `gh` CLI authenticated for `jfrog/jfrog-cli` (or export `GH_TOKEN`/`GITHUB_TOKEN`).
- Git workspace clean enough to create/push a branch.
- Make/build tooling used by `make update-all` and `build/build.sh`.

## Steps the script performs
1) Compute current version: `go run main.go -v`
2) Compute next version from the latest release tag via `gh release view`
3) Create branch `bump-ver-from-<current>-to-<next>` from `--ref` (default `master`)
4) Run `build/bump-version.sh <next>` with `BUMP_VERSION_SKIP_GIT=true`
5) Run `make update-all`
6) Build/check binary via `./build/build.sh` and `./jf --version`
7) Commit and push the branch
8) Open a PR titled `Bump version to <next>`
9) Annotate workflow and add the `ignore for release` label

## Usage
Run from the repo root (the script checks for `go.mod` with module `github.com/jfrog/jfrog-cli`):

```bash
.github/scripts/prepare-cli-release --version minor --ref master
```

### Key flags
- `--version {minor,patch}`: bump type (default: `minor`).
- `--ref <branch>`: base branch to start from (default: `master`).
- `--starting-step <step>`: resume from a specific step; earlier steps are skipped but required values are recomputed when needed.
- `--gh-token <token>`: GitHub token; otherwise uses `GH_TOKEN`/`GITHUB_TOKEN` or existing `gh` auth.
- `--dry-run`: print the planned actions and computed versions; no mutations.
- `--go-home <path>`: set `GOROOT` and prepend its `bin` to `PATH`.

### Common flows
- Minor release from master:
  ```bash
  .github/scripts/prepare-cli-release --version minor --ref master
  ```
- Patch release from a maintenance branch:
  ```bash
  .github/scripts/prepare-cli-release --version patch --ref release/3.2
  ```
- Dry run to preview:
  ```bash
  .github/scripts/prepare-cli-release --version minor --dry-run
  ```
- Resume after a previous partial run (e.g., start at dependency update):
  ```bash
  .github/scripts/prepare-cli-release --starting-step update-dependencies
  ```

## Outputs
- Branch: `bump-ver-from-<current>-to-<next>` pushed to `origin`.
- PR: Created against the base branch with title/body `Bump version to <next>`.
- Workflow notice printed with PR link and version.

## Troubleshooting tips
- Auth errors from `gh`: ensure `gh auth status` succeeds or pass `--gh-token`.
- Go not found: install Go or provide `--go-home /path/to/go`.
- Push failures: verify remote `origin` exists and you have permission to push.
- Missing tools: ensure `make`, `build/bump-version.sh`, and `build/build.sh` are executable in the repo root.

