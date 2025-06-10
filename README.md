# git-svc

git-svc is a small CLI tool that wraps `git worktree` operations and
creates symlinks for local development. Each linked directory is backed
by its own worktree created with `git worktree add`. All files for the
branch are checked out so existing tools (like `docker-compose`) keep
working without any special configuration.

## Installation

```
go install github.com/yudppp/git-svc@latest
```

## Usage

```
# add worktree for an existing branch and link packages/a
git svc init packages/a feature

# only checkout packages/a using sparse checkout
git svc init --sparse packages/a feature

# create a new branch from origin/main and link packages/a
git svc init packages/a -b feat-a origin/main

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
$ GITSVC_WORKTREE_ROOT=_trees git svc init --sparse packages/b other-branch
```
This creates a worktree under `_trees/other-branch` and links
`packages/b` to it while only checking out that directory.

### Typical workflow

1. Initialize a branch-specific worktree linked to your service:
   ```bash
   git svc init --sparse packages/a -b feature
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

### Git ignore recommendations

When using git-svc, you'll typically want to add the managed symlinks to your `.gitignore` file to avoid committing them to your repository. The symlinks are meant for local development only.

For example, if you're managing `packages/service-a`, add this to your `.gitignore`:

```
# git-svc managed symlinks
packages/service-a
```

Alternatively, you can ignore all symlinks with:

```
# Ignore all symlinks (be careful with this approach)
**/*
!**/*/*
```

### Windows notes

Creating symlinks on Windows requires Administrator privileges or using
WSL. If `git svc init` fails with a permission error, try running the
command prompt as Administrator or perform the operation inside WSL.

### FAQ

#### Can I use git-svc in CI/CD?

git-svc is designed for local development and is not typically needed in CI/CD environments. In CI, you usually want to work with the entire repository rather than isolated symlinks.

#### What happens if I accidentally commit a symlink?

If you commit a symlink created by git-svc, other developers will see a broken symlink since the worktree directories are local. Remove the symlink from git with `git rm <symlink>` and add it to `.gitignore`.

#### How do I migrate from an existing setup?

1. Ensure your current directory is backed up or committed
2. Run `git svc init <dir> <branch>` - git-svc will automatically backup your existing directory
3. The original directory will be renamed with `.gitsvc_backup` suffix
4. Use `git svc clean <dir>` to restore the original if needed
