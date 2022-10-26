package kube

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	cmdcreate "k8s.io/kubectl/pkg/cmd/create"
)

func CreateImagePullSecret(client *kubernetes.Clientset, namespacedName types.NamespacedName, hosts, username, password string) (secret *corev1.Secret, err error) {
	auths := map[string]cmdcreate.DockerConfigEntry{}
	for _, host := range strings.Split(hosts, ",") {
		auths[host] = cmdcreate.DockerConfigEntry{Auth: base64.StdEncoding.EncodeToString([]byte(username + ":" + password))}
	}
	config, err := json.Marshal(cmdcreate.DockerConfigJSON{
		Auths: auths,
	})
	secret, err = client.CoreV1().Secrets(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespacedName.Name,
			},
			Type: "kubernetes.io/dockerconfigjson",
			Data: map[string][]byte{
				".dockerconfigjson": config,
			},
		},
		metav1.CreateOptions{},
	)
	return
}

func DeleteSecret(client *kubernetes.Clientset, namespacedName types.NamespacedName) (secret *corev1.Secret, err error) {
	err = client.CoreV1().Secrets(namespacedName.Namespace).Delete(
		context.TODO(),
		namespacedName.Name,
		metav1.DeleteOptions{},
	)
	return
}
