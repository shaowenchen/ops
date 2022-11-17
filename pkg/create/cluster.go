package create

import (
	"context"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"

	"k8s.io/client-go/rest"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateHost(logger *opslog.Logger, restConfig *rest.Config, host *opsv1.Host, clear bool) (err error) {

	client, err := opskube.GetOpsClient(logger, restConfig)
	if err != nil {
		return
	}
	if clear {
		err = client.Delete(context.TODO(), host)
	} else {
		err = client.Create(context.TODO(), host)
	}
	return
}

func CreateCluster(logger *opslog.Logger, restConfig *rest.Config, cluster *opsv1.Cluster, clear bool) (err error) {
	cluster.Spec.Server = restConfig.Host
	client, err := opskube.GetOpsClient(logger, restConfig)
	if err != nil {
		return
	}
	if clear {
		err = client.Delete(context.TODO(), cluster)
	} else {
		err = client.Create(context.TODO(), cluster)
	}
	return
}

func CreateTask(logger *opslog.Logger, restConfig *rest.Config, t *opsv1.Task, clear bool) (err error) {
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}

	client, err := runtimeClient.New(restConfig, runtimeClient.Options{Scheme: scheme})
	if err != nil {
		return
	}
	if clear {
		err = client.Delete(context.TODO(), t)
	} else {
		err = client.Create(context.TODO(), t)
	}
	return
}
