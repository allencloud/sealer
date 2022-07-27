// Copyright © 2022 Alibaba Group Holding Ltd.
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

package hostpool

import (
	"fmt"
	"net"
)

type HostPool struct {
	// host is a map:
	// key has a type of string which is from net.Ip.String()
	hosts map[string]*Host
}

// New initializes a brand new HostPool instance.
func New(hostConfigs []*HostConfig) (*HostPool, error) {
	if len(hostConfigs) == 0 {
		return nil, fmt.Errorf("input HostConfigs cannot be empty")
	}
	var hostPool HostPool
	for _, hostConfig := range hostConfigs {
		if _, OK := hostPool.hosts[hostConfig.IP.String()]; OK {
			return nil, fmt.Errorf("there must not be duplicated host IP(%s) in cluster hosts", hostConfig.IP.String())
		}
		hostPool.hosts[hostConfig.IP.String()] = &Host{
			config: HostConfig{
				IP:        hostConfig.IP,
				Port:      hostConfig.Port,
				User:      hostConfig.User,
				Password:  hostConfig.Password,
				Encrypted: hostConfig.Encrypted,
			},
		}
	}
	return &hostPool, nil

}

// Initialize helps HostPool to setup all attributes for each host,
// like scpClient, sshClient and so on.
func (hp *HostPool) Initialize() error {
	for _, host := range hp.hosts {
		if err := host.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize host in HostPool: %v", err)
		}
	}
	return nil
}

func (hp *HostPool) CombinedOutput(inputHost net.IP, cmd string) ([]byte, error) {
	host, exist := hp.hosts[inputHost.String()]
	if !exist {
		return nil, fmt.Errorf("there is no host(%s) in host pool", inputHost.String())
	}
	return host.sshSession.CombinedOutput(cmd)
}

func (hp *HostPool) 
