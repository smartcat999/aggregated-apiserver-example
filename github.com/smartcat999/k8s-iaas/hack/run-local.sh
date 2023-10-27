#!/usr/bin/env bash

NODE_IP=$(kubectl get nodes -o jsonpath="{.items[0].metadata.annotations.flannel\\.alpha\\.coreos\\.com\\/public-ip}")
export GOPATH=`pwd`/../../../../

mkdir k8s-aggregated.default.svc \
&& kubectl get secrets k8s-aggregated -o=jsonpath={.data.tls\\.crt} | base64 -d > k8s-aggregated.default.svc/apiserver.crt \
&& kubectl get secrets k8s-aggregated -o=jsonpath={.data.tls\\.key} | base64 -d > k8s-aggregated.default.svc/apiserver.key

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: k8s-aggregated
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
 name: k8s-aggregated
subsets:
 - addresses:
     - ip: $NODE_IP
   ports:
     - port: 8443
EOF


go run cmd/apiserver/main.go --kubeconfig=/root/.kube/k3s.yaml --feature-gates=APIPriorityAndFairness=false \
--authentication-kubeconfig=/root/.kube/k3s.yaml \
--authorization-kubeconfig=/root/.kube/k3s.yaml \
--tls-cert-file=k8s-aggregated.default.svc/apiserver.crt \
--tls-private-key-file=k8s-aggregated.default.svc/apiserver.key --secure-port=8443
