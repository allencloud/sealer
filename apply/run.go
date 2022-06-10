// Copyright © 2021 Alibaba Group Holding Ltd.
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

package apply

import (
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/sealerio/sealer/types/api/v1"

	"github.com/sealerio/sealer/apply/applydriver"

	"github.com/sealerio/sealer/common"
	v2 "github.com/sealerio/sealer/types/api/v2"
	"github.com/sealerio/sealer/utils/net"
)

type ClusterArgs struct {
	cluster   *v2.Cluster
	imageName string
	runArgs   *Args
	hosts     []v2.Host
}

func PreProcessIPList(joinArgs *Args) error {
	if err := net.AssemblyIPList(&joinArgs.Masters); err != nil {
		return err
	}
	if err := net.AssemblyIPList(&joinArgs.Nodes); err != nil {
		return err
	}
	return nil
}

func ConstructClusterFromArg(runArgs *Args) *v2.Cluster {
	hosts := []v2.Host{}
	if net.IsIPList(runArgs.Masters) && (net.IsIPList(runArgs.Nodes) || runArgs.Nodes == "") {
		masters := strings.Split(runArgs.Masters, ",")
		nodes := strings.Split(runArgs.Nodes, ",")
		c.hosts = []v2.Host{}
		hosts := getHostsWithIpsPort(masters, runArgs.Port, common.MASTER)
		if len(nodes) != 0 {
			c.setHostWithIpsPort(nodes, runArgs.Port, common.NODE)
		}
		hosts = c.hosts
	} else if ip, err := net.GetLocalDefaultIP(); err != nil {
		return err
	} else {
		hosts = []v2.Host{
			{
				IPS:   []string{ip},
				Roles: []string{common.MASTER},
			},
		}
	}

	cluster := v2.Cluster{
		APIVersion: common.APIVersion,
		Kind:       common.Cluster,
		Name:       runArgs.ClusterName,
		Spec: v2.ClusterSpec{
			SSH: v1.SSH{
				User:     runArgs.User,
				PkPasswd: runArgs.PkPassword,
				Pk:       runArgs.Pk,
				Port:     strconv.Itoa(int(runArgs.Port)),
			},
			Hosts:   hosts,
			Env:     runArgs.CustomEnv,
			CMDArgs: runArgs.CMDArgs,
		},
	}

	if runArgs.Password != "" {
		cluster.Spec.SSH.Passwd = runArgs.Password
	}

	return &cluster
}

func (c *ClusterArgs) SetClusterArgs() error {
	err := PreProcessIPList(c.runArgs)
	if err != nil {
		return err
	}
	if net.IsIPList(c.runArgs.Masters) && (net.IsIPList(c.runArgs.Nodes) || c.runArgs.Nodes == "") {
		masters := strings.Split(c.runArgs.Masters, ",")
		nodes := strings.Split(c.runArgs.Nodes, ",")
		c.hosts = []v2.Host{}
		c.setHostWithIpsPort(masters, common.MASTER)
		if len(nodes) != 0 {
			c.setHostWithIpsPort(nodes, common.NODE)
		}
		c.cluster.Spec.Hosts = c.hosts
	} else {
		ip, err := net.GetLocalDefaultIP()
		if err != nil {
			return err
		}
		c.cluster.Spec.Hosts = []v2.Host{
			{
				IPS:   []string{ip},
				Roles: []string{common.MASTER},
			},
		}
	}
	return err
}

func getHostsWithIpsPort(ips []string, port uint16, role string) []v2.Host {
	//map[ssh port]*host
	hostMap := map[string]*v2.Host{}
	for i := range ips {
		ip, port := net.GetHostIPAndPortOrDefault(ips[i], strconv.Itoa(int(port)))
		if _, ok := hostMap[port]; !ok {
			hostMap[port] = &v2.Host{IPS: []string{ip}, Roles: []string{role}, SSH: v1.SSH{Port: port}}
			continue
		}
		hostMap[port].IPS = append(hostMap[port].IPS, ip)
	}

	hostsResult := []v2.Host{}
	_, master0Port := net.GetHostIPAndPortOrDefault(ips[0], strconv.Itoa(int(port)))
	for port, host := range hostMap {
		host.IPS = removeIPListDuplicatesAndEmpty(host.IPS)
		if port == master0Port && role == common.MASTER {
			hostsResult = append([]v2.Host{*host}, hostsResult...)
			continue
		}
		hostsResult = append(hostsResult, *host)
	}
	return hostsResult
}

func NewApplierFromArgs(imageName string, runArgs *Args) (applydriver.Interface, error) {
	if err := validateArgs(runArgs); err != nil {
		return nil, fmt.Errorf("failed to validate input args: %v", err)
	}
	cluster := ConstructClusterFromArg(runArgs)
	c := &ClusterArgs{
		cluster:   cluster,
		imageName: imageName,
		runArgs:   runArgs,
	}
	if err := c.SetClusterArgs(); err != nil {
		return nil, err
	}
	return NewApplier(c.cluster)
}

// validateArgs validates all the input args from sealer run command.
func validateArgs(runArgs *Args) error {
	// TODO: add detailed validation steps.
	return nil
}
