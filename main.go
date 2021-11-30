package main

import (
	"flag"
	"fmt"
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
	pprof := flag.Uint("pprof", 0, "the value is port (for example, 33060), enabling pprof debug mode")

	flag.Parse()

	if *pprof != 0 {
		go func() { http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *pprof), nil) }()
	}

	if err := NewUi(NewClient(*kubeconfig)).Run(); err != nil {
		panic(err)
	}
}
