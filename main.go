package main

func main() {
	workflow := New()
	workflow.Run("./samples/workflows.yaml")
}
