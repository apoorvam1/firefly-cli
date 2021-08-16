// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"fmt"

	"github.com/hyperledger-labs/firefly-cli/pkg/types"
)

type DependsOn map[string]map[string]string

type HealthCheck struct {
	Test     []string `yaml:"test,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

type LoggingConfig struct {
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

type ServiceDefinition struct {
	ServiceName string
	Service     *Service
	VolumeNames []string
}

type Service struct {
	Image       string                       `yaml:"image,omitempty"`
	Build       string                       `yaml:"build,omitempty"`
	Command     string                       `yaml:"command,omitempty"`
	Environment map[string]string            `yaml:"environment,omitempty"`
	Volumes     []string                     `yaml:"volumes,omitempty"`
	Ports       []string                     `yaml:"ports,omitempty"`
	DependsOn   map[string]map[string]string `yaml:"depends_on,omitempty"`
	HealthCheck *HealthCheck                 `yaml:"healthcheck,omitempty"`
	Logging     *LoggingConfig               `yaml:"logging,omitempty"`
}

type DockerComposeConfig struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]*Service `yaml:"services,omitempty"`
	Volumes  map[string]struct{} `yaml:"volumes,omitempty"`
}

var StandardLogOptions = &LoggingConfig{
	Driver: "json-file",
	Options: map[string]string{
		"max-size": "10m",
		"max-file": "1",
	},
}

func CreateDockerCompose(stack *types.Stack) *DockerComposeConfig {
	compose := &DockerComposeConfig{
		Version:  "2.1",
		Services: make(map[string]*Service),
		Volumes:  make(map[string]struct{}),
	}

	for _, member := range stack.Members {

		if !member.External {
			compose.Services["firefly_core_"+member.ID] = &Service{
				Image: "ghcr.io/hyperledger-labs/firefly:latest",
				Ports: []string{
					fmt.Sprintf("%d:%d", member.ExposedFireflyPort, member.ExposedFireflyPort),
					fmt.Sprintf("%d:%d", member.ExposedFireflyAdminPort, member.ExposedFireflyAdminPort),
				},
				Volumes: []string{fmt.Sprintf("firefly_core_%s:/etc/firefly", member.ID)},
				DependsOn: map[string]map[string]string{
					"dataexchange_" + member.ID: {"condition": "service_started"},
				},
				Logging: StandardLogOptions,
			}

			compose.Volumes["firefly_core_"+member.ID] = struct{}{}
		}

		if stack.Database == "postgres" {
			compose.Services["postgres_"+member.ID] = &Service{
				Image: "postgres",
				Ports: []string{fmt.Sprintf("%d:5432", member.ExposedPostgresPort)},
				Environment: map[string]string{
					"POSTGRES_PASSWORD": "f1refly",
					"PGDATA":            "/var/lib/postgresql/data/pgdata",
				},
				Volumes: []string{fmt.Sprintf("postgres_%s:/var/lib/postgresql/data", member.ID)},
				HealthCheck: &HealthCheck{
					Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
					Interval: "5s",
					Timeout:  "3s",
					Retries:  12,
				},
				Logging: StandardLogOptions,
			}

			compose.Volumes["postgres_"+member.ID] = struct{}{}

			compose.Services["firefly_core_"+member.ID].DependsOn["postgres_"+member.ID] = map[string]string{"condition": "service_healthy"}
		}

		compose.Services["ipfs_"+member.ID] = &Service{
			Image: "ipfs/go-ipfs",
			Ports: []string{
				fmt.Sprintf("%d:5001", member.ExposedIPFSApiPort),
				fmt.Sprintf("%d:8080", member.ExposedIPFSGWPort),
			},
			Environment: map[string]string{
				"IPFS_SWARM_KEY":    stack.SwarmKey,
				"LIBP2P_FORCE_PNET": "1",
			},
			Volumes: []string{
				fmt.Sprintf("ipfs_staging_%s:/export", member.ID),
				fmt.Sprintf("ipfs_data_%s:/data/ipfs", member.ID),
			},
			Logging: StandardLogOptions,
		}

		compose.Volumes["ipfs_staging_"+member.ID] = struct{}{}
		compose.Volumes["ipfs_data_"+member.ID] = struct{}{}

		compose.Services["dataexchange_"+member.ID] = &Service{
			Image:   "ghcr.io/hyperledger-labs/firefly-dataexchange-https:latest",
			Ports:   []string{fmt.Sprintf("%d:3000", member.ExposedDataexchangePort)},
			Volumes: []string{fmt.Sprintf("dataexchange_%s:/data", member.ID)},
			Logging: StandardLogOptions,
		}

		compose.Volumes["dataexchange_"+member.ID] = struct{}{}

	}

	return compose
}