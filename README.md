# service-watcher

Minimal Kubernetes service watcher intended for scripts.

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

In the sctipt dump all services with `service-watcher -services` and
process the output with [jq](https://stedolan.github.io/jq/).


## Build

```
go get k8s.io/client-go
go get k8s.io/apimachinery
go get github.com/Nordix/service-watcher
go install github.com/Nordix/service-watcher/cmd/...
```

