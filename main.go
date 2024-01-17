package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const cKubernetesNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func main() {
	// make the top level kubernetes API object
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secretStore, err := NewK8sSecretStore(clientset)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	go func() {
		if err := secretStore.Watch(ctx); err != nil {
			panic(err)
		}
	}()

	r := gin.Default()
	r.Any("/", func(c *gin.Context) {
		username, err := secretStore.Get(c, "test-credentials", "username")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		password, err := secretStore.Get(c, "test-credentials", "password")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		r := map[string]string{
			"username": username,
			"password": password,
		}

		c.JSON(http.StatusOK, r)

	})
	r.Run(":9000")
}
