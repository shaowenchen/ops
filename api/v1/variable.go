/*
Copyright 2024 shaowenchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	"regexp"
)

type Variable struct {
	Default  string   `json:"default,omitempty" yaml:"default,omitempty"`
	Display  string   `json:"display,omitempty" yaml:"display,omitempty"`
	Value    string   `json:"value,omitempty" yaml:"value,omitempty"`
	Desc     string   `json:"desc,omitempty" yaml:"desc,omitempty"`
	Regex    string   `json:"regex,omitempty" yaml:"regex,omitempty"`
	Required bool     `json:"required,omitempty" yaml:"required,omitempty"`
	Enums    []string `json:"enums,omitempty" yaml:"enums,omitempty"`
	Examples []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}

func (v Variable) GetValue() string {
	if v.Value != "" {
		return v.Value
	}
	return v.Default
}

func (v Variable) MergeLowPriorityVariable(others Variable) Variable {
	if v.Default == "" {
		v.Default = others.Default
	}
	if v.Display == "" {
		v.Display = others.Display
	}
	if v.Value == "" {
		v.Value = others.Value
	}
	if v.Desc == "" {
		v.Desc = others.Desc
	}
	if v.Required == false {
		v.Required = others.Required
	}
	if v.Regex == "" {
		v.Regex = others.Regex
	}
	if len(v.Enums) == 0 {
		v.Enums = others.Enums
	}
	if len(v.Examples) == 0 {
		v.Examples = others.Examples
	}
	return v
}

func (v Variable) MergeHighPriorityVariable(others Variable) Variable {
	if others.Default != "" {
		v.Default = others.Default
	}
	if others.Display != "" {
		v.Display = others.Display
	}
	if others.Value != "" {
		v.Value = others.Value
	}
	if others.Desc != "" {
		v.Desc = others.Desc
	}
	if others.Required != false {
		v.Required = others.Required
	}
	if others.Regex != "" {
		v.Regex = others.Regex
	}
	if len(others.Enums) > 0 {
		v.Enums = others.Enums
	}
	if len(others.Examples) > 0 {
		v.Examples = others.Examples
	}
	return v
}

func (v *Variable) Validate() bool {
	//validate required
	if v.Required && v.Value == "" {
		return false
	}
	//validate enum
	if len(v.Enums) > 0 {
		found := false
		for _, e := range v.Enums {
			if v.Value == e {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	//validate regex
	if v.Regex != "" {
		re, err := regexp.Compile(v.Regex)
		if err != nil {
			return false
		}
		if !re.MatchString(v.Value) {
			return false
		}
	}
	return true
}

type Variables map[string]Variable

func (objs Variables) MergeLowPriorityVariables(others Variables) Variables {
	for k, v := range objs {
		if _, ok := others[k]; !ok {
			continue
		}
		// merge others
		objs[k] = v.MergeLowPriorityVariable(others[k])
	}
	return objs
}

func (objs Variables) MergeHighPriorityVariables(others Variables) Variables {
	for k, v := range objs {
		if _, ok := others[k]; !ok {
			continue
		}
		// merge others
		objs[k] = v.MergeHighPriorityVariable(others[k])
	}
	return objs
}

func (objs Variables) GetVariables() (result map[string]string) {
	result = make(map[string]string)
	for k, v := range objs {
		if v.Value != "" {
			result[k] = v.Value
		} else {
			result[k] = v.Default
		}
	}
	return
}
