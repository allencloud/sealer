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

package kubernetes

import (
	"context"
	"fmt"
	"net"

	"github.com/sealerio/sealer/common"
	"github.com/sealerio/sealer/pkg/client/k8s"
	"github.com/sealerio/sealer/pkg/runtime/kubernetes/kubeadm/v1beta2"
	"github.com/sealerio/sealer/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Runtime) updateCert(certs []string) error {
	client, err := k8s.Newk8sClient()
	if err != nil {
		return err
	}
	cm, err := client.ConfigMap("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		return err
	}
	obj, err := utils.DecodeCRDFromString(cm.Data["ClusterConfiguration"], common.ClusterConfiguration)

	if err != nil {
		return err
	}
	clusterConfiguration, ok := obj.(*v1beta2.ClusterConfiguration)
	if !ok {
		return fmt.Errorf("failed to get ClusterConfiguration")
	}

	k.setCertSANS(append(clusterConfiguration.APIServer.CertSANs, certs...))
	ssh, err := k.getHostSSHClient(k.master0)
	if err != nil {
		return fmt.Errorf("failed to update cert: %v", err)
	}
	if err := ssh.CmdAsync(k.master0, "rm -rf /etc/kubernetes/admin.conf"); err != nil {
		return err
	}

	pipeline := []func() error{
		k.ConfigKubeadmOnMaster0,
		k.GenerateCert,
		k.CreateKubeConfig,
	}

	for _, f := range pipeline {
		if err := f(); err != nil {
			return fmt.Errorf("failed to init master0: %v", err)
		}
	}
	if err := k.SendJoinMasterKubeConfigs([]net.IP{k.master0}, AdminConf, ControllerConf, SchedulerConf, KubeletConf); err != nil {
		return err
	}

	if err := k.GetKubectlAndKubeconfig(); err != nil {
		return err
	}

	return nil
}
