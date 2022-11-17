package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetRestConfigByContent(kubeconfig string) (*rest.Config, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	return restConfig, err
}

func NewKubernetesClient(kubeconfigpath string) (client *kubernetes.Clientset, err error) {
	kubeconfigpath = GetAbsoluteFilePath(kubeconfigpath)
	restConfig, err := GetRestConfig(kubeconfigpath)
	if err != nil {
		return
	}
	return GetClientByRestconfig(restConfig)
}

func GetClientByRestconfig(restConfig *rest.Config) (client *kubernetes.Clientset, err error) {
	restConfig.QPS = 100
	restConfig.Burst = 100
	client, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return
	}
	return
}

func GetRestConfig(kubeconfigPath string) (*rest.Config, error) {
	if len(kubeconfigPath) > 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func NewConrollerRuntimeClient(kubeconfigPath string) (c runtimeClient.Client, err error) {
	restConfig, err := GetRestConfig(kubeconfigPath)
	if err != nil {
		return
	}
	c, err = runtimeClient.New(restConfig, runtimeClient.Options{})
	return
}

func GetAllNodesByKubeconfig(kubeconfigpath string) (nodes_ips []string, err error) {
	client, err := NewKubernetesClient(kubeconfigpath)
	if err != nil {
		return
	}
	nodes, err := GetAllNodesByClient(client)
	if err != nil {
		return
	}
	for _, node := range nodes.Items {
		node_ip := GetInterlIByNode(node)
		nodes_ips = append(nodes_ips, node_ip)
	}
	return
}

func GetInterlIByNode(node corev1.Node) (ip string) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			return addr.Address
		}
	}
	return
}

func IsStopedPod(pod *corev1.Pod) bool {
	status := pod.Status.Phase
	if status == corev1.PodFailed || status == corev1.PodSucceeded {
		return true
	}
	return false
}

func IsPendingPod(pod *corev1.Pod) bool {
	status := pod.Status.Phase
	if status == corev1.PodPending {
		return true
	}
	return false
}

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

func GetAllNodesByClient(client *kubernetes.Clientset) (nodes *corev1.NodeList, err error) {
	return client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}

func GetPodLog(ctx context.Context, client *kubernetes.Clientset, namespace, podName string) (log string, err error) {
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return
	}
	log = buf.String()

	return
}
