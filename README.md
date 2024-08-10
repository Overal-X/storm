# Formatio Storm

Storm is an automation agent that helps to run workflows on remote or local machines.

# Features

- [x] Read YAML configuration similar to GitHub workflow
- [ ] CLI to send workflow to machines
- [ ] Agent to execute workflows
- [ ] Machine inventory
- [ ] Execute commands as sudo user
- [ ] Provide inventory to store sudo user password
- [ ] Provide inventory to store ssh user password

# Bug

- [ ] Execute workflow jobs in the order they appear

# Setup

```sh
$ git clone git@github.com:Overal-X/formatio.storm.git
$ go mod tidy
$ go build -o storm
```

# Usage

With the example files

Run against remote machines from inventory

```sh
$ ./storm agent install -i ./samples/basic/inventory.yaml
$ ./storm agent run -i ./samples/basic/inventory.yaml ./samples/basic/workflow.yaml
```

Run worklow on current host

```sh
$ ./storm run ./samples/basic/workflow.yaml
```
