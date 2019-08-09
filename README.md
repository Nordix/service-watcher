# service-watcher

Minimal Kubernetes service watcher intended for scripts. You *can* use
`kubectl` for this but you probably don't want it in a pod.

Test using default `$HOME/.kube/`;
```
service-watcher -services | jq .
```

Normal use is to use `service-watcher` to call a script whenever any
service is updated;

```
service-watcher -watch -script /path/to/myscript
...
```

In the script dump all services with `service-watcher -services` and
process the output with [jq](https://stedolan.github.io/jq/).


## Build

```
go get k8s.io/client-go/...
go get k8s.io/apimachinery/...
go get github.com/Nordix/service-watcher
CGO_ENABLED=0 GOOS=linux go install -a \
  -ldflags "-extldflags '-static' -X main.version=$(date +%F:%T)" \
  github.com/Nordix/service-watcher/cmd/service-watcher
strip $GOPATH/bin/service-watcher
```

