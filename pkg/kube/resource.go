package kube

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const PrometheusServerKey = "prometheus-server"

const PromQLPodCpuUsageReqM = "quantile_over_time(0.666, irate(container_cpu_usage_seconds_total{namespace='%s', image!='', pod=~'%s', container='%s'}[5m])[1d:5m]) * 1000"
const PromQLPodCpuUsageLimitM = "quantile_over_time(0.999, irate(container_cpu_usage_seconds_total{namespace='%s', image!='', pod=~'%s', container='%s'}[1d])[3d:5m]) * 1000"

const PromQLPodMemUsageReqMB = "quantile_over_time(0.666, container_memory_working_set_bytes{namespace='%s', image!='', pod=~'%s', container='%s'}[1d])/1000/1000"
const PromQLPodMemUsageLimitMB = "quantile_over_time(0.999, container_memory_working_set_bytes{namespace='%s', image!='', pod=~'%s', container='%s'}[1d])/1000/1000"

func SetDeploymentRecommandResource(client *kubernetes.Clientset, namespacedName types.NamespacedName) (err error) {
	app, err := client.AppsV1().Deployments(namespacedName.Namespace).Get(context.TODO(), namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(app.Spec.Selector.MatchLabels).String(),
	}
	pods, err := client.CoreV1().Pods(namespacedName.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return errors.New("resource have not pod")
	}
	var podList []string
	for _, pod := range pods.Items {
		podList = append(podList, pod.Name)
	}
	// set req\limit for all containers
	for _, container := range pods.Items[0].Spec.Containers {
		res, err := GetPodsContainerRecommandResource(client, namespacedName.Namespace, podList, container.Name)
		if err != nil {
			return err
		}
		SetContainerResource(app.Spec.Template.Spec.Containers, container, res)
	}
	_, err = client.AppsV1().Deployments(namespacedName.Namespace).Update(context.TODO(), app, metav1.UpdateOptions{})
	return
}

func SetStatefulSetRecommandResource(client *kubernetes.Clientset, namespacedName types.NamespacedName) (err error) {
	app, err := client.AppsV1().StatefulSets(namespacedName.Namespace).Get(context.TODO(), namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(app.Spec.Selector.MatchLabels).String(),
	}
	pods, err := client.CoreV1().Pods(namespacedName.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return errors.New("resource have not pod")
	}
	var podList []string
	for _, pod := range pods.Items {
		podList = append(podList, pod.Name)
	}
	// set req\limit for all containers
	for _, container := range pods.Items[0].Spec.Containers {
		res, err := GetPodsContainerRecommandResource(client, namespacedName.Namespace, podList, container.Name)
		if err != nil {
			return err
		}
		SetContainerResource(app.Spec.Template.Spec.Containers, container, res)
	}
	_, err = client.AppsV1().StatefulSets(namespacedName.Namespace).Update(context.TODO(), app, metav1.UpdateOptions{})
	return
}

func SetDaemonSetRecommandResource(client *kubernetes.Clientset, namespacedName types.NamespacedName) (err error) {
	app, err := client.AppsV1().DaemonSets(namespacedName.Namespace).Get(context.TODO(), namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(app.Spec.Selector.MatchLabels).String(),
	}
	pods, err := client.CoreV1().Pods(namespacedName.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return errors.New("resource have not pod")
	}
	var podList []string
	for _, pod := range pods.Items {
		podList = append(podList, pod.Name)
	}
	// set req\limit for all containers
	for _, container := range pods.Items[0].Spec.Containers {
		res, err := GetPodsContainerRecommandResource(client, namespacedName.Namespace, podList, container.Name)
		if err != nil {
			return err
		}
		SetContainerResource(app.Spec.Template.Spec.Containers, container, res)
	}
	_, err = client.AppsV1().DaemonSets(namespacedName.Namespace).Update(context.TODO(), app, metav1.UpdateOptions{})
	return
}

func SetContainerResource(containers []corev1.Container, container corev1.Container, res corev1.ResourceRequirements) bool {
	for index, item := range containers {
		if item.Name == container.Name {
			containers[index].Resources = res
			return true
		}
	}
	return false
}

