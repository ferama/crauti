package certcache

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type SecretCache struct {
	clientSet *kubernetes.Clientset
}

const (
	// TODO: read from
	// /var/run/secrets/kubernetes.io/serviceaccount/namespace
	// ?
	namespace    = "test"
	secretMapKey = "cert"
)

func NewSecretCache(kubeconfig string) *SecretCache {
	s := &SecretCache{}

	config := s.getRestConfig(kubeconfig)

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	s.clientSet = clientSet
	return s
}

func (s *SecretCache) Get(ctx context.Context, key string) ([]byte, error) {
	var (
		data   []byte
		err    error
		done   = make(chan struct{})
		secret *corev1.Secret
	)

	go func() {
		secret, err = s.clientSet.CoreV1().
			Secrets(namespace).
			Get(context.TODO(), key, metav1.GetOptions{})

		data = secret.Data[secretMapKey]
		close(done)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}

	return data, err
}

func (s *SecretCache) Put(ctx context.Context, key string, data []byte) error {
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			secretMapKey: data,
		},
		Type: corev1.SecretTypeOpaque,
	}

	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)

		_, err = s.clientSet.CoreV1().
			Secrets(namespace).
			Create(context.TODO(), &secret, metav1.CreateOptions{})

	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return err
}

func (s *SecretCache) Delete(ctx context.Context, key string) error {
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)

		err = s.clientSet.CoreV1().
			Secrets(namespace).
			Delete(context.TODO(), key, metav1.DeleteOptions{})

	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return err
}

func (s *SecretCache) getRestConfig(kubeconfig string) *rest.Config {
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
