package certcache

import (
	"context"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// This kubernetes secrets based certifacte store backend
// BE VERY CAREFULL: do not use zerlog logging here. I don't actually
// know the reason, but a single call to the logger will break
// the functionlity of this object. Very Very Weird!
// The issue seems to affect zerlog only. If you need to put out
// some text for debugging purposes, please use go logger
type SecretCache struct {
	clientSet *kubernetes.Clientset
	namespace string
}

const (
	secretNamePrefix = "crauti-"
	secretMapKey     = "cert"
)

func NewSecretCache(kubeconfig string) *SecretCache {

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}

	s := &SecretCache{
		namespace: string(data),
	}

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
	// log.Printf("get for %s", key)
	var (
		data   []byte
		err    error
		done   = make(chan struct{})
		secret *corev1.Secret
	)

	go func() {
		secret, err = s.clientSet.CoreV1().
			Secrets(s.namespace).
			Get(ctx, s.keyToName(key), metav1.GetOptions{})

		data = secret.Data[secretMapKey]
		close(done)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}
	if err != nil {
		return nil, autocert.ErrCacheMiss
	}
	return data, err
}

func (s *SecretCache) Put(ctx context.Context, key string, data []byte) error {
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)

		select {
		case <-ctx.Done():
		default:
			secret := corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.keyToName(key),
					Namespace: s.namespace,
				},
				Data: map[string][]byte{
					secretMapKey: data,
				},
				Type: corev1.SecretTypeOpaque,
			}
			_, err = s.clientSet.CoreV1().
				Secrets(s.namespace).
				Create(ctx, &secret, metav1.CreateOptions{})
		}

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
			Secrets(s.namespace).
			Delete(ctx, s.keyToName(key), metav1.DeleteOptions{})

	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *SecretCache) keyToName(key string) string {
	key = strings.ReplaceAll(key, ".", "-")
	key = strings.ReplaceAll(key, "+", "-")
	key = strings.ReplaceAll(key, "_", "-")
	key = secretNamePrefix + key
	return key
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
