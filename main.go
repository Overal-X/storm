package main

import "storm.formatio.org/service"

func main() {
	workflow := service.WorkflowService("./samples/workflows.yaml")

	workflow.Run()
}
