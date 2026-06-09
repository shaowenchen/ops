package task

import (
	"testing"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/option"
)

func TestGetRealVariablesPriority(t *testing.T) {
	t.Setenv("OPS_TEST_PRIORITY_ENV", "env")
	t.Setenv("from_env", "env")

	task := &opsv1.Task{
		Spec: opsv1.TaskSpec{
			Variables: opsv1.Variables{
				"from_default": {Default: "yaml"},
				"from_env":     {Default: "yaml"},
				"from_cli":     {Default: "yaml"},
			},
		},
	}

	vars, err := GetRealVariables(task, option.TaskOption{
		DefaultVariables: map[string]string{
			"from_default":          "code-default",
			"OPS_TEST_PRIORITY_ENV": "code-default",
		},
		Variables: map[string]string{
			"from_cli": "cli",
		},
	})
	if err != nil {
		t.Fatalf("GetRealVariables returned error: %v", err)
	}

	if got := vars["from_default"]; got != "yaml" {
		t.Fatalf("YAML should override code default, got %q", got)
	}
	if got := vars["OPS_TEST_PRIORITY_ENV"]; got != "env" {
		t.Fatalf("env should override code default, got %q", got)
	}
	if got := vars["from_env"]; got != "yaml" {
		t.Fatalf("YAML should keep its own value, got %q", got)
	}
	if got := vars["from_cli"]; got != "cli" {
		t.Fatalf("CLI should override YAML, got %q", got)
	}
}
