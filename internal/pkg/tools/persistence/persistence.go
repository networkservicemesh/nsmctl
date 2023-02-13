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

// Package persistence allows to stora custom data for nsmctl in user cache
package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"

	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
)

// PathOf finds path for the resource in the nsmctl cache
func PathOf[T any](key string) string {
	var zero = new(T)
	if d, err := os.UserCacheDir(); err != nil {
		panic(err.Error())
	} else {
		var t = fmt.Sprintf("%T", zero)
		var pieces = strings.Split(t, ".")
		t = pieces[len(pieces)-1]
		t = strings.ToLower(t)
		return filepath.Join(d, "nsmctl", t, key)
	}
}

// Delete deletes resource from the nsmctl cache
func Delete[T any](key string) error {
	var filePath = PathOf[T](key)
	return os.Remove(filePath)
}

// Store stores resource in the nsmctl cache
func Store[T any](key string, value T) error {
	var b []byte
	var err error

	b, err = yaml.Marshal(value)
	if err != nil {
		return err
	}
	var filePath = PathOf[T](key)

	if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filePath, b, os.ModePerm)
}

// Load loads serializable resource from the nsmctl cache
func Load[T any](key string) (T, error) {
	var result T
	var err error
	var filePath = PathOf[T](key)

	// #nosec
	if _, err = os.Stat(filePath); err != nil {
		return result, err
	}

	var b []byte
	// #nosec
	if b, err = os.ReadFile(filePath); err != nil {
		return result, err
	}

	err = yaml.Unmarshal(b, &result)

	return result, err
}

// Storage creates a storage abstraction for serializable resource
func Storage[T storage.Resource]() *storage.Storage {
	return &storage.Storage{
		Get: func(ctx context.Context, name string) (storage.Resource, error) {
			return Load[T](name)
		},
		Delete: func(ctx context.Context, name string) error {
			return Delete[T](name)
		},
		Update: func(ctx context.Context, s string, r storage.Resource) error {
			return Store(s, r.(T))
		},
		List: func(ctx context.Context) ([]storage.Resource, error) {
			var files, err = os.ReadDir(PathOf[T](""))
			if err != nil {
				return nil, err
			}
			var result []storage.Resource
			for _, file := range files {
				var resource, err = Load[T](file.Name())
				if err != nil {
					return nil, err
				}
				result = append(result, resource)
			}

			return result, nil
		},
		Create: func(ctx context.Context) storage.Resource {
			var zero T
			var t = reflect.TypeOf(zero)
			if reflect.ValueOf(zero).Kind() == reflect.Ptr {
				t = t.Elem()
			}
			var res = reflect.New(t).Interface().(storage.Resource)
			return res
		},
	}
}
