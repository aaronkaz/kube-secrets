package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	r := gin.Default()
	r.Any("/", func(c *gin.Context) {
		ns, err := selfNamespace()
		if err != nil {
			c.AbortWithError(500, err)
		}

		r := make(map[string]interface{})

		r["namespace"] = ns

		secrets := clientset.CoreV1().Secrets(ns)
		kubesecretList, err := secrets.List(c, metav1.ListOptions{})
		if err != nil {
			c.AbortWithError(500, err)
		}
		ss := make([]string, len(kubesecretList.Items))
		for i, kubesecret := range kubesecretList.Items {
			ss[i] = kubesecret.Name
		}

		r["secrets"] = ss
		c.JSON(http.StatusOK, r)

	})
	r.Run(":9000")
}

func selfNamespace() (ns string, err error) {
	var fileData []byte
	if fileData, err = os.ReadFile(cKubernetesNamespaceFile); err != nil {
		err = errors.Wrapf(err, "error reading %s; can't get self pod", cKubernetesNamespaceFile)
		return
	}

	ns = strings.TrimSpace(string(fileData))
	return
}
