package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func RunScriptOnNodes(client *kubernetes.Clientset, nodeNames []string, namespacedName types.NamespacedName, script string) (err error) {
	for index, nodeName := range nodeNames {
		currentNameSpacedName := namespacedName
		currentNameSpacedName.Name = fmt.Sprintf("node%d-%s", index, currentNameSpacedName.Name)
		_, err = RunScriptOnNode(client, nodeName, currentNameSpacedName, script)
		if err != nil {
			PrintError(err.Error())
		}
	}
	return
}

func RunScriptOnNode(client *kubernetes.Clientset, nodeName string, namespacedName types.NamespacedName, script string) (pod *corev1.Pod, err error) {
	priviBool := true
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: corev1.PodSpec{
				NodeName: nodeName,
				Containers: []corev1.Container{
					{
						Name:    "etchosts",
						Image:   "docker.io/library/alpine:latest",
						Command: []string{"sh"},
						Args:    []string{"-c", "echo 'sudo " + script + "' | nsenter -t 1 -m -u -i -n"},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &priviBool,
						},
					},
				},
				HostIPC:       true,
				HostNetwork:   true,
				HostPID:       true,
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
		metav1.CreateOptions{},
	)
	return
}
