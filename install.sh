#!/bin/bash
set -e

version="0.0.2"

while [ -n "$1" ]
do
  case "$1" in
    -v)
        version=$2
        ;;
  esac
  shift
done

goos="linux"
goarch="amd64"

if [ `uname -s` == "Darwin" ];then
	goos="darwin"
fi

if [[ `arch` =~ "aarch64" ]];then
	goarch="arm64"
fi

filename="kubectl-lazy_"$version"_"$goos"_"$goarch".tar.gz"

rm -f $filename

curl -LJO "https://mirror.ghproxy.com/https://github.com/togettoyou/kubectl-lazy/releases/download/v"$version"/"$filename""

tar -xvf $filename

chmod +x ./kubectl-lazy

mv ./kubectl-lazy /usr/local/bin

kubectl plugin list