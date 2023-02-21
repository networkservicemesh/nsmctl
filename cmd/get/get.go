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

// Package get provides control to read resources
package get

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

const maxTabPrinterLen = 15

// Printer prints resources
type Printer interface {
	Print([]any)
}

type tabPrinter struct {
	out io.Writer
}

func (p *tabPrinter) Print(list []any) {
	w := tabwriter.NewWriter(p.out, 0, 0, 3, ' ', tabwriter.TabIndent)

	for j, item := range list {
		var outStr strings.Builder
		v := reflect.ValueOf(item)

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if j == 0 {
			for i := 0; i < v.NumField(); i++ {
				if !v.Type().Field(i).IsExported() {
					continue
				}

				var name = toSnakeCase(fmt.Sprint(v.Type().Field(i).Name))
				_, _ = outStr.WriteString(strings.ToUpper(name))

				if i+1 < v.NumField() {
					outStr.Write([]byte("\t"))
				}
			}
			fmt.Fprintln(w, outStr.String())
			outStr.Reset()
		}

		for i := 0; i < v.NumField(); i++ {
			if !v.Type().Field(i).IsExported() {
				continue
			}

			var value = fmt.Sprint(v.Field(i).Interface())
			if isComplex(v.Field(i).Kind()) {
				if len(value) > maxTabPrinterLen {
					value = value[:maxTabPrinterLen-3] + "..."
				}
			}
			outStr.WriteString(value)
			if i+1 < v.NumField() {
				outStr.Write([]byte("\t"))
			}
		}
		fmt.Fprintln(w, outStr.String())
	}
	_ = w.Flush()
}

func isComplex(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.String, reflect.Int32:
		return false
	default:
		return true
	}
}

// New creates a new instance of cobra.Command that allows to get/describe resources.
func New(storages map[string]*storage.Storage) *cobra.Command {
	var r = &cobra.Command{
		Use:               "get",
		Short:             "Get NSM resource/resouces",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long: `Gets NSM resources from the current NSM Domain. 
If no name passed gets list of the resources instead.
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var goTemplate string
			var err error
			var templ *template.Template

			goTemplate, err = cmd.Flags().GetString("go-template")
			if err != nil {
				return err
			}

			var p Printer = &tabPrinter{out: cmd.OutOrStdout()}

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

			if goTemplate != "" {
				templ, err = template.New("get/gotemplate").Parse(goTemplate)
				if err != nil {
					return err
				}
				var templArgs interface{} = items

				if len(args) == 2 {
					templArgs = items[0]
				}
				return templ.Execute(cmd.OutOrStdout(), templArgs)
			}

			p.Print(items)

			return nil
		},
	}
	r.Flags().StringP("go-template", "", "", "epects 'go-tempalte' ")
	return r
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
