package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetOrCreateNamespacedName(client *kubernetes.Clientset, namespace, name string) (namespacedName types.NamespacedName, err error) {
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		ns, err = client.CoreV1().Namespaces().Create(
			context.TODO(),
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			return
		}
	}
	namespacedName = types.NamespacedName{
		Namespace: ns.Name,
		Name:      name,
	}
	return
}

func BuildNamespacedName(namespace, name string) (namespacedName types.NamespacedName) {
	namespacedName.Name = name
	if len(namespace) == 0 {
		namespacedName.Namespace = corev1.NamespaceDefault
	} else {
		namespacedName.Namespace = namespace
	}
	return namespacedName
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

func GetAllNodesFromKubeconfig(kubeconfigpath string) (nodes_ips []string, err error) {
	client, err := NewKubernetesClient(kubeconfigpath)
	if err != nil {
		return
	}
	nodes, err := GetAllNodes(client)
	if err != nil {
		return
	}
	for _, node := range nodes.Items {
		node_ip := GetInterlIPFromKubeNode(node)
		nodes_ips = append(nodes_ips, node_ip)
	}
	return
}

func GetInterlIPFromKubeNode(node corev1.Node) (ip string) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			return addr.Address
		}
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
	return nil, fmt.Errorf("could not locate a kubeconfig")
}
