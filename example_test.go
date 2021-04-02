package yamlpatch_test

import (
	"github.com/waterfeeds/yamlpatch"
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

	applied, err := yamlpatch.ApplyPatch(input, patch)
	if err != nil {
		t.Fatal(err)
	}

	if want != applied {
		t.Fatal("yaml patch apply failed")
	}
}