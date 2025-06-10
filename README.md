# git-svc

git-svc is a small CLI tool that wraps `git worktree` operations and
creates symlinks for local development. Worktrees are added with
`git worktree add --no-checkout` and then configured using
`git sparse-checkout` so that only the specified directory is
checked out. This keeps worktrees lightweight while
ensuring existing tools (like `docker-compose`) keep working without
changes.

## Installation

```
go install github.com/yudppp/git-svc@latest
```

## Usage

```
# add worktree for branch feature and link packages/a
# only packages/a will be checked out using git sparse-checkout
git svc init packages/a feature

# pull latest changes for the worktree linked from packages/a
git svc pull packages/a

# remove the worktree and symlink
git svc clean packages/a

# list managed symlinks
git svc list
```

## Using as a library

Fetch the module with `go get` and import the `svc` package:

```bash
go get github.com/yudppp/git-svc
```

```go
import "github.com/yudppp/git-svc/svc"
```

## Configuration

Set a custom worktree root with the `--worktree-root` flag or the
`GITSVC_WORKTREE_ROOT` environment variable. Worktrees are created under
`.worktrees` by default.

Example:

```bash
$ GITSVC_WORKTREE_ROOT=_trees git svc init packages/b other-branch
```
Only `packages/b` will exist in the created worktree thanks to
`git worktree add --no-checkout` and `git sparse-checkout`.

### Typical workflow

1. Initialize a branch-specific worktree linked to your service
   (only that directory will be checked out):
   ```bash
   git svc init packages/a feature
   ```
2. Keep the worktree up to date:
   ```bash
   git svc pull packages/a
   ```
3. Remove the symlink and worktree when the branch is merged or no longer needed:
   ```bash
   git svc clean packages/a
   ```
4. List active links and their branches:
   ```bash
   git svc list
   ```

### Windows notes

Creating symlinks on Windows requires Administrator privileges or using
WSL. If `git svc init` fails with a permission error, try running the
command prompt as Administrator or perform the operation inside WSL.
