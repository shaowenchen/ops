package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	deschedulerapioptions "sigs.k8s.io/descheduler/cmd/descheduler/app/options"
	deschedulerapi "sigs.k8s.io/descheduler/pkg/api"
	"sigs.k8s.io/descheduler/pkg/descheduler"
	deschedulereutils "sigs.k8s.io/descheduler/pkg/descheduler/evictions/utils"
)

func SetDeploymentNodeName(client *kubernetes.Clientset, nodeName string, namespacedName types.NamespacedName) (err error) {
	deploy, err := client.AppsV1().Deployments(namespacedName.Namespace).Get(
		context.TODO(),
		namespacedName.Name,
		metav1.GetOptions{},
	)
	deploy.Spec.Template.Spec.NodeName = nodeName
	deploy.Spec.Template.Spec.NodeSelector = nil
	_, err = client.AppsV1().Deployments(namespacedName.Namespace).Update(
		context.TODO(),
		deploy,
		metav1.UpdateOptions{},
	)
	return
}

func SetDeploymentNodeSelector(client *kubernetes.Clientset, namespacedName types.NamespacedName, labelPair map[string]string) (err error) {
	deploy, err := client.AppsV1().Deployments(namespacedName.Namespace).Get(
		context.TODO(),
		namespacedName.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return
	}
	deploy.Spec.Template.Spec.NodeName = ""
	deploy.Spec.Template.Spec.NodeSelector = labelPair
	if deploy.Spec.Template.Spec.Tolerations == nil {
		deploy.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	}
	allTolerationsKey := []string{}
	for _, tolerate := range deploy.Spec.Template.Spec.Tolerations {
		allTolerationsKey = append(allTolerationsKey, tolerate.Key)
	}

	for labelKey, labelValue := range labelPair {
		deploy.Spec.Template.Spec.Tolerations = []corev1.Toleration{
			corev1.Toleration{
				Effect:   corev1.TaintEffectNoSchedule,
				Key:      labelKey,
				Operator: corev1.TolerationOpEqual,
				Value:    labelValue,
			},
		}
	}
	_, err = client.AppsV1().Deployments(namespacedName.Namespace).Update(
		context.TODO(),
		deploy,
		metav1.UpdateOptions{},
	)
	return
}

func CreateLimitRange(client *kubernetes.Clientset, namespacedName types.NamespacedName, reqMem, limitMem, reqCPU, limitCPU string) (err error) {
	defaultReqRes := corev1.ResourceList{}
	defaultRes := corev1.ResourceList{}
	if len(reqMem) > 0 {
		reqMemQuantity := resource.QuantityValue{}
		reqMemQuantity.Set(reqMem)
		defaultReqRes[corev1.ResourceMemory] = reqMemQuantity.Quantity
	}
	if len(limitMem) > 0 {
		limitMemQuantity := resource.QuantityValue{}
		limitMemQuantity.Set(limitMem)
		defaultRes[corev1.ResourceMemory] = limitMemQuantity.Quantity
	}
	if len(reqCPU) > 0 {
		reqCPUQuantity := resource.QuantityValue{}
		reqCPUQuantity.Set(reqCPU)
		defaultReqRes[corev1.ResourceCPU] = reqCPUQuantity.Quantity
	}
	if len(limitCPU) > 0 {
		limitCPUQuantity := resource.QuantityValue{}
		limitCPUQuantity.Set(limitCPU)
		defaultRes[corev1.ResourceCPU] = limitCPUQuantity.Quantity
	}
	_, err = client.CoreV1().LimitRanges(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespacedName.Name,
			},
			Spec: corev1.LimitRangeSpec{
				Limits: []corev1.LimitRangeItem{
					corev1.LimitRangeItem{
						Type:           corev1.LimitTypeContainer,
						Default:        defaultRes,
						DefaultRequest: defaultReqRes,
					}},
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}
	return
}

func DeleteLimitRange(client *kubernetes.Clientset, namespacedName types.NamespacedName) (err error) {
	return client.CoreV1().LimitRanges(namespacedName.Namespace).Delete(
		context.TODO(),
		namespacedName.Name,
		metav1.DeleteOptions{},
	)
}

func RunDeScheduler(config *rest.Config, client *kubernetes.Clientset, removeDuplicates, nodeUtilization bool, highpercent int16) (err error) {
	rs, err := deschedulerapioptions.NewDeschedulerServer()
	if err != nil {
		return
	}
	rs.Client = client
	var SecureServing *apiserver.SecureServingInfo
	if err := rs.SecureServing.ApplyTo(&SecureServing, &config); err != nil {
		return err
	}
	strategList := deschedulerapi.StrategyList{}
	if removeDuplicates {
		strategList["RemoveDuplicates"] = deschedulerapi.DeschedulerStrategy{
			Enabled: true,
			Params:  &deschedulerapi.StrategyParameters{},
		}
	}
	if nodeUtilization {
		strategList["LowNodeUtilization"] = deschedulerapi.DeschedulerStrategy{
			Enabled: true,
			Params:  &deschedulerapi.StrategyParameters{
				NodeResourceUtilizationThresholds: &deschedulerapi.NodeResourceUtilizationThresholds{
					Thresholds: deschedulerapi.ResourceThresholds{
						corev1.ResourceCPU: deschedulerapi.Percentage(20),
						corev1.ResourceMemory: deschedulerapi.Percentage(20),
						corev1.ResourcePods: deschedulerapi.Percentage(20),
					},
					TargetThresholds: deschedulerapi.ResourceThresholds{
						corev1.ResourceCPU: deschedulerapi.Percentage(highpercent),
						corev1.ResourceMemory: deschedulerapi.Percentage(highpercent),
						corev1.ResourcePods: deschedulerapi.Percentage(highpercent),
					},
				},
			},
		}
		strategList["HighNodeUtilization"] = deschedulerapi.DeschedulerStrategy{
			Enabled: true,
			Params:  &deschedulerapi.StrategyParameters{
				NodeResourceUtilizationThresholds: &deschedulerapi.NodeResourceUtilizationThresholds{
					Thresholds: deschedulerapi.ResourceThresholds{
						corev1.ResourceCPU: deschedulerapi.Percentage(highpercent),
						corev1.ResourceMemory: deschedulerapi.Percentage(highpercent),
						corev1.ResourcePods: deschedulerapi.Percentage(highpercent),
					},
				},
			},
		}
	}
	evictionPolicyGroupVersion, err := deschedulereutils.SupportEviction(client)
	if err != nil || len(evictionPolicyGroupVersion) == 0 {
		fmt.Printf("Error when checking support for eviction: %v\n", err)
		return
	}
	err = descheduler.RunDeschedulerStrategies(context.TODO(), rs, &deschedulerapi.DeschedulerPolicy{
		Strategies: strategList,
	}, evictionPolicyGroupVersion)
	return
}
