package kube

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func RunScriptOnEachNode(client *kubernetes.Clientset, namespacedName types.NamespacedName, script string) (err error) {
	nodeNames, err := GetAllNodeNames(client)
	if err != nil {
		return PrintError(ErrorMsgGetNode(err))
	}
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

func RunScriptOnNode(client *kubernetes.Clientset, nodeName string, namespacedName types.NamespacedName, script string) (job *batchv1.Job, err error) {
	priviBool := true
	job, err = client.BatchV1().Jobs(OpsCliNamespace).Create(
		context.TODO(),
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						NodeName: nodeName,
						Containers: []corev1.Container{
							{
								Name:    "etchosts",
								Image:   "docker.io/library/alpine:latest",
								Command: []string{"sh"},
								Args:    []string{"-c", "echo '" + script + "' | nsenter -t 1 -m -u -i -n"},
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
			},
		},
		metav1.CreateOptions{},
	)
	return
}
