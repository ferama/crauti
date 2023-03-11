package kube

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/ferama/crauti/pkg/gateway/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const resyncTime = 10 * time.Second

type SvcHandler struct {
	server *server.Server

	svcUpdater *svcUpdater
}

func NewSvcHandler(
	server *server.Server,
	stopper chan struct{}) *SvcHandler {

	svc := &SvcHandler{
		server:     server,
		svcUpdater: newSvcUpdater(server),
	}
	config := svc.getRestConfig()

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	factory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		resyncTime,
	)
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

func (s *SvcHandler) getRestConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config
	}
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
