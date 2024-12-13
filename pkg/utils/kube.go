package utils

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/shaowenchen/ops/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetClusterUID(k8sClient runtimeClient.Client) (string, error) {
	kubeSystemNamespace := &corev1.Namespace{}
	err := k8sClient.Get(context.TODO(), runtimeClient.ObjectKey{Name: "kube-system"}, kubeSystemNamespace)
	if err != nil {
		return "", err
	}
	return string(kubeSystemNamespace.UID), nil
}

func GetRestConfigByContent(kubeconfig string) (*rest.Config, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	return restConfig, err
}

func NewKubernetesClient(kubeconfigpath string) (client *kubernetes.Clientset, err error) {
	kubeconfigpath = GetAbsoluteFilePath(kubeconfigpath)
	restConfig, err := GetRestConfig(kubeconfigpath)
	if err != nil {
		restConfig, err = GetInClusterConfig()
		if err != nil {
			return
		}
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
	// try KUBECONFIG
	if kubeconfig := os.Getenv("KUBECONFIG"); len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	// try inCluster
	if _, err := rest.InClusterConfig(); err == nil {
		return rest.InClusterConfig()
	}
	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func GetInClusterConfig() (*rest.Config, error) {
	c, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetServerUrl(kubeconfigPath string) (string, error) {
	restConfig, err := GetRestConfig(kubeconfigPath)
	if err != nil {
		return "", err
	}
	return restConfig.Host, nil
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

func IsNodeReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func IsSucceededPod(pod *corev1.Pod) bool {
	status := pod.Status.Phase
	if status == corev1.PodSucceeded {
		return true
	}
	return false
}

func IsUnknownPod(pod *corev1.Pod) bool {
	status := pod.Status.ContainerStatuses
	if len(status) > 0 && status[0].State.Terminated != nil {
		return status[0].State.Terminated.Reason == "Unknown"
	}
	return false
}

func IsFailedPod(pod *corev1.Pod) bool {
	status := pod.Status.Phase
	if status == corev1.PodFailed {
		return true
	}
	if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
		if pod.Status.ContainerStatuses[0].LastTerminationState.Terminated != nil {
			return pod.Status.ContainerStatuses[0].LastTerminationState.Terminated.ExitCode > 0
		}
		if pod.Status.ContainerStatuses[0].State.Waiting != nil && strings.Contains(pod.Status.ContainerStatuses[0].State.Waiting.Reason, "BackOff") {
			return true
		}
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

func GetAllNodesByReconcileClient(client runtimeClient.Client) (nodes *corev1.NodeList, err error) {
	nodes = &corev1.NodeList{}
	err = client.List(context.TODO(), nodes)
	return
}

func GetAllReadyNodesByClient(client *kubernetes.Clientset) (nodes *corev1.NodeList, err error) {
	nodes, err = GetAllNodesByClient(client)
	if err != nil {
		return
	}
	for i, node := range nodes.Items {
		if !IsNodeReady(&node) {
			nodes.Items = append(nodes.Items[:i], nodes.Items[i+1:]...)
		}
	}
	return
}

func GetAllReadyNodesByReconcileClient(client runtimeClient.Client) (nodes *corev1.NodeList, err error) {
	nodes, err = GetAllNodesByReconcileClient(client)
	if err != nil {
		return
	}
	for i, node := range nodes.Items {
		if !IsNodeReady(&node) {
			nodes.Items = append(nodes.Items[:i], nodes.Items[i+1:]...)
		}
	}
	return
}

func GetAnyReadyNodesByReconcileClient(client runtimeClient.Client) (node *corev1.Node, err error) {
	nodes, err := GetAllReadyNodesByReconcileClient(client)
	if err != nil || len(nodes.Items) == 0 {
		return
	}
	// random select a ready node
	random := rand.Intn(len(nodes.Items))
	return &nodes.Items[random], nil
}

func GetNodeByClient(client *kubernetes.Clientset, nodeName string) (node *corev1.Node, err error) {
	return client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
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

func IsMasterNode(node *corev1.Node) bool {
	_, ok := node.Labels[constants.LabelNodeRoleMaster]
	if !ok {
		_, ok = node.Labels[constants.LabelNodeRoleControlPlane]
	}
	return ok
}

func GetAnyMaster(client *kubernetes.Clientset) (master *corev1.Node, err error) {
	nodes, err := GetAllNodesByClient(client)
	if err != nil {
		return
	}
	for _, node := range nodes.Items {
		if IsMasterNode(&node) {
			return &node, nil
		}
	}
	return nil, errors.New("not found master")
}

func GetNodeInternalIp(node *corev1.Node) (ip string) {
	for _, add := range node.Status.Addresses {
		if add.Type == corev1.NodeInternalIP {
			return add.Address
		}
	}
	return ip
}

func GetCertNotAfterDays(c *rest.Config) (days int, err error) {
	certDerBlock, _ := pem.Decode(c.CertData)
	if certDerBlock == nil {
		err = errors.New("failed to decode PEM block containing the client certificate")
		return
	}
	x509Cert, err := x509.ParseCertificate(certDerBlock.Bytes)
	if err != nil {
		return
	}
	days = int(x509Cert.NotAfter.Sub(time.Now()).Hours() / 24)
	return
}
