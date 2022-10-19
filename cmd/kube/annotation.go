package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var annotateOption kube.AnnotateOption

var annotationCmd = &cobra.Command{
	Use:   "annotate",
	Short: "annotate pod for kubernetes",
	Run: func(cmd *cobra.Command, args []string){
		kube.ActionAnnotate(annotateOption)
	},
}

func init() {
	annotationCmd.Flags().StringVarP(&annotateOption.Kubeconfig, "kubeconfig", "", "", "")
	annotationCmd.Flags().StringVarP(&annotateOption.Namespace, "namespace", "", "", "")
	annotationCmd.Flags().StringVarP(&annotateOption.Type, "type", "", "", "velero")
	annotationCmd.MarkFlagRequired("type")
	annotationCmd.Flags().BoolVarP(&annotateOption.Clear, "clear", "", false, "")
	annotationCmd.Flags().BoolVarP(&annotateOption.All, "all", "", false, "")
}
