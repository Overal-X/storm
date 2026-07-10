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
curl -fsSL https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | bash
```

For Windows

```sh
irm https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.ps1 | iex
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
				uses: Overal-X/storm/actions/setup-storm@v0.1.1
			- name: Check Storm
				run: storm version
```

Compatibility note:

- The legacy subpath also works: `Overal-X/storm/.github/actions/setup-storm@v0.1.1`
- Prefer pinning to a tag or commit SHA instead of `@main` in production workflows.

Single-file deploy action (setup SSH + setup Storm + deploy):

```yaml
jobs:
	deploy:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v4
			- name: Storm deploy
				uses: Overal-X/storm/actions/storm-deploy@v0.1.1
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

# Development

```sh
git clone git@github.com:Overal-X/formatio.storm.git
go mod tidy
go run ./cmd help
```
