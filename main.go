package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	go func() { http.ListenAndServe("0.0.0.0:6060", nil) }()

	if err := NewUi(NewClient(*kubeconfig)).Run(); err != nil {
		panic(err)
	}
}
