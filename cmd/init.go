// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/stacks"
)

var initOptions stacks.InitOptions
var databaseSelection string
var blockchainProviderSelection string
var tokensProviderSelection string

var initCmd = &cobra.Command{
	Use:   "init [stack_name] [member_count]",
	Short: "Create a new FireFly local dev stack",
	Long:  `Create a new FireFly local dev stack`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var stackName string
		stackManager := stacks.NewStackManager(logger)

		if err := validateDatabaseProvider(databaseSelection); err != nil {
			return err
		}
		if err := validateBlockchainProvider(blockchainProviderSelection); err != nil {
			return err
		}
		if err := validateTokensProvider(tokensProviderSelection); err != nil {
			return err
		}

		fmt.Println("initializing new FireFly stack...")

		if len(args) > 0 {
			stackName = args[0]
			err := validateName(stackName)
			if err != nil {
				return err
			}
		} else {
			stackName, _ = prompt("stack name: ", validateName)
			fmt.Println("You selected " + stackName)
		}

		var memberCountInput string
		if len(args) > 1 {
			memberCountInput = args[1]
			if err := validateCount(memberCountInput); err != nil {
				return err
			}
		} else {
			memberCountInput, _ = prompt("number of members: ", validateCount)
		}
		memberCount, _ := strconv.Atoi(memberCountInput)

		initOptions.Verbose = verbose
		initOptions.DatabaseSelection, _ = stacks.DatabaseSelectionFromString(databaseSelection)
		initOptions.TokensProvider, _ = stacks.TokensProviderFromString(tokensProviderSelection)

		if err := stackManager.InitStack(stackName, memberCount, &initOptions); err != nil {
			return err
		}

		fmt.Printf("Stack '%s' created!\nTo start your new stack run:\n\n%s start %s\n", stackName, rootCmd.Use, stackName)
		fmt.Printf("\nYour docker compose file for this stack can be found at: %s\n\n", filepath.Join(constants.StacksDir, stackName, "docker-compose.yml"))
		return nil
	},
}

func validateName(stackName string) error {
	if strings.TrimSpace(stackName) == "" {
		return errors.New("stack name must not be empty")
	}
	if exists, err := stacks.CheckExists(stackName); exists {
		return fmt.Errorf("stack '%s' already exists", stackName)
	} else {
		return err
	}
}

func validateCount(input string) error {
	if i, err := strconv.Atoi(input); err != nil {
		return errors.New("invalid number")
	} else if i <= 0 {
		return errors.New("number of members must be greater than zero")
	} else if initOptions.ExternalProcesses >= i {
		return errors.New("number of external processes should not be equal to or greater than the number of members in the network - at least one FireFly core container must exist to be able to extrat and deploy smart contracts")
	}
	return nil
}

func validateDatabaseProvider(input string) error {
	_, err := stacks.DatabaseSelectionFromString(input)
	if err != nil {
		return err
	}
	return nil
}

func validateBlockchainProvider(input string) error {
	blockchainSelection, err := stacks.BlockchainProviderFromString(input)
	if err != nil {
		return err
	}

	if blockchainSelection != stacks.GoEthereum {
		return errors.New("geth is currently the only supported blockchain provider - support for other providers is coming soon")
	}
	return nil
}

func validateTokensProvider(input string) error {
	_, err := stacks.TokensProviderFromString(input)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	initCmd.Flags().IntVarP(&initOptions.FireFlyBasePort, "firefly-base-port", "p", 5000, "Mapped port base of FireFly core API (1 added for each member)")
	initCmd.Flags().IntVarP(&initOptions.ServicesBasePort, "services-base-port", "s", 5100, "Mapped port base of services (100 added for each member)")
	initCmd.Flags().StringVarP(&databaseSelection, "database", "d", "sqlite3", fmt.Sprintf("Database type to use. Options are: %v", stacks.DBSelectionStrings))
	initCmd.Flags().StringVarP(&blockchainProviderSelection, "blockchain-provider", "", "geth", fmt.Sprintf("Blockchain provider to use. Options are: %v", stacks.BlockchainProviderStrings))
	initCmd.Flags().StringVarP(&tokensProviderSelection, "tokens-provider", "", "erc1155", fmt.Sprintf("Tokens provider to use. Options are: %v", stacks.TokensProviderStrings))
	initCmd.Flags().IntVarP(&initOptions.ExternalProcesses, "external", "e", 0, "Manage a number of FireFly core processes outside of the docker-compose stack - useful for development and debugging")

	rootCmd.AddCommand(initCmd)
}
