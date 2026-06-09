package v1

import "testing"

func TestTaskRunMergeVariablesKeepsTaskRunPriority(t *testing.T) {
	tr := &TaskRun{
		Spec: TaskRunSpec{
			Variables: map[string]string{
				"from_task_value": "taskrun",
			},
		},
	}
	task := &Task{
		Spec: TaskSpec{
			Variables: Variables{
				"from_task_value":   {Value: "task-yaml"},
				"from_task_default": {Default: "task-default"},
			},
		},
	}

	tr.MergeVariables(task)

	if got := tr.Spec.Variables["from_task_value"]; got != "taskrun" {
		t.Fatalf("TaskRun variable should override Task YAML value, got %q", got)
	}
	if got := tr.Spec.Variables["from_task_default"]; got != "task-default" {
		t.Fatalf("Task YAML should fill missing TaskRun variable, got %q", got)
	}
}
