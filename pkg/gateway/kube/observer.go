package kube

import (
	"fmt"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const resyncTime = 10 * time.Second

type Observer struct {
	gateway *gateway.Gateway

	svcUpdater *svcUpdater
}

func NewObserver(
	gateway *gateway.Gateway,
	kubeconfig string,
	stopper chan struct{}) *Observer {

	o := &Observer{
		gateway:    gateway,
		svcUpdater: newSvcUpdater(gateway),
	}
	config := o.getRestConfig(kubeconfig)

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var factory informers.SharedInformerFactory
	if conf.ConfInst.Gateway.Kubernetes.WatchNamespace != "" {
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientSet,
			resyncTime,
			informers.WithNamespace(conf.ConfInst.Gateway.Kubernetes.WatchNamespace),
		)
	} else {
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientSet,
			resyncTime,
			informers.WithNamespace(corev1.NamespaceAll),
		)
	}
	serviceInformer := factory.Core().V1().Services().Informer()

	defer runtime.HandleCrash()

	// start informer ->
	go factory.Start(stopper)

	// start to sync and call list
	if !cache.WaitForCacheSync(stopper, serviceInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return nil
	}

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.svcUpdater.onAdd, // register add eventhandler
		UpdateFunc: o.svcUpdater.onUpdate,
		DeleteFunc: o.svcUpdater.onDelete,
	})

	return o
}

func (s *Observer) getRestConfig(kubeconfig string) *rest.Config {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config
	}
	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}
