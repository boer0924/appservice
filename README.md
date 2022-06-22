
```shell
kubebuilder init --domain boer.xyz --repo github.com/boerlabs/appservice

kubebuilder create api --group app --version v1beta1 --kind AppService

make
make manifests

make generate

make install

make run

kubectl apply -f config/samples/

make docker-build docker-push IMG=<some-registry>/<project-name>:tag
make docker-build docker-push IMG=registry.cn-beijing.aliyuncs.com/boer/appservice-operator:1.1.1

make deploy IMG=<some-registry>/<project-name>:tag
make deploy IMG=registry.cn-beijing.aliyuncs.com/boer/appservice-operator:1.1.1

kubectl delete -f config/samples/
make uninstall
make undeploy
```

- https://kubernetes.github.io/ingress-nginx/examples/rewrite/
- https://github.com/kubernetes-sigs/application