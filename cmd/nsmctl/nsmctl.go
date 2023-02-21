// Copyright (c) 2022-2023 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package nsmctl contains root command of nsmctl
package nsmctl

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/cmd/create"
	"github.com/networkservicemesh/nsmctl/cmd/delete"
	"github.com/networkservicemesh/nsmctl/cmd/describe"
	"github.com/networkservicemesh/nsmctl/cmd/generate"
	"github.com/networkservicemesh/nsmctl/cmd/get"
	"github.com/networkservicemesh/nsmctl/cmd/use"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/domain"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/persistence"
)

// New creates new cmd/nsmctl
func New() *cobra.Command {
	nsmctlCmd := &cobra.Command{
		Use:               "nsmctl",
		Short:             "NSM command line tool",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Version:           "0.0.1",
		Long: `Network Service Mesh Command Line Tool
	`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Add flag --help to get commands")
			fmt.Println("See more information about NSM https://networkservicemesh.io/docs/concepts/enterprise_users/")
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var domainName, err = cmd.Flags().GetString("domain")
			if err != nil {
				return err
			}

			if domainName != "" {
				v, vErr := persistence.Load[*domain.Domain](domainName)
				if vErr != nil {
					return vErr
				}
				domain.SetCurrent(v)
			}

			return nil
		},
	}

	var storages = defaultResources()

	nsmctlCmd.AddCommand(get.New(storages))
	nsmctlCmd.AddCommand(create.New(storages))
	nsmctlCmd.AddCommand(delete.New(storages))
	nsmctlCmd.AddCommand(describe.New(storages))
	nsmctlCmd.AddCommand(use.New())
	nsmctlCmd.AddCommand(generate.New())

	addCommonFlags(nsmctlCmd)

	return nsmctlCmd
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("domain", "d", "", "nsm domain that should be used for control")

	for _, child := range cmd.Commands() {
		addCommonFlags(child)
	}
}
