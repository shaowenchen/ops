package kube

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ClearPod(client *kubernetes.Clientset, namespace string, statusList []string) (err error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, pod := range pods.Items {
		podStatus := GetPodStatus(&pod)
		for _, statue := range statusList {
			if strings.ToLower(podStatus) == strings.ToLower(statue) {
				err = client.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			}
		}
	}
	return
}

func GetPodStatus(pod *corev1.Pod) string {
	for _, cond := range pod.Status.Conditions {
		if string(cond.Type) == ContainersReady {
			if string(cond.Status) != ConditionTrue {
				return "Unavailable"
			}
		} else if string(cond.Type) == PodInitialized && string(cond.Status) != ConditionTrue {
			return "Initializing"
		} else if string(cond.Type) == PodReady {
			if string(cond.Status) != ConditionTrue {
				return "Unavailable"
			}
			for _, containerState := range pod.Status.ContainerStatuses {
				if !containerState.Ready {
					return "Unavailable"
				}
			}
		} else if string(cond.Type) == PodScheduled && string(cond.Status) != ConditionTrue {
			return "Scheduling"
		}
	}
	return string(pod.Status.Phase)
}
