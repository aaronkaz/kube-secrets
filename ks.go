package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretsInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error)
}

type SecretsValues map[string]string
type secretsCache map[string]SecretsValues

type k8sSecretStore struct {
	Namespace string

	mu      sync.Mutex
	secrets SecretsInterface
	cache   secretsCache
}

type StoreOption func(*k8sSecretStore) error

func NewK8sSecretStore(clientSet *kubernetes.Clientset, opts ...StoreOption) (*k8sSecretStore, error) {
	store := &k8sSecretStore{
		cache: make(secretsCache, 0),
	}

	for _, o := range opts {
		if err := o(store); err != nil {
			return nil, err
		}
	}

	// default namespace to self
	if store.Namespace == "" {
		ns, err := SelfNamespace()
		if err != nil {
			return nil, err
		}
		store.Namespace = ns
	}

	store.secrets = clientSet.CoreV1().Secrets(store.Namespace)

	return store, nil
}

func (s *k8sSecretStore) Get(ctx context.Context, secretName, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if values, ok := s.cache[secretName]; ok {
		log.Println("gettin secret from cache...")
		return GetSecretValue(values, key)
	}

	// get and store secret
	secret, err := s.secrets.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	values, err := ParseSecret(secret)
	if err != nil {
		return "", err
	}
	s.cache[secretName] = values

	return GetSecretValue(values, key)
}

func SelfNamespace() (ns string, err error) {
	var fileData []byte
	if fileData, err = os.ReadFile(cKubernetesNamespaceFile); err != nil {
		err = errors.Wrapf(err, "error reading %s; can't get self pod", cKubernetesNamespaceFile)
		return
	}

	ns = strings.TrimSpace(string(fileData))
	return
}

func ParseSecret(secret *v1.Secret) (SecretsValues, error) {
	vs := make(SecretsValues, len(secret.Data))
	for k, v := range secret.Data {
		kb, err := base64.StdEncoding.DecodeString(string(v))
		if err != nil {
			return nil, err
		}
		vs[k] = string(kb)
	}

	return vs, nil
}

func GetSecretValue(vals SecretsValues, key string) (string, error) {
	if v, ok := vals[key]; ok {
		return v, nil
	}

	return "", fmt.Errorf("key `%s` not found in secret", key)
}
