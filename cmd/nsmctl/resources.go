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

package nsmctl

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edwarnicke/grpcfd"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/domain"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/persistence"
	"github.com/networkservicemesh/nsmctl/internal/pkg/tools/storage"
	"github.com/networkservicemesh/sdk/pkg/registry/common/grpcmetadata"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/spiffejwt"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

func defaultResources() map[string]*storage.Storage {
	var result = make(map[string]*storage.Storage)

	registerAliases(result, persistence.Storage[*domain.Domain](), "domain", "domains")
	registerAliases(result, newConnectionsStorage(), "conn", "conns", "connection", "connections")
	registerAliases(result, newNSStorage(), "networkservice", "networkservices", "netsvc", "netsvcs")
	registerAliases(result, newNSEStorage(), "networkserviceendpoints", "endpoints", "networkserviceendpoint", "endpoint", "nse", "nses")

	return result
}

func registerAliases(m map[string]*storage.Storage, v *storage.Storage, aliases ...string) {
	for _, k := range aliases {
		m[k] = v
	}
}

func newConnectionsStorage() *storage.Storage {
	return &storage.Storage{
		Get: func(ctx context.Context, s string) (storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.ManagerService)
			if err != nil {
				return nil, err
			}
			monitorCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			var monitorConnectionClient = networkservice.NewMonitorConnectionClient(cc)
			stream, err := monitorConnectionClient.MonitorConnections(monitorCtx, &networkservice.MonitorScopeSelector{PathSegments: []*networkservice.PathSegment{
				{
					Id: s,
				},
			}})
			if err != nil {
				return nil, err
			}
			resp, err := stream.Recv()
			if err != nil {
				return nil, err
			}
			if v, ok := resp.Connections[s]; ok {
				return v, nil
			}
			return nil, errors.New("connection with id " + s + " is not found")
		},
		Delete: func(ctx context.Context, s string) error {
			return errors.New("connections are readonly")
		},
		Create: func(ctx context.Context) storage.Resource {
			return new(registry.NetworkServiceEndpoint)
		},
		List: func(ctx context.Context) ([]storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.ManagerService)
			if err != nil {
				return nil, err
			}
			monitorCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			var monitorConnectionClient = networkservice.NewMonitorConnectionClient(cc)
			stream, err := monitorConnectionClient.MonitorConnections(monitorCtx, &networkservice.MonitorScopeSelector{PathSegments: []*networkservice.PathSegment{
				{},
			}})
			if err != nil {
				return nil, err
			}
			resp, err := stream.Recv()
			if err != nil {
				return nil, err
			}

			var result []storage.Resource

			for _, item := range resp.Connections {
				result = append(result, item)
			}

			return result, nil
		},
	}
}

func newNSStorage() *storage.Storage {
	return &storage.Storage{
		Get: func(ctx context.Context, s string) (storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return nil, err
			}
			var nsClient = registry.NewNetworkServiceRegistryClient(cc)

			stream, err := nsClient.Find(ctx, &registry.NetworkServiceQuery{
				NetworkService: &registry.NetworkService{
					Name: s,
				}})
			if err != nil {
				return nil, err
			}
			var list = registry.ReadNetworkServiceList(stream)

			if len(list) == 0 {
				return nil, errors.New(s + " is not found")
			}
			return list[0], nil
		},
		Delete: func(ctx context.Context, s string) error {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return err
			}
			var nsClient = next.NewNetworkServiceRegistryClient(
				grpcmetadata.NewNetworkServiceRegistryClient(),
				registry.NewNetworkServiceRegistryClient(cc),
			)

			_, err = nsClient.Unregister(ctx, &registry.NetworkService{Name: s})
			return err
		},
		Create: func(ctx context.Context) storage.Resource {
			return new(registry.NetworkService)
		},
		Update: func(ctx context.Context, s string, r storage.Resource) error {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return err
			}
			var nsClient = next.NewNetworkServiceRegistryClient(
				grpcmetadata.NewNetworkServiceRegistryClient(),
				registry.NewNetworkServiceRegistryClient(cc),
			)

			_, err = nsClient.Register(ctx, r.(*registry.NetworkService))

			return err
		},
		List: func(ctx context.Context) ([]storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return nil, err
			}
			var nseClient = registry.NewNetworkServiceRegistryClient(cc)

			stream, err := nseClient.Find(ctx,
				&registry.NetworkServiceQuery{
					NetworkService: &registry.NetworkService{},
				},
			)
			if err != nil {
				return nil, err
			}
			var list = registry.ReadNetworkServiceList(stream)
			var result []storage.Resource

			for _, item := range list {
				result = append(result, item)
			}

			return result, nil
		},
	}
}

