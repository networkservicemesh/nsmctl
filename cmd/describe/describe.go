// Copyright (c) 2023 Cisco and/or its affiliates.
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

// Package describe provides control to read resources
package describe

import (
	"errors"
	"io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

// Printer prints resources
type Printer interface {
	Print([]any)
}

type yamlPrinter struct {
	out io.Writer
}

func (d *yamlPrinter) Print(itens []any) {
	for _, item := range itens {
		var b, _ = yaml.Marshal(item)
		_, _ = d.out.Write(b)
		_, _ = d.out.Write([]byte("\n"))
	}
}

// New creates a new instance of cobra.Command that allows to describe resources
func New(storages map[string]*storage.Storage) *cobra.Command {
	var r = &cobra.Command{
		Use:               "describe",
		Short:             "describes NSM resouces",
		Aliases:           []string{"describe"},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long: `Describes NSM resources from the current NSM Domain. 
If no name passed describes list of the resources instead.
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var p Printer = &yamlPrinter{out: cmd.OutOrStdout()}

			var items []interface{}

			if len(args) == 0 {
				return errors.New("resource type is required")
			}

			var resourceType = args[0]

			var s, ok = storages[resourceType]

			if !ok {
				return errors.New("unknown type " + resourceType)
			}

			if len(args) == 1 {
				var list, _ = s.List(cmd.Context())

				for _, item := range list {
					items = append(items, item)
				}
			}

			if len(args) > 1 {
				for _, item := range args[1:] {
					v, getErr := s.Get(cmd.Context(), item)
					if getErr != nil {
						return getErr
					}
					items = append(items, v)
				}
			}

			p.Print(items)

			return nil
		},
	}
	return r
}
