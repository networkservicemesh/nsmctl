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

// Package delete provides control to delete resources
package delete

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

// New creates a new  *cobra.Command that allows to delete NSM resources
func New(storages map[string]*storage.Storage) *cobra.Command {
	return &cobra.Command{
		Use:               "delete",
		Short:             "TODO",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long: `TODO
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("resource type is required")
			}

			var resourceType = args[0]

			var s, ok = storages[resourceType]

			if !ok {
				return errors.New("unknown type " + resourceType)
			}

			for _, item := range args[1:] {
				if err := s.Delete(cmd.Context(), item); err != nil {
					return err
				}
				fmt.Println("removed " + resourceType + " " + item)
			}
			return nil
		},
	}
}
