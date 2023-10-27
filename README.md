##### 1. install
###### 1.1 安装apiserver-builder
```shell
curl -SsOL https://github.com/kubernetes-sigs/apiserver-builder-alpha/releases/download/v1.23.0/apiserver-boot-linux-amd64.tar.gz && \
tar -zxvf apiserver-boot-linux-amd64.tar.gz && mv bin/apiserver-boot /usr/local/bin/
```
###### 1.2 构建/部署
```shell
export GOPATH=`pwd`/../
cd github.com/smartcat999/k8s-iaas
go mod tidy

go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 \
&& mkdir ./bin/ \
&& cp $GOPATH/bin/controller-gen ./bin/

make controller-gen \
&& make manifests

# 部署服务
apiserver-boot run in-cluster --image=2030047311/k8s-iaas:0.0.1 --name=k8s-iaas --namespace=default

```
##### 1.3 本地运行
```shell
# 设置GOPATH
export GOPATH=`pwd`/../
cd github.com/smartcat999/k8s-iaas


# 导出service证书
mkdir k8s-iaas.default.svc \
&& kubectl get secrets k8s-iaas -o=jsonpath={.data.tls\\.crt} | base64 -d > k8s-iaas.default.svc/apiserver.crt \
&& kubectl get secrets k8s-iaas -o=jsonpath={.data.tls\\.key} | base64 -d > k8s-iaas.default.svc/apiserver.key

# 构建二进制文件,输出到 ./bin/目录下
apiserver-boot build executables

# 设置KUBECONFIG
export KUBECONFIG=/root/.kube/k3s.yaml

# 本地启动服务,默认监听所有地址
bin/apiserver --kubeconfig=$KUBECONFIG --feature-gates=APIPriorityAndFairness=false \
--authentication-kubeconfig=$KUBECONFIG --authorization-kubeconfig=$KUBECONFIG \
--tls-cert-file=k8s-iaas.default.svc/apiserver.crt \
--tls-private-key-file=k8s-iaas.default.svc/apiserver.key \
--secure-port=8443

# 配置service解析到本地服务
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: k8s-iaas
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 8443
  type: ClusterIP
---
apiVersion: v1
kind: Endpoints
metadata:
 name: k8s-iaas
subsets:
 - addresses:
     - ip: $LOCAL_SERVICE_IP
   ports:
     - port: 8443
EOF
```
###### 1.4 新增crd
```shell
# 在 github.com/smartcat999/k8s-iaas 目录下 新建crd

apiserver-boot create group version resource --group db --version v1alpha1 --kind Instance --non-namespaced=false
```

##### 2 异常
1. cgroups: cgroup mountpoint does not exist: unknown
    ```shell
    mkdir /sys/fs/cgroup/systemd
    mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd
    ```
