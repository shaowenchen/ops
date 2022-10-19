package kube

import (
	"context"
	"fmt"

	"github.com/shaowenchen/opscli/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func RunScriptOnNodes(client *kubernetes.Clientset, nodes []v1.Node, namespacedName types.NamespacedName, script string) (err error) {
	for index, node := range nodes {
		currentNameSpacedName := namespacedName
		currentNameSpacedName.Name = fmt.Sprintf("node%d-%s", index, currentNameSpacedName.Name)
		_, err = RunScriptOnNode(client, node, currentNameSpacedName, script)
		if err != nil {
			return utils.LogError(err)
		}
	}
	return
}

func RunScriptOnNode(client *kubernetes.Clientset, node v1.Node, namespacedName types.NamespacedName, script string) (pod *corev1.Pod, err error) {
	priviBool := true
	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:    "etchosts",
						Image:   "docker.io/library/alpine:latest",
						Command: []string{"sh"},
						Args:    []string{"-c", "echo \"sudo " + script + "\" | nsenter -t 1 -m -u -i -n"},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &priviBool,
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
				HostIPC:       true,
				HostNetwork:   true,
				HostPID:       true,
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
			},
		},
		metav1.CreateOptions{},
	)
	return
}
