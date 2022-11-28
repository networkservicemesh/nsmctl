// Copyright (c) 2022 Cisco and/or its affiliates.
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

package nsmctl

import (
	"fmt"

	"github.com/networkservicemesh/nsmctl/cmd/generate"
	"github.com/spf13/cobra"
)

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
	}

	nsmctlCmd.AddCommand(generate.New())
	return nsmctlCmd
}
