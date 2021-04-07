package yamlpatch

import (
	"fmt"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"github.com/waterfeeds/yamlpatch/jsonpath"
	"gopkg.in/yaml.v3"
	"strings"
)

func ApplyPatch(a, b string) (string, error) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(a), &n); err != nil {
		return "", err
	}

	ops, err := Compare([]byte(a), []byte(b))
	if err != nil {
		return "", err
	}

	if err = applyOnNode(&n, ops); err != nil {
		return "", err
	}

	var bytes strings.Builder
	e := yaml.NewEncoder(&bytes)
	e.SetIndent(2)
	if err = e.Encode(&n); err != nil {
		return "", err
	}

	return bytes.String(), nil
}

func Apply(yamlS string, ops []PatchOperation) (string, error) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(yamlS), &n); err != nil {
		return "", err
	}

	if err := applyOnNode(&n, ops); err != nil {
		return "", err
	}

	var bytes strings.Builder
	e := yaml.NewEncoder(&bytes)
	e.SetIndent(2)
	if err := e.Encode(&n); err != nil {
		return "", err
	}

	return bytes.String(), nil
}

// Apply applies the set of operations to the YAML node in order.
func applyOnNode(n *yaml.Node, ops []PatchOperation) error {
	for _, operation := range ops {
		if operation.Op != "replace" {
			continue
		}
		if err := apply(n, operation); err != nil {
			return fmt.Errorf("yamlpatch.Apply error: %w", err)
		}
	}
	return nil
}

func apply(n *yaml.Node, o PatchOperation) error {
	if o.Value.IsZero() {
		return fmt.Errorf("missing value in patch (op=%s)", o.Op)
	}

	targetPath, err := compilePath(o)
	if err != nil {
		return fmt.Errorf("invalid patch: %w", err)
	}
	nodes, err := targetPath.Find(n)
	if err != nil {
		return fmt.Errorf("could not find the path in YAML: %w", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("any node did not match (path=%s, jsonpath=%s)", o.JSONPointer, o.JSONPath)
	}
	for _, node := range nodes {
		node.Kind = o.Value.Kind
		node.Style = o.Value.Style
		node.Value = o.Value.Value
		node.Tag = o.Value.Tag
		node.Content = o.Value.Content
	}
	return nil
}

func compilePath(o PatchOperation) (*yamlpath.Path, error) {
	if o.JSONPath != "" && o.JSONPointer != "" {
		return nil, fmt.Errorf("do not set both path and jsonpath (jsonpointer=%s, jsonpath=%s)", o.JSONPointer, o.JSONPath)
	}

	if o.JSONPath != "" {
		compiled, err := yamlpath.NewPath(o.JSONPath)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON Path (path=%s): %w", o.JSONPath, err)
		}
		return compiled, nil
	}

	jsonPath := jsonpath.FromJSONPointer(o.JSONPointer)
	compiled, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON Path (path=%s) -> (jsonpath=%s): %w", o.JSONPointer, jsonPath, err)
	}
	return compiled, nil
}