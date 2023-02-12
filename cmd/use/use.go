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

// Package use contains control to current NSM domains
package use

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/domain"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/persistence"
)

// New creates a new cobra.Command instance that allows to manage current NSM domain
func New() *cobra.Command {
	return &cobra.Command{
		Use:               "use",
		Short:             "TODO",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long: `TODO
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("wrong parameters len")
			}
			if args[0] != "domain" {
				return errors.New("unknown type " + args[0])
			}
			var name = args[1]
			var domainStorage = persistence.Storage[*domain.Domain]()

			rawDomain, err := domainStorage.Get(cmd.Context(), name)
			if err != nil {
				return err
			}

			domains, err := domainStorage.List(context.Background())

			if err != nil {
				return err
			}

			for _, d := range domains {
				var castedDomain = d.(*domain.Domain)
				if castedDomain.IsDefault {
					castedDomain.IsDefault = false
				}
				_ = domainStorage.Update(context.Background(), castedDomain.Name, castedDomain)
			}

			var d = rawDomain.(*domain.Domain)

			d.IsDefault = true

			_ = domainStorage.Update(context.Background(), d.Name, d)

			return nil
		},
	}
}
