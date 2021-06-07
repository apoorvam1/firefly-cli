/*
Copyright © 2021 Kaleido, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/kaleido-io/firefly-cli/internal/stacks"
	"github.com/nguyer/promptui"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset <stack_name>",
	Short: "Clear all data in a stack",
	Long: `Clear all data in a stack

This command clears all data in a stack, but leaves the stack itself.
This is useful for testing when you want to start with a clean slate
but don't want to actually recreate the resources in the stack itself.
The stack must be stopped to run this command.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no stack specified")
		}
		stackName := args[0]

		if exists, err := stacks.CheckExists(stackName); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("stack '%s' does not exist", stackName)
		}

		if !force {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("reset all data in FireFly stack '%s'", stackName),
				IsConfirm: true,
			}

			fmt.Println("WARNING: This will completely remove your all transactions and data from your FireFly stack. Are you sure you want to do that?")
			result, err := prompt.Run()

			if err != nil || strings.ToLower(result) != "y" {
				fmt.Printf("canceled")
				return nil
			}
		}

		if stack, err := stacks.LoadStack(stackName); err != nil {
			return err
		} else {
			fmt.Printf("resetting FireFly stack '%s'... ", stackName)
			stack.ResetStack(verbose)
			fmt.Println("done")
		}

		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&force, "force", "f", false, "Reset the stack without prompting for confirmation")
	rootCmd.AddCommand(resetCmd)
}
