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

package vpp

import (
	_ "embed"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/project"
	"github.com/spf13/cobra"
)

//go:embed main.go.tmpl
var mainFileTemplate string

//go:embed dockerfile.vpp.tmpl
var dockerFileTemplate string

func New(proj *project.Project) *cobra.Command {
	var result = &cobra.Command{
		Use:               "vpp",
		Short:             "generates a vpp nse",
		DisableAutoGenTag: true,
		Long:              `generates network service mesh endpoint based on Vector Packet Proccessing platform. See more details https://wiki.fd.io/view/VPP/What_is_VPP%3F`,

		RunE: func(cmd *cobra.Command, args []string) error {
			var labels, _ = cmd.Flags().GetStringToString("labels")
			var services, _ = cmd.Flags().GetStringArray("services")
			var vpp, _ = cmd.Flags().GetString("vpp")

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
			proj.Files = append(proj.Files, &project.File{
				Path:     "Dockerfile",
				Template: dockerFileTemplate,
				Parameters: struct {
					project.Project
					VPP string
				}{
					VPP: vpp,
				},
			})

			return nil
		},
	}

	result.Flags().StringP("vpp", "", "v22.06-rc0-147-gb2b1a4ad2", "version of vpp")

	return result
}
