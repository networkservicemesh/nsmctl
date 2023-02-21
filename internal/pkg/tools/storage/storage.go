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

// Package storage contains abstractions for data layer to work with different sources
package storage

import (
	"fmt"

	"golang.org/x/net/context"
)

// Resource represents NSM resource
type Resource interface {
	fmt.Stringer
}

// Storage is abstraction on data layer
type Storage struct {
	Get    func(context.Context, string) (Resource, error)
	Delete func(context.Context, string) error
	Update func(context.Context, string, Resource) error
	List   func(context.Context) ([]Resource, error)
	Create func(context.Context) Resource
}

// Select selects resources by criteria
func (si *Storage) Select(selector func(Resource) bool) ([]Resource, error) {
	var list, err = si.List(context.Background())
	if err != nil {
		return nil, err
	}
	var result []Resource

	for _, item := range list {
		if selector(item) {
			result = append(result, item)
		}
	}

	return result, nil
}
