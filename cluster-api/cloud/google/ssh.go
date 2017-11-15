/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package google

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"

)

const (
	MachineControllerSshKeySecret = "machine-controller-sshkeys"
)

// It creates secret to store priate key.
func (gce *GCEClient) SetupSSHAccess(privateKeyFile string, user string) error {
	err := run("kubectl", "create", "secret", "generic", "-n", "kube-system", MachineControllerSshKeySecret, "--from-file=private="+privateKeyFile, "--from-literal=user="+user)
	if err != nil {
		return fmt.Errorf("couldn't create service account key as credential: %v", err)
	}
	return err
}

func (gce *GCEClient) remoteSshCommand(master *clusterv1.Machine, cmd string) (string, error) {
	glog.Infof("Remote SSH execution '%s' on %s", cmd, master.ObjectMeta.Name)

	command := fmt.Sprintf("echo STARTFILE; %s", cmd)
	c := exec.Command("ssh", "-i", gce.sshCreds.privateKeyPath, gce.sshCreds.user+"@"+sanitizeMasterIP(gce.masterIP), command)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error: %v, output: %s", err, string(out))
	}
	result := strings.TrimSpace(string(out))
	parts := strings.Split(result, "STARTFILE")
	if len(parts) != 2 {
		return "", nil
	}
	// TODO: Check error.
	return strings.TrimSpace(parts[1]), nil
}