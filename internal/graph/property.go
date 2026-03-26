package graph

import (
	"fmt"
	"strconv"
)

type PropertyValue struct {
	String *string
	Int    *int64
	Float  *float64
	Bool   *bool
}

func NewStringProperty(v string) PropertyValue {
	return PropertyValue{String: &v}
}

func NewIntProperty(v int64) PropertyValue {
	return PropertyValue{Int: &v}
}

func NewFloatProperty(v float64) PropertyValue {
	return PropertyValue{Float: &v}
}

func NewBoolProperty(v bool) PropertyValue {
	return PropertyValue{Bool: &v}
}

func (p PropertyValue) Type() PropertyType {
	if p.String != nil {
		return PropertyTypeString
	}
	if p.Int != nil {
		return PropertyTypeInt
	}
	if p.Float != nil {
		return PropertyTypeFloat
	}
	if p.Bool != nil {
		return PropertyTypeBool
	}
	return PropertyTypeString
}

func (p PropertyValue) StringValue() string {
	if p.String != nil {
		return *p.String
	}
	return ""
}

func (p PropertyValue) IntValue() int64 {
	if p.Int != nil {
		return *p.Int
	}
	return 0
}

func (p PropertyValue) FloatValue() float64 {
	if p.Float != nil {
		return *p.Float
	}
	return 0
}

func (p PropertyValue) BoolValue() bool {
	if p.Bool != nil {
		return *p.Bool
	}
	return false
}

func ToPropertyValue(v interface{}) PropertyValue {
	switch val := v.(type) {
	case string:
		return NewStringProperty(val)
	case int:
		return NewIntProperty(int64(val))
	case int64:
		return NewIntProperty(val)
	case float64:
		return NewFloatProperty(val)
	case bool:
		return NewBoolProperty(val)
	default:
		return NewStringProperty(fmt.Sprintf("%v", v))
	}
}

func EncodePropertyValue(v PropertyValue) string {
	switch v.Type() {
	case PropertyTypeString:
		return v.StringValue()
	case PropertyTypeInt:
		return strconv.FormatInt(v.IntValue(), 10)
	case PropertyTypeFloat:
		return strconv.FormatFloat(v.FloatValue(), 'f', -1, 64)
	case PropertyTypeBool:
		if v.BoolValue() {
			return "1"
		}
		return "0"
	default:
		return v.StringValue()
	}
}
