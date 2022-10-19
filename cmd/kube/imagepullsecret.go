package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var imagePulllSecretOption kube.ImagePulllSecretOption

var imagePulllSecretCmd = &cobra.Command{
	Use:   "imagepulllsecret",
	Short: "config ImagePullSecret to kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		kube.ActionImagePullSecret(imagePulllSecretOption)
	},
}

func init() {
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Kubeconfig, "kubeconfig", "", "", "")
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Namespace, "namespace", "", "", "")
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Name, "name", "", "", "")
	imagePulllSecretCmd.MarkFlagRequired("name")
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Host, "host", "", "", "e.g., https://domain.com,https://domain.com:5000 ")
	imagePulllSecretCmd.MarkFlagRequired("host")
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Username, "username", "", "", "e.g., admin")
	imagePulllSecretCmd.Flags().StringVarP(&imagePulllSecretOption.Password, "password", "", "", "e.g., password")
	imagePulllSecretCmd.Flags().BoolVarP(&imagePulllSecretOption.Clear, "clear", "", false, "")
	imagePulllSecretCmd.Flags().BoolVarP(&imagePulllSecretOption.All, "all", "", false, "")
}
