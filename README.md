# kubectl-lazy

## Install

```shell
curl -sSL https://mirror.ghproxy.com/https://raw.githubusercontent.com/togettoyou/kubectl-lazy/main/install.sh | sh
```

Or you can specify the version:

```shell
curl -sSL https://mirror.ghproxy.com/https://raw.githubusercontent.com/togettoyou/kubectl-lazy/main/install.sh | sh -s -- -v 0.0.1
```

## Run

```shell
kubectl lazy
```

Or you can specify kubeconfig:

```shell
kubectl lazy -kubeconfig /root/.kube/config
```

Enable the pprof debug mode:

```shell
kubectl lazy -pprof 8888
```

## Uninstall

```shell
curl -sSL https://mirror.ghproxy.com/https://raw.githubusercontent.com/togettoyou/kubectl-lazy/main/uninstall.sh | sh
```

## Features

- [x] pod infos
- [x] pod events
- [x] pod logs
