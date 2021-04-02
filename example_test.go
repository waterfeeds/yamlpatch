package yamlpatch_test

import (
	"github.com/waterfeeds/yamlpatch"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestApplyPatch(t *testing.T) {
	var input = `
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
`

	var patch = `
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:latest
`

	var want = `apiVersion: apps/v1
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
`

	var n yaml.Node
	if err := yaml.Unmarshal([]byte(input), &n); err != nil {
		t.Fatal(err)
	}

	ops, err := yamlpatch.Compare([]byte(input), []byte(patch))
	if err != nil {
		t.Fatal(err)
	}

	if err = yamlpatch.Apply(&n, ops); err != nil {
		t.Fatal(err)
	}

	var bytes strings.Builder
	e := yaml.NewEncoder(&bytes)
	e.SetIndent(2)
	if err := e.Encode(&n); err != nil {
		t.Fatal(err)
	}

	if want != bytes.String() {
		t.Fatal("yaml patch apply failed")
	}
}