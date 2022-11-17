package kube

import (
	"context"
	"fmt"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func SyncUpdateStatus() (err error) {
	for range time.Tick(10 * time.Second) {
		fmt.Println("sync update cluster status")
		scheme, err := opsv1.SchemeBuilder.Build()
		if err != nil {
			fmt.Println(err)
			return err
		}
		restConfig, err := utils.GetRestConfig(constants.GetCurrentUserConfigPath())
		if err != nil {
			fmt.Println(err)
			return err
		}
		opsClient, err := runtimeClient.New(restConfig, runtimeClient.Options{Scheme: scheme})
		if err != nil {
			fmt.Println(err)
			return err
		}
		cluserList := &opsv1.ClusterList{}
		err = opsClient.List(context.TODO(), cluserList)
		if err != nil {
			fmt.Println(err)
			return err
		}
		for _, cluster := range cluserList.Items {
			kc, err := NewClusterConnection(&cluster)
			if err != nil {
				fmt.Println(err)
				continue
			}
			s, err := kc.GetStatus()
			if err != nil {
				fmt.Println(err)
				continue
			}
			cluster.Status = *s
			err = opsClient.Status().Update(context.TODO(), &cluster)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
	return
}
