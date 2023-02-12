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

// Package domain contains implementation of NSM domain
package domain

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/persistence"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

var current *Domain

// Domain represents environment where is running NSM instance what we want to connect
type Domain struct {
	Name             string
	DNSServerAddress string
	RegistryService  string
	ManagerService   string
	Path             string
	IsDefault        bool
	IsInsecure       bool
}

// SetCurrent replaces the current NSM domain
func SetCurrent(d *Domain) {
	current = d
}

// Current returns current NSM domain
func Current() (*Domain, error) {
	if current != nil {
		return current, nil
	}

	var result, err = persistence.Storage[*Domain]().Select(func(r storage.Resource) bool { return r.(*Domain).IsDefault })

	if err != nil {
		return nil, err
	}

	if len(result) != 1 {
		return nil, errors.New("something went wrong with domains, please use 'nsmctl use domain $DOMAIN_NAME'")
	}

	return result[0].(*Domain), nil
}

func (d *Domain) String() string {
	return "NSM Domain " + d.Name
}

// New creates a new domain
func New(name string) *Domain {
	return &Domain{
		Name:            name,
		ManagerService:  "nsmgr-proxy.nsm-system",
		RegistryService: "registry.nsm-system",
	}
}

// FQDN returns a fully qualified domain name of the service
func (d *Domain) FQDN(service string) string {
	return fmt.Sprintf("%v.%v.", service, d.Name)
}