func GetPodsContainerRecommandResource(client *kubernetes.Clientset, namespace string, podList []string, containerName string) (res corev1.ResourceRequirements, err error) {
	println(containerName)
	serverUrl, err := GetPrometheusServerUrl(client)
	if err != nil {
		return
	}
	// get and set Request\Limit for mem
	memReqQueryResult, err := PromQueryRangeMatrix(serverUrl, fmt.Sprintf(PromQLPodMemUsageReqMB,
		namespace, strings.Join(podList, "|"), containerName))
	if err != nil {
		return
	}
	memReqContainersValue := GetMatrixContainerValue(memReqQueryResult)
	memReq := int64(memReqContainersValue[containerName])
	memReq = notLess(memReq, 50)
	println("memReq:", memReq)

	cpuReqQueryResult, err := PromQueryRangeMatrix(serverUrl, fmt.Sprintf(PromQLPodCpuUsageReqM, namespace, strings.Join(podList, "|"), containerName))
	if err != nil {
		return
	}
	cpuReqContainersValue := GetMatrixContainerValue(cpuReqQueryResult)
	cpuReq := int64(cpuReqContainersValue[containerName])
	cpuReq = notLess(cpuReq, 50)
	println("cpuReq:", cpuReq)

	request := corev1.ResourceList{}

	request[corev1.ResourceMemory] = *resource.NewQuantity(
		int64(memReq*1000*1000),
		resource.BinarySI)
	request[corev1.ResourceCPU] = *resource.NewMilliQuantity(
		int64(cpuReq),
		resource.DecimalSI)

	// get and set Request\Limit for cpu
	memLimitQueryResult, err := PromQueryRangeMatrix(serverUrl, fmt.Sprintf(PromQLPodMemUsageLimitMB, namespace, strings.Join(podList, "|"), containerName))
	if err != nil {
		return
	}
	memLimitContainersValue := GetMatrixContainerValue(memLimitQueryResult)
	memLimit := int64(memLimitContainersValue[containerName])
	memLimit = notLess(memLimit, 100)
	println("memLimit:", memLimit)

	cpuLimitQueryResult, err := PromQueryRangeMatrix(serverUrl, fmt.Sprintf(PromQLPodCpuUsageLimitM, namespace, strings.Join(podList, "|"), containerName))
	if err != nil {
		return
	}
	cpuLimitContainersValue := GetMatrixContainerValue(cpuLimitQueryResult)
	cpuLimit := int64(cpuLimitContainersValue[containerName])
	cpuLimit = notLess(cpuLimit, 100)
	println("cpuLimit:", cpuLimit)

	limits := corev1.ResourceList{}
	limits[corev1.ResourceMemory] = *resource.NewQuantity(
		int64(memLimit*1000*1000),
		resource.BinarySI)
	limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(
		int64(cpuLimit),
		resource.DecimalSI)

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: request,
	}, nil
}

func GetMatrixContainerValue(matrix prommodel.Matrix) (result map[string]prommodel.SampleValue) {
	result = make(map[string]prommodel.SampleValue)
	for _, sample := range matrix {
		for _, value := range sample.Values {
			if float64(value.Value) > float64(result[string(sample.Metric["container"])]) {
				result[string(sample.Metric["container"])] = value.Value
			}
		}
	}
	return
}

func PromQueryRangeMatrix(serverUrl string, promQl string) (result prommodel.Matrix, err error) {
	client, err := promapi.NewClient(promapi.Config{
		Address: serverUrl,
	})
	if err != nil {
		return
	}
	v1api := promv1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := promv1.Range{
		Start: time.Now().Add(-time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}
	retValue, _, err := v1api.QueryRange(ctx, promQl, r, promv1.WithTimeout(5*time.Second))
	if err != nil {
		return
	}
	result, ok := retValue.(prommodel.Matrix)
	if !ok {
		err = errors.New("not Matrix type")
	}
	return
}

func PromQuery(serverUrl string, promQl string) (result *prommodel.Sample, err error) {
	client, err := promapi.NewClient(promapi.Config{
		Address: serverUrl,
	})
	if err != nil {
		return
	}
	v1api := promv1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	retValue, _, err := v1api.Query(ctx, promQl, time.Now())
	if err != nil {
		return
	}
	switch {
	case retValue.Type() == prommodel.ValScalar:
		// handle scalar stuff
	case retValue.Type() == prommodel.ValVector:
		// handle vector stuff
		results := retValue.(prommodel.Vector)
		if len(results) == 0 {
			err = errors.New("no result")
			return
		}
		result = results[0]
	case retValue.Type() == prommodel.ValMatrix:
		// handle matrix stuff
	case retValue.Type() == prommodel.ValString:
		// handle string stuff
	}

	return result, nil
}

func GetPrometheusServerUrl(client *kubernetes.Clientset) (url string, err error) {
	node, err := utils.GetAnyMaster(client)
	if err != nil {
		return
	}
	ip := utils.GetNodeInternalIp(node)
	port, err := GetPrometheusServerNodePort(client, "")
	if err != nil {
		return
	}
	return fmt.Sprintf("http://%s:%d", ip, port), nil
}

func GetPrometheusServerNodePort(client *kubernetes.Clientset, name string) (nodePort int32, err error) {
	if len(name) == 0 {
		name = PrometheusServerKey
	}
	svcs, err := client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, svc := range svcs.Items {
		if !strings.Contains(svc.Name, name) {
			continue
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == "http" && port.Port == 80 {
				return port.NodePort, nil
			}
		}
	}
	return
}

func notLess(value int64, baseline int64) int64 {
	if value > baseline {
		return value
	} else {
		return baseline
	}
}
