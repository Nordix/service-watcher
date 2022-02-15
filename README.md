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

In the script dump all services with `service-watcher -services` and
process the output with [jq](https://stedolan.github.io/jq/).


## Build

```
CGO_ENABLED=0 GOOS=linux go build -o service-watcher \
  -ldflags "-extldflags '-static' -X main.version=$(date +%F:%T)" ./cmd/service-watcher
strip service-watcher
```
