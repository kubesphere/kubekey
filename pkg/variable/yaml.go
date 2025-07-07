// Package variable provides functionality for handling variables in YAML format.
package variable

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"gopkg.in/yaml.v3"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// YAML tag constants used for parsing different types of nodes
const (
	nullTag      = "!!null"      // Represents null/nil values
	boolTag      = "!!bool"      // Boolean values
	strTag       = "!!str"       // String values
	intTag       = "!!int"       // Integer values
	floatTag     = "!!float"     // Floating point values
	timestampTag = "!!timestamp" // Timestamp values
	seqTag       = "!!seq"       // Sequence/array values
	mapTag       = "!!map"       // Map/object values
	binaryTag    = "!!binary"    // Binary data
	mergeTag     = "!!merge"     // Merge key indicator
)

// parseYamlNode parses a YAML node into a map[string]any.
// It handles both document nodes and other node types.
func parseYamlNode(ctx map[string]any, node yaml.Node) (map[string]any, error) {
	// parse node
	switch node.Kind {
	case yaml.DocumentNode:
		for _, dn := range node.Content {
			if err := processNode(ctx, dn); err != nil {
				return nil, err
			}
		}
	default:
		if err := processNode(ctx, &node); err != nil {
			return nil, err
		}
	}
	var result map[string]any

	return result, errors.Wrap(node.Decode(&result), "failed to decode node to map")
}

// processNode recursively processes a YAML node and updates the context map.
// It handles mapping nodes (objects), sequence nodes (arrays), and scalar nodes (values).
func processNode(ctx map[string]any, node *yaml.Node, path ...string) error {
	switch node.Kind {
	case yaml.MappingNode:
		if len(node.Content)%2 != 0 {
			return errors.New("mapping node has odd number of content nodes")
		}
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			if keyNode.Kind != yaml.ScalarNode {
				return errors.New("map key must be scalar")
			}
			newPath := append(path, mapTag+keyNode.Value)
			if err := processNode(ctx, valueNode, newPath...); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for i, item := range node.Content {
			elemPath := append(path, fmt.Sprintf("%s%d", seqTag, i))
			if err := processNode(ctx, item, elemPath...); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		value, err := parseScalarValue(ctx, node)
		if err != nil {
			return err
		}
		// set context value
		if err := setContextValue(ctx, value, path...); err != nil {
			return err
		}
	default:
		return errors.Errorf("unsupported node kind: %d", node.Kind)
	}

	return nil
}

// parseScalarValue parses a scalar YAML node into its corresponding Go value.
// It handles null, boolean, string, integer, and float values.
func parseScalarValue(ctx map[string]any, node *yaml.Node) (any, error) {
	switch node.Tag {
	case nullTag:
		return nil, nil
	case boolTag:
		v, err := strconv.ParseBool(node.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %q to bool", node.Value)
		}

		return v, nil
	case strTag, "":
		if kkprojectv1.IsTmplSyntax(node.Value) {
			pv, err := tmpl.ParseFunc(ctx, node.Value, func(b []byte) string { return string(b) })
			if err != nil {
				return nil, err
			}
			// change node value
			node.Value = pv

			return pv, nil
		}

		return node.Value, nil
	case intTag:
		v, err := strconv.Atoi(node.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %q to int", node.Value)
		}

		return int64(v), nil
	case floatTag:
		v, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %q to float", node.Value)
		}

		return float64(v), nil
	default:
		return node.Value, nil
	}
}

// setContextValue sets a value in the context map at the specified path.
// The path is a sequence of tags and values that describe the location in the nested structure.
func setContextValue(ctx map[string]any, value any, path ...string) error {
	current := reflect.ValueOf(ctx)
	var parents []reflect.Value
	var keys []any
	for i := range path {
		tag, val := path[i][:5], path[i][5:] // Split into tag and value
		isLast := i == len(path)-1
		// Handle interface values
		current = derefInterface(current)
		var err error
		switch tag {
		case mapTag:
			current, err = handleMap(current, val, isLast, value, &parents, &keys, path, i)
		case seqTag:
			current, err = handleSlice(current, val, isLast, value, &parents, &keys, path, i)
		default:
			return errors.Errorf("unsupported tag: %s", tag)
		}
		if err != nil {
			return err
		}
		if isLast {
			return updateParents(parents, keys, current)
		}
	}
	return nil
}

