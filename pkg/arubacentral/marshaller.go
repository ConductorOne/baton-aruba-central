package arubacentral

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// toStructpbValue converts a native Go type to a structpb.Value.
func toStructpbValue(val interface{}) (*structpb.Value, error) {
	switch v := val.(type) {
	case string:
		return structpb.NewStringValue(v), nil

	case int:
		return structpb.NewNumberValue(float64(v)), nil

	case float64:
		return structpb.NewNumberValue(v), nil

	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
