# yamlpatch

This is a command line tool to apply a JSON Patch to a YAML Document preserving position and comments.


## Features

- Support both JSON Pointer and JSON Path (depends on [vmware-labs/yaml-jsonpath](https://github.com/vmware-labs/yaml-jsonpath))
- Passed the [conformance tests](https://github.com/json-patch/json-patch-tests) of JSON Patch
- Single binary

**Note**: currently only `op=replace` mode is implemented

## Thanks for the contribution of these projects
https://github.com/int128/yamlpatch  
https://github.com/mattbaird/jsonpatch

## Getting Started

TODO: install

### Example: Replace a field in Kubernetes YAML

Input:

```yaml
# https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: nginx
          image: nginx
```

Patch:

```yaml
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:latest
```

ApplyPatch

```yaml
# https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:latest
```