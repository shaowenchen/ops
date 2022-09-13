package kube

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func GetOpscliNamespacedName(client *kubernetes.Clientset, name string) (namespacedName types.NamespacedName, err error) {
	opscliNamespace, err := client.CoreV1().Namespaces().Get(context.TODO(), OpsCliNamespace, metav1.GetOptions{})
	if err != nil {
		opscliNamespace, err = client.CoreV1().Namespaces().Create(
			context.TODO(),
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: OpsCliNamespace,
				},
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			return
		}
	}
	namespacedName = types.NamespacedName{
		Namespace: opscliNamespace.Name,
		Name:      name,
	}
	return
}

func SplitNamespacedName(namespacedNamesStr string) (namespacedNames []types.NamespacedName) {
	namespacedNameStrList := strings.Split(namespacedNamesStr, ",")
	for _, namenamespacedNameStr := range namespacedNameStrList {
		namespacedNameSplit := strings.Split(namenamespacedNameStr, "/")
		if len(namespacedNameSplit) == 1 {
			namespacedNames = append(namespacedNames,
				types.NamespacedName{
					Namespace: corev1.NamespaceDefault,
					Name:      namespacedNameSplit[0],
				})
		} else if len(namespacedNameSplit) == 2 {
			namespacedNames = append(namespacedNames,
				types.NamespacedName{
					Namespace: namespacedNameSplit[0],
					Name:      namespacedNameSplit[1],
				})
		} else {
			continue
		}
	}
	return
}

func SplitAllNamespacedName(client *kubernetes.Clientset, namespacedNamesStr string) (namespacedNames []types.NamespacedName, err error) {
	partNamespacedNames := SplitNamespacedName(namespacedNamesStr)
	names := []string{}
	allNamespaces, err := GetAllNamespaces(client)
	if err != nil {
		return
	}
	for _, namespacedName := range partNamespacedNames {
		if IsContainKey(names, namespacedName.Name) {
			continue
		}
		names = append(names, namespacedName.Name)
		for _, namespace := range allNamespaces {
			namespacedNames = append(namespacedNames, types.NamespacedName{Namespace: namespace, Name: namespacedName.Name})
		}

	}
	return
}

func SplitStr(str string)(strList []string){
	return strings.Split(str, ",")
}

func IsContainKey(targets []string, target string) bool {
	for _, item := range targets {
		if item == target {
			return true
		}
	}
	return false
}

func SplitKeyValues(str string) (pair map[string]string) {
	keyLabels := strings.Split(str, ",")
	for _, keyLabel := range keyLabels {
		keyLabelPair := strings.Split(keyLabel, "=")
		if len(keyLabelPair) == 2 {
			if pair == nil {
				pair = make(map[string]string)
			}
			pair[keyLabelPair[0]] = keyLabelPair[1]
		}
	}
	return
}

func GetAllNamespaces(client *kubernetes.Clientset) (namespaces []string, err error) {
	allNamespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, namespace := range allNamespaces.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return
}

func GetAllNodes(client *kubernetes.Clientset) (nodes *corev1.NodeList, err error) {
	return client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}

func NewKubernetesClient(kubeconfigpath string) (client *kubernetes.Clientset, err error) {
	kubeconfig, err := GetRestConfig(kubeconfigpath)
	if err != nil {
		return
	}
	kubeconfig.QPS = 100
	kubeconfig.Burst = 100
	client, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return
	}
	return
}

func GetRestConfig(kubeconfigPath string) (*rest.Config, error) {
	if len(kubeconfigPath) > 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}