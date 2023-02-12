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

// Package create provides control to create resources
package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

// New creates a new instance of cobra.Command that allows to create resources.
func New(storages map[string]*storage.Storage) *cobra.Command {
	var r = &cobra.Command{
		Use:               "create",
		Aliases:           []string{"apply"},
		Short:             "Creates a new resource",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long: `creates a new resource based on the passed type and file. 
Can create an emptry resouces if passed two arguments (type and name).
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err      error
				filePath string
			)
			filePath, err = cmd.Flags().GetString("from-file")
			if err != nil {
				return err
			}

			if len(args) > 0 {
				var t, n string
				t = args[0]
				if len(args) > 1 {
					n = args[1]
				}
				var s, ok = storages[t]

				if !ok {
					return errors.New("unknown type " + t)
				}

				var result = storages[t].Create(cmd.Context())

				if filePath != "" {
					var b []byte
					// #nosec
					b, err = ioutil.ReadFile(filePath)
					if err != nil {
						return err
					}
					if err = yaml.Unmarshal(b, result); err != nil {
						return err
					}
				}

				if n == "" {
					n = getName(result)
				}

				var err = s.Update(cmd.Context(), n, result)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "created "+n)
				return nil
			}

			return errors.New("")
		},
	}
	r.Flags().StringP("from-file", "f", "", "represents format of out")

	return r
}

func getName(r storage.Resource) string {
	var v = reflect.ValueOf(r)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v.FieldByName("Name").String()
}
