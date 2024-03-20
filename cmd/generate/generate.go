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

// Package generate contains cobra implementation for cmd/gen
package generate

import (
	_ "embed"

	"os"
	"path/filepath"
	"strings"

	"github.com/edwarnicke/exechelper"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/cmd/generate/endpoint"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/project"
)

//go:embed dockerfile.tmpl
var dockerFileTemplate string

//go:embed imports.go.tmpl
var importsFileTemplate string

var errSpecifyTheTarget = errors.New("specify the target [nse, nse vpp]")

// New creates new cmd/gen instance
func New() *cobra.Command {
	var result *cobra.Command
	var proj = new(project.Project)

	result = &cobra.Command{
		Use:               "generate",
		Short:             "gen",
		Aliases:           []string{"gen"},
		DisableAutoGenTag: true,
		TraverseChildren:  true,
		Long:              `generates something`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errSpecifyTheTarget
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			var opts []*exechelper.Option
			var err error

			opts = append(opts, exechelper.WithStdout(cmd.OutOrStdout()), exechelper.WithStderr(cmd.ErrOrStderr()))
			if proj.Path != "" {
				opts = append(opts, exechelper.WithDir(proj.Path))
			}

			if err = proj.Save(); err != nil {
				return err
			}

			_ = exechelper.Run("go mod init", opts...)
			if err = exechelper.Run("go mod tidy", opts...); err != nil {
				return err
			}
			return nil
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := exechelper.Run("docker -v", exechelper.WithStdout(os.Stdout)); err != nil {
				return errors.Wrap(err, "docker is required")
			}
			var goVersionStream strings.Builder

			if err := exechelper.Run("go version", exechelper.WithStdout(&goVersionStream)); err != nil {
				return errors.Wrap(err, "go is required")
			}

			_, _ = os.Stdout.WriteString(goVersionStream.String())

			proj.Path, _ = cmd.Flags().GetString("path")
			proj.Name, _ = cmd.Flags().GetString("name")
			proj.Spire, _ = cmd.Flags().GetString("spire")
			proj.Go, _ = cmd.Flags().GetString("go")

			if !strings.Contains(goVersionStream.String(), proj.Go) {
				return errors.New("missed go with version " + proj.Go)
			}

			proj.Files = append(proj.Files,
				&project.File{
					Path:     "Dockerfile",
					Template: dockerFileTemplate,
				},
				&project.File{
					Path:     filepath.Join("internal", "pkg", "imports", "imports.go"),
					Template: importsFileTemplate,
				},
			)
			return nil
		},
	}

	result.AddCommand(endpoint.New(proj))

	addFlags(result)
	inheritPersistentBehaviour(result, result.Parent())

	return result
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "path to the project")
	cmd.Flags().StringP("name", "n", "app", "name of the generating app")
	cmd.Flags().StringP("spire", "s", "1.9.1", "version of spire")
	cmd.Flags().StringP("go", "g", "1.21", "version of go")

	for _, child := range cmd.Commands() {
		addFlags(child)
	}
}

func inheritPersistentBehaviour(cmd, parent *cobra.Command) {
	for _, child := range cmd.Commands() {
		if parent != nil {
			if parent.PersistentPreRunE != nil && cmd.PersistentPreRunE != nil {
				var persistentPreRunE = cmd.PersistentPreRunE
				cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
					if err := parent.PersistentPreRunE(cmd, args); err != nil {
						return nil
					}
					return persistentPreRunE(cmd, args)
				}
			}

			if parent.PersistentPostRunE != nil && cmd.PersistentPostRunE != nil {
				var persistentPostRunE = cmd.PersistentPostRunE
				cmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
					if err := parent.PersistentPostRunE(cmd, args); err != nil {
						return nil
					}
					return persistentPostRunE(cmd, args)
				}
			}
		}

		inheritPersistentBehaviour(child, cmd)
	}
}
