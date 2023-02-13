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
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/edwarnicke/exechelper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/domain"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/persistence"
	"github.com/networkservicemesh/sdk/pkg/tools/sandbox"
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

	files, err := os.ReadDir(dir)

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

	files, err := os.ReadDir(dir)

	s.Require().NoError(err)

	s.Require().Len(files, 6)

	s.RequireExec("go build ./...", exechelper.WithDir(dir))
}

func (s *MainSuite) Test_SandboxAndNSMControl() {
	var ctx, cancel = context.WithCancel(s.ctx)
	defer cancel()
	var d = sandbox.NewBuilder(ctx, s.T()).SetNodesCount(1).Build()

	defer func() {
		_ = persistence.Delete[*domain.Domain]("test")
	}()

	_ = persistence.Store("test", &domain.Domain{
		Name:            "test",
		ManagerService:  net.JoinHostPort(d.Nodes[0].NSMgr.URL.Hostname(), d.Nodes[0].NSMgr.URL.Port()),
		RegistryService: net.JoinHostPort(d.Registry.URL.Hostname(), d.Registry.URL.Port()),
		IsInsecure:      true,
	})

	nseReg := &registry.NetworkServiceEndpoint{
		Name:                "final-endpoint",
		NetworkServiceNames: []string{"ns"},
	}

	nsRegistryClient := d.NewNSRegistryClient(ctx, sandbox.GenerateTestToken)

	_, err := nsRegistryClient.Register(ctx, &registry.NetworkService{Name: "ns"})
	require.NoError(s.T(), err)

	_ = d.Nodes[0].NewEndpoint(ctx, nseReg, sandbox.GenerateTestToken)

	nsc := d.Nodes[0].NewClient(ctx, sandbox.GenerateTestToken)

	request := &networkservice.NetworkServiceRequest{
		MechanismPreferences: []*networkservice.Mechanism{
			{Cls: cls.LOCAL, Type: "kernel"},
		},
		Connection: &networkservice.Connection{
			Id:             "1",
			NetworkService: "ns",
			Labels:         make(map[string]string),
		},
	}

	_, err = nsc.Request(ctx, request)
	require.NoError(s.T(), err)

	s.RequireExec("nsmctl get domains --domain test")
	s.RequireExec("nsmctl get nses --domain test")
	s.RequireExec("nsmctl get netsvc --domain test")
	s.RequireExec("nsmctl get connections --domain test")

	s.RequireExec("nsmctl describe -o yaml domains")
	s.RequireExec("nsmctl describe -o yaml nses --domain test")
	s.RequireExec("nsmctl describe -o yaml netsvc --domain test")
	s.RequireExec("nsmctl describe -o yaml connections --domain test")

	var p = filepath.Join(s.T().TempDir(), "mse.yaml")
	_ = os.WriteFile(p, []byte("name: my-nse"), os.ModePerm)
	s.RequireExec("nsmctl apply nse --domain test -f " + p)
	s.RequireExec("nsmctl get nse --domain test my-nse")
	s.RequireExec("nsmctl delete nse --domain test my-nse")

	p = filepath.Join(s.T().TempDir(), "ns.yaml")
	_ = os.WriteFile(p, []byte("name: my-ns"), os.ModePerm)
	s.RequireExec("nsmctl apply netsvc --domain test -f " + p)
	s.RequireExec("nsmctl get netsvc --domain test my-ns")
	s.RequireExec("nsmctl delete netsvc --domain test my-ns")
}

func Test_RunSystemTests(t *testing.T) {
	suite.Run(t, new(MainSuite))
}
