package yamlpatch

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
)

var ErrBadYamlDoc = fmt.Errorf("invalid yaml Document")

func newPatch(operation, path string, value interface{}) PatchOperation {
	var n yaml.Node
	n.Encode(value)

	return PatchOperation{Op: operation, JSONPointer: path, Value: n}
}

// 'a' is original, 'b' is the modified document. Both are to be given as json encoded content.
// The function will return an array of PatchOperations
//
// An error will be returned if any of the two documents are invalid.
func Compare(a, b []byte) ([]PatchOperation, error) {
	var aI interface{}
	var bI interface{}

	err := yaml.Unmarshal(a, &aI)
	if err != nil {
		return nil, ErrBadYamlDoc
	}
	err = yaml.Unmarshal(b, &bI)
	if err != nil {
		return nil, ErrBadYamlDoc
	}

	return handleValues(aI, bI, "", []PatchOperation{})
}

// Returns true if the values matches (must be json types)
// The types of the values must match, otherwise it will always return false
// If two map[string]interface{} are given, all elements must match.
func matchesValue(av, bv interface{}) bool {
	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
		return false
	}
	switch at := av.(type) {
	case string:
		bt := bv.(string)
		if bt == at {
			return true
		}
	case int:
		bt := bv.(int)
		if bt == at {
			return true
		}
	case float64:
		bt := bv.(float64)
		if bt == at {
			return true
		}
	case bool:
		bt := bv.(bool)
		if bt == at {
			return true
		}
	case map[string]interface{}:
		bt := bv.(map[string]interface{})
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	case []interface{}:
		bt := bv.([]interface{})
		if len(bt) != len(at) {
			return false
		}
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	}
	return false
}

// From http://tools.ietf.org/html/rfc6901#section-4 :
//
// Evaluation of each reference token begins by decoding any escaped
// character sequence.  This is performed by first transforming any
// occurrence of the sequence '~1' to '/', and then transforming any
// occurrence of the sequence '~0' to '~'.
//   TODO decode support:
//   var rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")

var rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")

func makePath(path string, newPart interface{}) string {
	key := rfc6901Encoder.Replace(fmt.Sprintf("%v", newPart))
	if path == "" {
		return "/" + key
	}
	if strings.HasSuffix(path, "/") {
		return path + key
	}

	return path + "/" + key
}

// diffYaml returns the (recursive) difference between a and b as an array of PatchOperations.
func diffYaml(a, b map[interface{}]interface{}, path string, patch []PatchOperation) ([]PatchOperation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			patch = append(patch, newPatch("add", p, bv))
			continue
		}
		// If types have changed, replace completely
		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
			patch = append(patch, newPatch("replace", p, bv))
			continue
		}
		// Types are the same, compare values
		var err error
		patch, err = handleValues(av, bv, p, patch)
		if err != nil {
			return nil, err
		}
	}
	// Now add all deleted values as nil
	for key := range a {
		_, found := b[key]
		if !found {
			p := makePath(path, key)

			patch = append(patch, newPatch("remove", p, nil))
		}
	}
	return patch, nil
}

// diff returns the (recursive) difference between a and b as an array of PatchOperations.
func diff(a, b map[string]interface{}, path string, patch []PatchOperation) ([]PatchOperation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			patch = append(patch, newPatch("add", p, bv))
			continue
		}
		// If types have changed, replace completely
		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
			patch = append(patch, newPatch("replace", p, bv))
			continue
		}
		// Types are the same, compare values
		var err error
		patch, err = handleValues(av, bv, p, patch)
		if err != nil {
			return nil, err
		}
	}
	// Now add all deleted values as nil
	for key := range a {
		_, found := b[key]
		if !found {
			p := makePath(path, key)

			patch = append(patch, newPatch("remove", p, nil))
		}
	}
	return patch, nil
}

func handleValues(av, bv interface{}, p string, patch []PatchOperation) ([]PatchOperation, error) {
	var err error
	switch at := av.(type) {
	case map[string]interface{}:
		bt := bv.(map[string]interface{})
		patch, err = diff(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case map[interface{}]interface{}:
		bt := bv.(map[interface{}]interface{})
		patch, err = diffYaml(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case string, float64, bool, int:
		if !matchesValue(av, bv) {
			patch = append(patch, newPatch("replace", p, bv))
		}
	case []interface{}:
		bt, ok := bv.([]interface{})
		if !ok {
			// array replaced by non-array
			patch = append(patch, newPatch("replace", p, bv))
		} else if len(at) != len(bt) {
			// arrays are not the same length
			patch = append(patch, newPatch("replace", p, bv))

		} else {
			for i := range bt {
				patch, err = handleValues(at[i], bt[i], makePath(p, i), patch)
				if err != nil {
					return nil, err
				}
			}
		}
	case nil:
		switch bv.(type) {
		case nil:
			// Both nil, fine.
		default:
			patch = append(patch, newPatch("add", p, bv))
		}
	default:
		panic(fmt.Sprintf("Unknown type:%T ", av))
	}
	return patch, nil
}

// compareArray generates remove and add operations for `av` and `bv`.
func compareArray(av, bv []interface{}, p string) []PatchOperation {
	retval := []PatchOperation{}

	// Find elements that need to be removed
	processArray(av, bv, func(i int, value interface{}) {
		retval = append(retval, newPatch("remove", makePath(p, i), nil))
	})

	// Find elements that need to be added.
	// NOTE we pass in `bv` then `av` so that processArray can find the missing elements.
	processArray(bv, av, func(i int, value interface{}) {
		retval = append(retval, newPatch("add", makePath(p, i), value))
	})

	return retval
}

// processArray processes `av` and `bv` calling `applyOp` whenever a value is absent.
// It keeps track of which indexes have already had `applyOp` called for and automatically skips them so you can process duplicate objects correctly.
func processArray(av, bv []interface{}, applyOp func(i int, value interface{})) {
	foundIndexes := make(map[int]struct{}, len(av))
	reverseFoundIndexes := make(map[int]struct{}, len(av))
	for i, v := range av {
		for i2, v2 := range bv {
			if _, ok := reverseFoundIndexes[i2]; ok {
				// We already found this index.
				continue
			}
			if reflect.DeepEqual(v, v2) {
				// Mark this index as found since it matches exactly.
				foundIndexes[i] = struct{}{}
				reverseFoundIndexes[i2] = struct{}{}
				break
			}
		}
		if _, ok := foundIndexes[i]; !ok {
			applyOp(i, v)
		}
	}
}

