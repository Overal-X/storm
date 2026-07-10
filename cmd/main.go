package main

import (
	"fmt"
	"os"

	storm "github.com/Overal-X/storm"
	"github.com/spf13/cobra"
)

var (
	version   = "dev" // default value
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "storm",
	Short: "Formatio Storm",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This is the version of your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
	},
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
		format, _ := cmd.Flags().GetInt("format")
		contextFlags, _ := cmd.Flags().GetStringArray("context")
		contextFormat, _ := cmd.Flags().GetString("context-format")

		runOpts := []storm.RunOption{}

		agent := storm.NewAgent()
		runOpts = append(runOpts,
			agent.AgentWithFiles(workflowFile, inventoryFile),
			agent.AgentWithCallback(func(i any) { fmt.Println(i) }, format),
		)

		if len(contextFlags) > 0 {
			contexts, err := storm.ParseContextFlags(contextFlags, contextFormat)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			runOpts = append(runOpts, agent.AgentWithContexts(contexts))
		}

		err := agent.Run(runOpts...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Manage inventory with AES encryption",
}

var encryptInventoryCmd = &cobra.Command{
	Use: "encrypt",
	Run: func(cmd *cobra.Command, args []string) {
		inventoryFile, _ := cmd.Flags().GetString("inventory")
		encryptionKey, _ := cmd.Flags().GetString("encryption-key")

		encryptedInventory, err := storm.NewInventory().Encrypt(inventoryFile, encryptionKey)
		if err != nil {
			fmt.Println(err)

			os.Exit(1)
		}

		fmt.Println(*encryptedInventory)
	},
}

var decryptInventoryCmd = &cobra.Command{
	Use: "decrypt",
	Run: func(cmd *cobra.Command, args []string) {
		encryptedInventory, _ := cmd.Flags().GetString("encrypted-inventory")
		encryptionKey, _ := cmd.Flags().GetString("encryption-key")
		format, _ := cmd.Flags().GetString("format")

		var byteEncryptedInventory []byte
		var err error

		if format == "file" {
			byteEncryptedInventory, err = os.ReadFile(encryptedInventory)
			if err != nil {
				fmt.Println(err)

				os.Exit(1)
			}
		} else {
			byteEncryptedInventory = []byte(encryptedInventory)
		}

		decryptedInventory, err := storm.NewInventory().Decrypt(string(byteEncryptedInventory), encryptionKey)
		if err != nil {
			fmt.Println(err)

			os.Exit(1)
		}

		fmt.Println(*decryptedInventory)
	},
}

var agentInstallCmd = &cobra.Command{
	Use: "install",
	Run: func(cmd *cobra.Command, args []string) {
		inventoryFile, _ := cmd.Flags().GetString("inventory")
		installationMode, _ := cmd.Flags().GetString("mode")

		agent := storm.NewAgent()
		err := agent.Install(storm.InstallArgs{
			If:   inventoryFile,
			Mode: installationMode,
		})
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

		agent := storm.NewAgent()
		err := agent.Uninstall(storm.UninstallArgs{If: inventoryFile})
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
		directory, _ := cmd.Flags().GetString("directory")
		format, _ := cmd.Flags().GetInt("format")
		contextFlags, _ := cmd.Flags().GetStringArray("context")
		contextFormat, _ := cmd.Flags().GetString("context-format")

		if trashWorkflow {
			defer os.Remove(workflowFile)
		}

		workflow := storm.NewWorkflow()

		runOpts := []storm.WorkflowRunOptions{
			workflow.WorkflowWithFile(workflowFile),
			workflow.WorkflowWithCallback(func(i any) { fmt.Println(i) }, format),
		}

		if directory != "" {
			runOpts = append(runOpts, workflow.WorkflowWithDirectory(directory))
		}

		if len(contextFlags) > 0 {
			contexts, err := storm.ParseContextFlags(contextFlags, contextFormat)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			runOpts = append(runOpts, workflow.WorkflowWithContexts(contexts))
		}

		err := workflow.Run(runOpts...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)

	agentInstallCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentInstallCmd.Flags().StringP("mode", "m", "prod", "formatio storm installation type (prod or dev)")
	agentCmd.AddCommand(agentInstallCmd)

	agentUninstallCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentCmd.AddCommand(agentUninstallCmd)

	agentRunWorkflowCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")
	agentRunWorkflowCmd.Flags().IntP("format", "f", 1, "available options are; 1 => plain, 2 => struct, 3 => json")
	agentRunWorkflowCmd.Flags().StringArrayP("context", "c", nil, "template context as name:value (repeatable)")
	agentRunWorkflowCmd.Flags().String("context-format", "json", "format of context values: json or base64")
	agentCmd.AddCommand(agentRunWorkflowCmd)

	inventoryCmd.AddCommand(encryptInventoryCmd)
	encryptInventoryCmd.Flags().StringP("inventory", "i", "./inventory.yaml", "formatio storm inventory")

	inventoryCmd.AddCommand(decryptInventoryCmd)
	decryptInventoryCmd.Flags().StringP("encrypted-inventory", "e", "./inventory.yaml.enc", "encrypted inventory file")
	decryptInventoryCmd.Flags().StringP("format", "f", "file", "available options are; plain, file")

	inventoryCmd.PersistentFlags().StringP("encryption-key", "k", "", "encryption key")
	rootCmd.AddCommand(inventoryCmd)

	runWorkflowCmd.Flags().BoolP("trash-workflow", "t", true, "remove workflow file if the workflow is complete")
	runWorkflowCmd.Flags().StringP("directory", "d", ".", "directory to run the workflow from")
	runWorkflowCmd.Flags().IntP("format", "f", 1, "available options are; 1 => plain, 2 => struct, 3 => json")
	runWorkflowCmd.Flags().StringArrayP("context", "c", nil, "template context as name:value (repeatable)")
	runWorkflowCmd.Flags().String("context-format", "json", "format of context values: json or base64")
	rootCmd.AddCommand(runWorkflowCmd)

	rootCmd.AddCommand(agentCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err) // TODO: use logger
		os.Exit(1)
	}
}
