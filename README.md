##### 1. install
###### 1.1 安装apiserver-builder
```shell
curl -SsOL https://github.com/kubernetes-sigs/apiserver-builder-alpha/releases/download/v1.23.0/apiserver-boot-linux-amd64.tar.gz && \
tar -zxvf apiserver-boot-linux-amd64.tar.gz && mv bin/apiserver-boot /usr/local/bin/
```
###### 1.2 构建/部署
```shell
export GOPATH=`pwd`/../
cd github.com/smartcat999/k8s-aggregated
go mod tidy

go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 \
&& mkdir ./bin/ \
&& cp $GOPATH/bin/controller-gen ./bin/

make controller-gen \
&& make manifests

# 部署服务
apiserver-boot run in-cluster --image=2030047311/k8s-aggregated:0.0.1 --name=k8s-aggregated --namespace=default
```
###### 1.3 异常
1. cgroups: cgroup mountpoint does not exist: unknown
    ```shell
    mkdir /sys/fs/cgroup/systemd
    mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd
    ```
