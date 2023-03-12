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

type SvcHandler struct {
	server *gateway.Server

	svcUpdater *svcUpdater
}

func NewSvcHandler(
	server *gateway.Server,
	kubeconfig string,
	stopper chan struct{}) *SvcHandler {

	svc := &SvcHandler{
		server:     server,
		svcUpdater: newSvcUpdater(server),
	}
	config := svc.getRestConfig(kubeconfig)

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	var factory informers.SharedInformerFactory
	if conf.Config.Kubernetes.WatchNamespace != "" {
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientSet,
			resyncTime,
			informers.WithNamespace(conf.Config.Kubernetes.WatchNamespace),
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
		AddFunc:    svc.onAdd, // register add eventhandler
		UpdateFunc: svc.onUpdate,
		DeleteFunc: svc.onDelete,
	})

	return svc
}

func (s *SvcHandler) getRestConfig(kubeconfig string) *rest.Config {
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

func (s *SvcHandler) onAdd(obj interface{}) {
	svc := obj.(*corev1.Service)
	s.svcUpdater.add(fmt.Sprintf("%s/%s", svc.Namespace, svc.Name), *svc.DeepCopy())
}

func (s *SvcHandler) onUpdate(oldObj interface{}, newObj interface{}) {
	// oldSvc := oldObj.(*corev1.Service)
	newSvc := newObj.(*corev1.Service)
	s.svcUpdater.add(fmt.Sprintf("%s/%s", newSvc.Namespace, newSvc.Name), *newSvc.DeepCopy())
}

func (s *SvcHandler) onDelete(obj interface{}) {
	svc := obj.(*corev1.Service)
	s.svcUpdater.delete(fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
}