func newNSEStorage() *storage.Storage {
	return &storage.Storage{
		Get: func(ctx context.Context, s string) (storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return nil, err
			}
			var nseClient = next.NewNetworkServiceEndpointRegistryClient(
				grpcmetadata.NewNetworkServiceEndpointRegistryClient(),
				registry.NewNetworkServiceEndpointRegistryClient(cc),
			)

			stream, err := nseClient.Find(ctx, &registry.NetworkServiceEndpointQuery{
				NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
					Name: s,
				}})
			if err != nil {
				return nil, err
			}
			var list = registry.ReadNetworkServiceEndpointList(stream)

			if len(list) == 0 {
				return nil, errors.New(s + " is not found")
			}
			return list[0], nil
		},
		Delete: func(ctx context.Context, s string) error {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return err
			}
			var nseClient = next.NewNetworkServiceEndpointRegistryClient(
				grpcmetadata.NewNetworkServiceEndpointRegistryClient(),
				registry.NewNetworkServiceEndpointRegistryClient(cc),
			)
			_, err = nseClient.Unregister(ctx, &registry.NetworkServiceEndpoint{Name: s})
			return err
		},
		Create: func(ctx context.Context) storage.Resource {
			return new(registry.NetworkServiceEndpoint)
		},
		List: func(ctx context.Context) ([]storage.Resource, error) {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return nil, err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return nil, err
			}
			var nseClient = registry.NewNetworkServiceEndpointRegistryClient(cc)

			stream, err := nseClient.Find(ctx,
				&registry.NetworkServiceEndpointQuery{
					NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{},
				},
			)
			if err != nil {
				return nil, err
			}
			var list = registry.ReadNetworkServiceEndpointList(stream)
			var result []storage.Resource

			for _, item := range list {
				result = append(result, item)
			}

			return result, nil
		},
		Update: func(ctx context.Context, s string, r storage.Resource) error {
			var cc grpc.ClientConnInterface
			var d, err = domain.Current()
			if err != nil {
				return err
			}
			cc, err = dial(ctx, d, d.RegistryService)
			if err != nil {
				return err
			}
			var nseClient = next.NewNetworkServiceEndpointRegistryClient(
				grpcmetadata.NewNetworkServiceEndpointRegistryClient(),
				registry.NewNetworkServiceEndpointRegistryClient(cc),
			)

			_, err = nseClient.Register(ctx, r.(*registry.NetworkServiceEndpoint))

			return err
		},
	}
}

func dial(ctx context.Context, d *domain.Domain, target string) (grpc.ClientConnInterface, error) {
	if !strings.Contains(target, ":") {
		var dialer net.Dialer
		var r = net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				if d.DNSServerAddress != "" {
					return dialer.DialContext(ctx, network, d.DNSServerAddress)
				}
				return dialer.DialContext(ctx, network, address)
			},
		}
		serviceDomain := d.FQDN(target)

		_, records, err := r.LookupSRV(ctx, "", "", serviceDomain)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("resolver.LookupSERV return empty result")
		}
		port := strconv.Itoa(int(records[0].Port))

		ips, err := r.LookupIPAddr(ctx, serviceDomain)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("resolver.LookupIPAddr return empty result")
		}
		ipAddr := ips[0].IP

		target = fmt.Sprintf("%v:%v", ipAddr.String(), port)
	}

	if os.Getenv(workloadapi.SocketEnv) == "" {
		_ = os.Setenv(workloadapi.SocketEnv, "unix:///tmp/spire-agent/public/api.sock")
	}

	var dialOptions []grpc.DialOption

	if d.IsInsecure {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		source, err := workloadapi.NewX509Source(ctx)
		if err != nil {
			return nil, err
		}

		tlsClientConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())
		tlsClientConfig.MinVersion = tls.VersionTLS12

		dialOptions = append(dialOptions,
			grpc.WithTransportCredentials(
				grpcfd.TransportCredentials(credentials.NewTLS(tlsClientConfig))),
			grpc.WithDefaultCallOptions(
				grpc.PerRPCCredentials(token.NewPerRPCCredentials(spiffejwt.TokenGeneratorFunc(source, time.Hour))),
			),
			grpcfd.WithChainStreamInterceptor(),
			grpcfd.WithChainUnaryInterceptor(),
		)
	}

	dialOptions = append([]grpc.DialOption{
		grpc.WithBlock(),
	}, dialOptions...)

	return grpc.DialContext(ctx, target, dialOptions...)
}
