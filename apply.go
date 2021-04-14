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

	if err = apply(&n, ops); err != nil {
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

	if err := apply(&n, ops); err != nil {
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
func apply(n *yaml.Node, ops []PatchOperation) error {
	for _, operation := range ops {
		switch operation.Op {
		case "add":
			if err := applyAdd(n, operation); err != nil {
				return fmt.Errorf("apply add error: %w", err)
			}
		case "replace":
			if err := applyReplace(n, operation); err != nil {
				return fmt.Errorf("apply replace error: %w", err)
			}
		default:
		}
	}
	return nil
}

func applyAdd(n *yaml.Node, o PatchOperation) error {
	if o.Value.IsZero() {
		return fmt.Errorf("missing value in patch (op=%s)", o.Op)
	}

	targetPath, err := compileParentPath(o)
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

	var addNodeKey, addNodeVal yaml.Node
	addNodeKey.Encode(findAbsPath(o.JSONPointer))
	addNodeVal.Encode(o.Value)

	// append add node to parent's content
	nodes[0].Content = append(nodes[0].Content, &addNodeKey, &addNodeVal)

	return nil
}

func applyReplace(n *yaml.Node, o PatchOperation) error {
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

func compileParentPath(o PatchOperation) (*yamlpath.Path, error) {
	parentPath := findParentPath(o.JSONPointer)
	jsonPath := jsonpath.FromJSONPointer(parentPath)

	compiled, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON Path (path=%s) -> (jsonpath=%s): %w", o.JSONPointer, jsonPath, err)
	}

	return compiled, nil
}

func findParentPath(path string) string {
	pathArr := strings.Split(path, "/")
	if len(pathArr) <= 1 {
		return ""
	}

	return strings.Join(pathArr[:len(pathArr)-1], "/")
}

func findAbsPath(path string) string {
	pathArr := strings.Split(path, "/")
	if len(pathArr) < 1 {
		return ""
	}

	return pathArr[len(pathArr)-1]
}

func compilePath(o PatchOperation) (*yamlpath.Path, error) {
	jsonPath := jsonpath.FromJSONPointer(o.JSONPointer)

	compiled, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON Path (path=%s) -> (jsonpath=%s): %w", o.JSONPointer, jsonPath, err)
	}

	return compiled, nil
}