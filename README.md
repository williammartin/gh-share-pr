# gh-share-pr

A Go-based GitHub CLI extension that prints and copies a pull request as one
shareable Markdown line, including its repository and change counts.

## Install

Install the latest release with the [GitHub CLI](https://cli.github.com/):

```sh
gh extension install williammartin/gh-share-pr
```

To build and install from source, you also need Go 1.24 or later:

```sh
gh repo clone williammartin/gh-share-pr gh-share-pr
cd gh-share-pr
go build -o gh-share-pr .
gh extension install .
```

The release workflow publishes precompiled binaries when a `v*` tag is pushed.

Originally requested in [cli/cli#14618](https://github.com/cli/cli/issues/14618).

## Usage

```sh
gh share-pr
gh share-pr 123
gh share-pr https://github.com/cli/cli/pull/123
gh share-pr feature-branch
gh share-pr 123 -R cli/cli
```

With no pull request argument, the extension uses the pull request for the
current branch. You can also select a pull request by number, URL, or branch,
and use `-R OWNER/REPO` to target another repository.

The formatted line is written to stdout and copied to the system clipboard:

```text
[Fix the bug](https://github.com/cli/cli/pull/123) in cli/cli (+84/-8)
```

If clipboard access fails, the extension reports the problem on stderr while
leaving stdout pipeable.