// handleMap handles setting or creating map values during context updates.
// It manages the creation of new map entries and handles both terminal and non-terminal path segments.
func handleMap(current reflect.Value, key string, isLast bool, value any,
	parents *[]reflect.Value, keys *[]any, path []string, i int) (reflect.Value, error) {
	if current.Kind() != reflect.Map {
		return reflect.Value{}, errors.Errorf("expected map, got %s of path %q", current.Kind(), strings.Join(func() []string {
			out := make([]string, len(path))
			for i, p := range path {
				if len(p) > 5 {
					out[i] = p[5:]
				} else {
					out[i] = ""
				}
			}
			return out
		}(), "."))
	}
	rKey := reflect.ValueOf(key)
	existing := current.MapIndex(rKey)
	if isLast {
		if existing.IsValid() && !isNil(existing) {
			// skip
			return current, nil
		}
		if value == nil {
			current.SetMapIndex(rKey, reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()))
		} else {
			current.SetMapIndex(rKey, reflect.ValueOf(value))
		}
		return current, nil
	}
	// Get or create next value
	var next reflect.Value
	if !existing.IsValid() || isNil(existing) {
		if i+1 >= len(path) {
			return reflect.Value{}, errors.Errorf("path incomplete after index %d", i)
		}
		if strings.HasPrefix(path[i+1], mapTag) {
			next = reflect.ValueOf(make(map[string]any))
		} else if strings.HasPrefix(path[i+1], seqTag) {
			next = reflect.ValueOf([]any{})
		} else {
			next = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
		}
		current.SetMapIndex(rKey, next)
	} else {
		next = derefInterface(existing)
	}
	*parents = append(*parents, current)
	*keys = append(*keys, key)
	return next, nil
}

// handleSlice handles setting or creating slice values during context updates.
// It manages slice growth, element creation, and handles both terminal and non-terminal path segments.
func handleSlice(current reflect.Value, val string, isLast bool, value any,
	parents *[]reflect.Value, keys *[]any, path []string, i int) (reflect.Value, error) {
	// Parse index from path value
	index, err := strconv.Atoi(val)
	if err != nil {
		return reflect.Value{}, errors.Errorf("invalid index %s: %w", val, err)
	}
	if current.Kind() != reflect.Slice {
		return reflect.Value{}, errors.Errorf("expected slice, got %s", current.Kind())
	}
	// Grow slice if requested index is beyond current length
	if index >= current.Len() {
		newLen := index + 1
		newSlice := reflect.MakeSlice(current.Type(), newLen, newLen)
		reflect.Copy(newSlice, current)
		current = newSlice
		if err := updateParents(*parents, *keys, current); err != nil {
			return reflect.Value{}, err
		}
	}
	// Handle setting the final value
	if isLast {
		existingItem := current.Index(index)
		if existingItem.IsValid() && !isNil(existingItem) {
			// skip
			return current, nil
		}
		if value == nil {
			current.Index(index).Set(reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()))
		} else {
			current.Index(index).Set(reflect.ValueOf(value))
		}
		return current, nil
	}
	// Get or initialize nested value
	item := current.Index(index)
	if isNil(item) {
		if i+1 >= len(path) {
			return reflect.Value{}, errors.Errorf("path incomplete after index %d", i)
		}

		var newItem reflect.Value
		switch {
		case strings.HasPrefix(path[i+1], mapTag):
			newItem = reflect.ValueOf(make(map[string]any))
		case strings.HasPrefix(path[i+1], seqTag):
			newItem = reflect.ValueOf([]any{})
		default:
			newItem = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
		}
		current.Index(index).Set(newItem)
		item = newItem
	}
	item = derefInterface(item)
	*parents = append(*parents, current)
	*keys = append(*keys, index)
	return item, nil
}

// derefInterface dereferences an interface value to get its underlying value.
// If the value is not an interface, it returns the original value.
func derefInterface(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		return v.Elem()
	}
	return v
}

// updateParents updates all parent containers after modifying a nested value.
// It walks back up the parent chain, updating each container with the modified value.
func updateParents(parents []reflect.Value, keys []any, value reflect.Value) error {
	for i := len(parents) - 1; i >= 0; i-- {
		parent := parents[i]
		key := keys[i]

		switch parent.Kind() {
		case reflect.Map:
			k := reflect.ValueOf(key)
			parent.SetMapIndex(k, value)
		case reflect.Slice:
			idx, ok := key.(int)
			if !ok {
				return errors.Errorf("expected int key for slice index, got %T", key)
			}
			parent.Index(idx).Set(value)
		default:
			return errors.Errorf("unexpected parent kind: %s", parent.Kind())
		}
		value = parent
	}

	return nil
}

// isNil checks if a reflect.Value is nil, handling interface, map and slice types.
// It returns true if the value is invalid or if it's a nil interface, map, or slice.
func isNil(v reflect.Value) bool {
	return !v.IsValid() || (v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.IsNil()
}
