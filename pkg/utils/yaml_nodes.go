package utils

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// CreateValueNode generate yaml node by input anything
func CreateValueNode(v any) (*yaml.Node, error) {
	switch val := v.(type) {
	case string:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: val,
		}, nil
	case int, int8, int16, int32, int64:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!int",
			Value: fmt.Sprintf("%v", val),
		}, nil
	case float32, float64:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!float",
			Value: fmt.Sprintf("%v", val),
		}, nil
	case bool:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: fmt.Sprintf("%v", val),
		}, nil
	case map[string]any:
		node := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		err := addMapContent(node, val)
		return node, err
	case []any:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		err := addSliceContent(node, val)
		return node, err
	default:
		var node yaml.Node
		err := node.Encode(val)
		if err != nil {
			return nil, fmt.Errorf("unsupported type: %T", v)
		}
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			return node.Content[0], nil
		}
		return &node, nil
	}
}

// addMapContent add map value into node
func addMapContent(mapNode *yaml.Node, m map[string]any) error {
	for k, v := range m {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: k,
		}

		valueNode, err := CreateValueNode(v)
		if err != nil {
			return err
		}

		mapNode.Content = append(mapNode.Content, keyNode, valueNode)
	}
	return nil
}

// addSliceContent add slice value into node
func addSliceContent(seqNode *yaml.Node, s []any) error {
	for _, item := range s {
		valueNode, err := CreateValueNode(item)
		if err != nil {
			return err
		}
		seqNode.Content = append(seqNode.Content, valueNode)
	}
	return nil
}
