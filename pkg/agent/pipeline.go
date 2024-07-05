package agent

import (
	"strings"
)

type LLMPipeline struct {
	Desc      string         `json:"desc"`
	Namespace string         `json:"namespace"`
	Name      string         `json:"name"`
	NodeName  string         `json:"nodeName"`
	TypeRef   string         `json:"typeRef"`
	NameRef   string         `json:"nameRef"`
	Variables []VariablePair `json:"variablePairs"`
	LLMTasks  []LLMTask      `json:"llmTasks"`
}

func (p LLMPipeline) GetFullVariables() []VariablePair {
	// maybe need to find from LLMTasks
	vars := make([]VariablePair, 0)
	for _, v := range p.Variables {
		vars = append(vars, v)
	}
	// add extra variables
	vars = append(vars, VariablePair{
		Key:          "namespace",
		DefaultValue: UniversalValue{Type: UniversalValueTypeString, Str: p.Namespace},
		Required:     true,
		Desc:         "the namespace of the pipeline",
	})

	vars = append(vars, VariablePair{
		Key:          "typeRef",
		DefaultValue: UniversalValue{Type: UniversalValueTypeString, Str: VarTypeRefValue},
		Required:     true,
	})

	vars = append(vars, VariablePair{
		Key:      "nameRef",
		Required: true,
	})

	vars = append(vars, VariablePair{
		Key:          "nodeName",
		DefaultValue: UniversalValue{Type: UniversalValueTypeString, Str: p.NodeName},
		Required:     false,
	})
	return vars
}

type VariablePair struct {
	Key          string
	DefaultValue UniversalValue
	Value        UniversalValue
	Desc         string
	Regx         string
	Required     bool
	Enum         []string
	Examples     []UniversalValue
}

type UniversalValue struct {
	Str   string
	Array []string
	Type  UniversalValueType
}

func (u *UniversalValue) String() string {
	if u == nil {
		return ""
	}
	if u.Type == UniversalValueTypeString {
		return u.Str
	}
	if u.Type == UniversalValueTypeArray {
		return strings.Join(u.Array, ",")
	}
	return ""
}

type UniversalValueType string

const (
	UniversalValueTypeString UniversalValueType = "string"
	UniversalValueTypeArray  UniversalValueType = "array"
)

// Object  DataType = "object"
// Number  DataType = "number"
// Integer DataType = "integer"
// String  DataType = "string"
// Array   DataType = "array"
// Null    DataType = "null"
// Boolean DataType = "boolean"

func NewString(s string) *UniversalValue {
	return &UniversalValue{Type: UniversalValueTypeString, Str: s}
}

func NewList(list []string) *UniversalValue {
	return &UniversalValue{Type: UniversalValueTypeArray, Array: list}
}
