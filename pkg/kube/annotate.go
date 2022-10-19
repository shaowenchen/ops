package kube

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const VeleroBackupAnnotationKey = "backup.velero.io/backup-volumes"

func AnnotateVeleroPod(client *kubernetes.Clientset, namespace string, clear bool) (updatedPodNames []string, err error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	for _, pod := range pods.Items {
		var volumes []string
		for _, podVolume := range pod.Spec.Volumes {
			if podVolume.PersistentVolumeClaim != nil {
				volumes = append(volumes, podVolume.Name)
			}
		}
		if len(volumes) > 0 {
			if clear {
				delete(pod.Annotations, VeleroBackupAnnotationKey)
			} else {
				pod.Annotations[VeleroBackupAnnotationKey] = strings.Join(volumes, ",")
			}
			updatedPod, err := client.CoreV1().Pods(namespace).Update(context.TODO(), &pod, metav1.UpdateOptions{})
			updatedPodNames = append(updatedPodNames, updatedPod.Namespace+"/"+updatedPod.Name)
			if err != nil {
				return updatedPodNames, err
			}
		}
	}
	return
}
