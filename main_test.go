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

//go:build linux
// +build linux

package main_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/edwarnicke/exechelper"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *MainSuite) RequireExec(cmd string, opts ...*exechelper.Option) {
	var initOpts = []*exechelper.Option{exechelper.WithStderr(os.Stderr), exechelper.WithContext(s.ctx), exechelper.WithStdout(os.Stdout)}
	initOpts = append(initOpts, opts...)
	var execErr = exechelper.Run(cmd, initOpts...)
	s.NoError(execErr)
}

func (s *MainSuite) SetupSuite() {
	var cancel func()
	s.ctx, cancel = context.WithTimeout(context.Background(), time.Minute*10)
	s.T().Cleanup(func() { cancel() })

	s.RequireExec("go install")
}

func (s *MainSuite) TestHelp() {
	s.RequireExec("nsmctl --help")
}

func (s *MainSuite) Test_Generate_NetworkServiceEndpoint() {
	var dir = filepath.Join(os.Getenv("GOPATH"), "src", "my_nse_folder")

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	s.RequireExec("nsmctl gen nse --name nse-1 --labels app=my-nse,version=v1.0.0 --path " + dir)

	files, err := ioutil.ReadDir(dir)

	s.Require().NoError(err)

	s.Require().Len(files, 6)

	s.RequireExec("go build ./...", exechelper.WithDir(dir))
}

func (s *MainSuite) Test_Generate_NetworkServiceEndpointVpp() {
	var dir = filepath.Join(os.Getenv("GOPATH"), "src", "my_nse_vpp_folder")

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	s.RequireExec("nsmctl gen nse vpp --name nse-1 --labels app=my-nse,version=v1.0.0 --path " + dir)

	files, err := ioutil.ReadDir(dir)

	s.Require().NoError(err)

	s.Require().Len(files, 6)

	s.RequireExec("go build ./...", exechelper.WithDir(dir))
}

func Test_RunSystemTests(t *testing.T) {
	suite.Run(t, new(MainSuite))
}
