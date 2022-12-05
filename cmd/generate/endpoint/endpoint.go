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

// Package endpoint contains a command for generating nsm endpoints
package endpoint

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/cmd/generate/endpoint/vpp"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/project"
)

//go:embed main.go.tmpl
var mainFileTemplate string

//go:embed deployment.yaml.tmpl
var deploymentFileTemplate string

// New creates a new cobra.Command instance for cmd/gen/nse.
func New(proj *project.Project) *cobra.Command {
	var result = &cobra.Command{
		Use:               "endpoint",
		Short:             "generates nse",
		Aliases:           []string{"nse"},
		DisableAutoGenTag: true,
		Long:              `generates network service mesh endpoint. See more details https://networkservicemesh.io/docs/concepts/architecture/#endpoints`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			proj.Files = append(proj.Files, &project.File{
				Path:     "deployment.yaml",
				Template: deploymentFileTemplate,
			})
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var labels, _ = cmd.Flags().GetStringToString("labels")
			var services, _ = cmd.Flags().GetStringArray("services")

			proj.Files = append(proj.Files, &project.File{
				Path:     "main.go",
				Template: mainFileTemplate,
				Parameters: struct {
					Name     string
					Labels   map[string]string
					Services []string
				}{
					Name:     proj.Name,
					Labels:   labels,
					Services: services,
				},
			})

			return nil
		},
	}

	addFlags(result)

	result.AddCommand(vpp.New(proj))

	return result
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("services", "", []string{"my-networkservice"}, "list of network servcies")
	cmd.Flags().StringToStringP("labels", "l", nil, "name of the generating app")

	for _, child := range cmd.Commands() {
		addFlags(child)
	}
}
