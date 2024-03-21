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

// Package project contains data structures and functions for cmd/gen
package project

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// File represents file of the project that will be generated
type File struct {
	Template   string
	Path       string
	Parameters any
}

// Project represents a set of Files
type Project struct {
	Name, Path, Go, Spire string
	Files                 []*File
}

// Save saves project on the filesystem
func (p *Project) Save() error {
	_ = os.MkdirAll(p.Path, os.ModePerm)

	for _, file := range p.Files {
		fmt.Printf("CREATING: %s\n", file.Path)
		temp, err := template.New("").Parse(file.Template)
		if err != nil {
			return err
		}

		filePath := filepath.Join(p.Path, file.Path)
		fileDir, _ := filepath.Split(filePath)

		_ = os.MkdirAll(fileDir, os.ModePerm)

		sb := new(strings.Builder)

		parameters := file.Parameters

		if parameters == nil {
			parameters = p
		}

		if err = temp.Execute(sb, parameters); err != nil {
			return err
		}

		var content = strings.ReplaceAll(sb.String(), "&lt;", "<")

		_ = os.WriteFile(filePath, []byte(content), os.ModePerm)

		fmt.Printf("✅ %v -- CREATED\n", filePath)
	}

	return nil
}
