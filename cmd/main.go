package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "storm",
	Short: "Formatio Storm",
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Storm agent commands",
}

var agentRunWorkflowCmd = &cobra.Command{
	Use:  "run",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowFile := args[0]
		inventoryFile, _ := cmd.Flags().GetString("inventory")

		agent := NewAgent()
		err := agent.RunWithFiles(inventoryFile, workflowFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var agentInstallCmd = &cobra.Command{
	Use: "install",
	Run: func(cmd *cobra.Command, args []string) {
		inventoryFile, _ := cmd.Flags().GetString("inventory")
		installationMode, _ := cmd.Flags().GetString("mode")

		agent := NewAgent()
		err := agent.Install(inventoryFile, installationMode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var agentUninstallCmd = &cobra.Command{
	Use: "uninstall",
	Run: func(cmd *cobra.Command, args []string) {
		inventoryFile, _ := cmd.Flags().GetString("inventory")

		agent := NewAgent()
		err := agent.Uninstall(inventoryFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var runWorkflowCmd = &cobra.Command{
	Use:  "run",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowFile := args[0]
		trashWorkflow, _ := cmd.Flags().GetBool("trash-workflow")

		if trashWorkflow {
			defer os.Remove(workflowFile)
		}

		workflow := NewWorkflow()
		err := workflow.RunWithFile(workflowFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func main() {
	agentInstallCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentInstallCmd.Flags().StringP("mode", "m", "prod", "formatio storm installation type (prod or dev)")
	agentCmd.AddCommand(agentInstallCmd)

	agentUninstallCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentCmd.AddCommand(agentUninstallCmd)

	agentRunWorkflowCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentCmd.AddCommand(agentRunWorkflowCmd)

	runWorkflowCmd.Flags().BoolP("trash-workflow", "t", false, "remove workflow file if the workflow is complete")
	rootCmd.AddCommand(runWorkflowCmd)

	rootCmd.AddCommand(agentCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err) // TODO: use logger
		os.Exit(1)
	}
}
