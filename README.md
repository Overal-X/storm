# Formatio Storm

Storm is an automation agent that helps to run workflows on remote or local machines.

# Useful links
- [Storm vs GitHub Actions (self-hosted runner)](https://external.ink?to=https://www.linkedin.com/posts/struckchure_devops-automatio-github-activity-7234504216193470465-Psf9)
- [Storm vs Ansible (Coming soon!)](#)
- [Storm deployment on EC2](https://external.ink?to=https://github.com/struckchure/storm-with-github-workflow)

# Installation

For use in golang

```sh
go get https://github.com/overal-x/formatio.storm
```

For Linux and MacOS

```sh
curl -fsSL https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.sh | bash
```

For Windows

```sh
irm https://raw.githubusercontent.com/Overal-X/formatio.storm/main/scripts/install.ps1 | iex
```

Or download binaries from [release page](https://github.com/Overal-X/formatio.storm/releases)

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
