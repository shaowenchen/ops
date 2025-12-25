package kube

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func buildVolumesAndMounts(mounts []option.MountConfig) ([]v1.Volume, []v1.VolumeMount) {
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}

	for i, mount := range mounts {
		volumeName := fmt.Sprintf("mount-%d", i)

		// Handle secret mount
		if mount.Secret != nil {
			volumes = append(volumes, v1.Volume{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: mount.Secret.Name,
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      volumeName,
				MountPath: mount.Secret.MountPath,
			})
		} else if mount.ConfigMap != nil {
			// Handle configMap mount
			volumes = append(volumes, v1.Volume{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: mount.ConfigMap.Name,
						},
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      volumeName,
				MountPath: mount.ConfigMap.MountPath,
			})
		} else if mount.HostPath != "" {
			// Handle hostPath mount
			volumes = append(volumes, v1.Volume{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: mount.HostPath,
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      volumeName,
				MountPath: mount.MountPath,
			})
		}
	}

	return volumes, volumeMounts
}

func RunShellOnNode(client *kubernetes.Clientset, node *v1.Node, namespacedName types.NamespacedName, image string, mode string, shell string, mounts []option.MountConfig) (pod *corev1.Pod, err error) {
	if image == "" {
		image = constants.DefaultRuntimeImage
	}
	// choose interpreter
	usePython := false
	lines := strings.Split(shell, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "python") {
		usePython = true
	}
	priviBool := true
	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	pull := corev1.PullIfNotPresent
	cmdArg := []string{}
	shellBase64 := utils.EncodingStringToBase64(shell)
	// mode
	if mode == constants.ModeContainer {
		cmdArg = []string{"-c", "echo " + shellBase64 + " | base64 -d | bash"}
		pull = corev1.PullAlways
	} else {
		cmdArg = []string{"-c", "echo " + shellBase64 + " | base64 -d | nsenter -t 1 -m -u -i -n"}
	}
	if usePython {
		cmdArg[1] = cmdArg[1] + " -- python3 /dev/stdin"
	}
	hostFlag := true
	volumes, volumeMounts := buildVolumesAndMounts(mounts)
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
				Labels: map[string]string{
					constants.LabelOpsTaskKey: constants.LabelOpsTaskValue,
				},
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:    "shell",
						Image:   image,
						Command: []string{"bash"},
						Args:    cmdArg,
						SecurityContext: &corev1.SecurityContext{
							Privileged: &priviBool,
						},
						ImagePullPolicy: pull,
						VolumeMounts:    volumeMounts,
					},
				},
				HostIPC:       hostFlag,
				HostNetwork:   hostFlag,
				HostPID:       hostFlag,
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
				Volumes:       volumes,
			},
		},
		metav1.CreateOptions{},
	)
	return
}

func RunFileOnNode(client *kubernetes.Clientset, node *v1.Node, namespacedName types.NamespacedName, fileOpt option.FileOption) (pod *corev1.Pod, err error) {
	// find mount path for host root, default to /host if not found
	hostMountPath := "/host"
	for _, mount := range fileOpt.Mounts {
		if mount.HostPath == "/" {
			hostMountPath = mount.MountPath
			break
		}
	}
	hostLocalfile := hostMountPath + fileOpt.LocalFile
	cmd := ""
	switch fileOpt.GetStorageType() {
	case constants.RemoteStorageTypeS3:
		if fileOpt.IsDownloadDirection() {
			cmd = utils.ShellOpscliDownS3(fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket,
				fileOpt.AK, fileOpt.SK, hostLocalfile, fileOpt.RemoteFile)
		} else if fileOpt.IsUploadDirection() {
			cmd = utils.ShellOpscliUploadS3(fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket,
				fileOpt.AK, fileOpt.SK, hostLocalfile, fileOpt.RemoteFile)
		}
	case constants.RemoteStorageTypeServer:
		if fileOpt.IsDownloadDirection() {
			cmd = utils.ShellOpscliDownServer(fileOpt.Api, fileOpt.AesKey, hostLocalfile, fileOpt.RemoteFile)
		} else if fileOpt.IsUploadDirection() {
			cmd = utils.ShellOpscliUploadServer(fileOpt.Api, fileOpt.AesKey, hostLocalfile, fileOpt.RemoteFile)
		}
	case constants.RemoteStorageTypeImage:
		if fileOpt.IsDownloadDirection() {
			cmd = fmt.Sprintf("cp -rbf %s %s", fileOpt.RemoteFile, hostLocalfile)
		}
	}
	if cmd == "" {
		err = errors.New("empty cmd")
		return
	}

	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	volumes, volumeMounts := buildVolumesAndMounts(fileOpt.Mounts)
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
				Labels: map[string]string{
					constants.LabelOpsTaskKey: constants.LabelOpsTaskValue,
				},
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:            "file",
						Image:           fileOpt.RuntimeImage,
						Command:         []string{"bash"},
						Args:            []string{"-c", cmd},
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts:    volumeMounts,
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
				Volumes:       volumes,
			},
		},
		metav1.CreateOptions{},
	)
	return
}

func DownloadS3FileOnNode(client *kubernetes.Clientset, node *v1.Node, namespacedName types.NamespacedName, fileOpt option.FileOption) (pod *corev1.Pod, err error) {
	// find mount path for host root, default to /host if not found
	hostMountPath := "/host"
	for _, mount := range fileOpt.Mounts {
		if mount.HostPath == "/" {
			hostMountPath = mount.MountPath
			break
		}
	}
	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	volumes, volumeMounts := buildVolumesAndMounts(fileOpt.Mounts)
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
				Labels: map[string]string{
					constants.LabelOpsTaskKey: constants.LabelOpsTaskValue,
				},
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:    "file",
						Image:   fileOpt.RuntimeImage,
						Command: []string{"bash"},
						Args: []string{"-c", fmt.Sprintf("opscli file --direction upload"+
							" --endpoint %s --ak %s --sk %s --region %s --bucket %s --localfile %s%s --remotefile s3://%s",
							fileOpt.Endpoint, fileOpt.AK, fileOpt.SK, fileOpt.Region, fileOpt.Bucket, hostMountPath, fileOpt.LocalFile, fileOpt.LocalFile)},
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts:    volumeMounts,
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
				Volumes:       volumes,
			},
		},
		metav1.CreateOptions{},
	)
	return
}
