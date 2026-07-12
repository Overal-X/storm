# Storm

Storm is an automation agent that helps to run workflows on remote or local machines.

# Useful links

- [Storm vs GitHub Actions (self-hosted runner)](https://www.linkedin.com/posts/struckchure_devops-automatio-github-activity-7234504216193470465-Psf9)
- [Storm vs Ansible (Coming soon!)](#)
- [Storm deployment on EC2](https://github.com/struckchure/storm-with-github-workflow)

# Installation

For use in golang

```sh
go get github.com/overal-x/storm
```

For Linux and MacOS

```sh
curl -fsSL https://raw.githubusercontent.com/overal-x/formatio.storm/main/scripts/install.sh | bash
```

For Windows

```sh
irm https://raw.githubusercontent.com/overal-x/formatio.storm/main/scripts/install.ps1 | iex
```

Or download binaries from [release page](https://github.com/overal-x/storm/releases)

# GitHub Action

Use the reusable setup action from another repository workflow:

```yaml
jobs:
	build:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v4
			- name: Setup Storm
				uses: overal-x/storm/actions/setup-storm@v0.1.1
			- name: Check Storm
				run: storm version
```

Compatibility note:

- The legacy subpath also works: `overal-x/storm/.github/actions/setup-storm@v0.1.1`
- Prefer pinning to a tag or commit SHA instead of `@main` in production workflows.

Single-file deploy action (setup SSH + setup Storm + deploy):

```yaml
jobs:
	deploy:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v4
			- name: Storm deploy
				uses: overal-x/storm/actions/storm-deploy@v0.1.1
				with:
					ssh_key: ${{ secrets.SSH_KEY_B64 }}
					github_token: ${{ secrets.GITHUB_TOKEN }}
					github_repo: ${{ github.repository }}
					github_ref: ${{ github.ref_name }}
					inventory: .storm/inventory.yaml
					workflow: .storm/workflow.yaml
```

# Usage

With the example files

Run against remote machines from inventory

```sh
storm agent install -i ./samples/basic/inventory.yaml
storm agent run -i ./samples/basic/inventory.yaml ./samples/basic/workflow.yaml
```

Run worklow on current host

```sh
storm run ./samples/basic/workflow.yaml
```

## Contexts

Workflows can reference context values with `${{ <name>.<key> }}` expressions
(for example `${{ github.REPO_NAME }}`). Contexts are supplied at run time and
work identically for both `storm run` and `storm agent run`.

**Inline** — repeatable `--context`/`-c` flag in `name:value` form, where the
value is JSON (or base64-encoded JSON with `--context-format base64`):

```sh
storm run -c 'github:{"REPO_NAME":"formatio.storm"}' ./samples/workflow-context.yaml
```

**From a file** — repeatable `--context-file` flag. Files may be **YAML or
JSON**, in one of two forms:

- Whole-contexts file — the file's top-level keys are the context names:

  ```sh
  storm run --context-file ./samples/context.yaml ./samples/workflow-context.yaml
  ```

- Named single-context file — the file holds one context's key/value map,
  assigned under `<name>`:

  ```sh
  storm run --context-file github:./gh.json ./samples/workflow-context.yaml
  ```

**Precedence** — `--context-file` provides the base; inline `--context` flags
override individual keys. Both flags may be repeated and combined.

## Defaults

A workflow-level `defaults` block sets fallbacks applied to every job step,
so you don't have to repeat `shell` or `directory` on each step:

```yaml
name: My workflow

defaults:
  directory: ./app
  shell: /bin/bash

jobs:
  - name: build
    runs-on: self-hosted
    steps:
      - name: install # runs in ./app with /bin/bash
        run: npm ci
      - name: test # overrides just the directory
        directory: ./app/tests
        run: npm test
```

Precedence for each step, highest first: the step's own `directory`/`shell`,
then `defaults`, then the workflow-level `directory` (for directory) or
`$SHELL` (for shell).

# Development

```sh
git clone git@github.com:overal-x/formatio.storm.git
go mod tidy
go run ./cmd help
```
