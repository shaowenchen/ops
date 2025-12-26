package task

import (
	"fmt"

	"errors"
	"os"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"gopkg.in/yaml.v3"
)

func GetRealVariables(t *opsv1.Task, taskOpt option.TaskOption) (map[string]string, error) {
	globalVariables := make(map[string]string)
	// cli > env > yaml
	utils.MergeMap(globalVariables, t.Spec.Variables.GetVariables())
	utils.MergeMap(globalVariables, utils.GetAllOsEnv())
	utils.MergeMap(globalVariables, taskOpt.Variables)

	globalVariables = RenderVarsVariables(globalVariables)
	// check variable in task is not empty
	for k, v := range t.Spec.Variables {
		if len(globalVariables[k]) == 0 && v.Required {
			return nil, errors.New("please set variable: " + k)
		}
	}
	return globalVariables, nil
}

func RenderTask(t *opsv1.Task, allVars map[string]string) (*opsv1.Task, error) {
	for i, s := range t.Spec.Steps {
		sp := RenderStepVariables(&s, allVars)
		t.Spec.Steps[i] = *sp
	}
	return t, nil
}

func ReadTaskYaml(filePath string) (tasks []opsv1.Task, err error) {
	fileArray, err := utils.GetFileArray(filePath)
	if err != nil {
		return
	}
	for _, f := range fileArray {
		yfile, err1 := os.ReadFile(f)
		if err1 != nil {
			return nil, err1
		}
		task := opsv1.Task{}
		err = yaml.Unmarshal(yfile, &task)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}
	return
}

func RenderStepVariables(step *opsv1.Step, vars map[string]string) *opsv1.Step {
	return RenderStepVariablesWithPathRefs(step, vars, nil)
}

// RenderStepVariablesWithPathRefs renders step variables with support for path references
// taskResults: map[taskName]map[resultKey]value
func RenderStepVariablesWithPathRefs(step *opsv1.Step, vars map[string]string, taskResults map[string]map[string]string) *opsv1.Step {
	// replace all
	// but in the case, result=${message} will cause error
	// so we need to replace twice
	// this implementation is not good
	f := func() {
		// First resolve path references, then regular variables
		step.Name = RenderStringWithPathRefs(step.Name, vars, taskResults)
		step.Content = RenderStringWithPathRefs(step.Content, vars, taskResults)
		step.LocalFile = RenderStringWithPathRefs(step.LocalFile, vars, taskResults)
		step.RemoteFile = RenderStringWithPathRefs(step.RemoteFile, vars, taskResults)
	}
	f()
	f()

	return step
}

func RenderVarsVariables(vars map[string]string) map[string]string {
	for key := range vars {
		vars[key] = RenderString(vars[key], vars)
	}
	return vars
}

func RenderString(target string, vars map[string]string) string {
	for key, value := range vars {
		if strings.Contains(target, fmt.Sprintf(`${%s}`, key)) {
			target = strings.ReplaceAll(target, fmt.Sprintf(`${%s}`, key), value)
		}
	}
	return target
}

// ExtractVariableReferences extracts variable references from a string
// Returns a set of variable names referenced in the format ${varName}
func ExtractVariableReferences(target string) map[string]bool {
	variables := make(map[string]bool)
	if target == "" {
		return variables
	}

	// Pattern to match ${varName}
	// This regex matches ${...} but excludes path references like ${tasks.xxx.results.yyy} and ${steps.xxx.output}
	start := 0
	for {
		idx := strings.Index(target[start:], "${")
		if idx == -1 {
			break
		}
		idx += start
		// Find the closing }
		endIdx := strings.Index(target[idx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		// Extract the variable reference
		varRef := target[idx+2 : endIdx] // +2 to skip "${"

		// Skip path references (tasks.xxx.results.yyy, steps.xxx.output)
		if !strings.Contains(varRef, ".") {
			variables[varRef] = true
		}

		start = endIdx + 1
	}

	return variables
}

// GetTaskRequiredVariables extracts all variables that a task needs
// This includes:
// 1. Variables defined in task.Spec.Variables
// 2. Variables referenced in step content, when, localfile, remotefile, etc.
func GetTaskRequiredVariables(t *opsv1.Task) map[string]bool {
	requiredVars := make(map[string]bool)

	// Add variables defined in task.Spec.Variables
	for varName := range t.Spec.Variables {
		requiredVars[varName] = true
	}

	// Extract variables from all steps
	for _, step := range t.Spec.Steps {
		// Extract from step content
		for varName := range ExtractVariableReferences(step.Content) {
			requiredVars[varName] = true
		}
		// Extract from step when condition
		for varName := range ExtractVariableReferences(step.When) {
			requiredVars[varName] = true
		}
		// Extract from step localfile
		for varName := range ExtractVariableReferences(step.LocalFile) {
			requiredVars[varName] = true
		}
		// Extract from step remotefile
		for varName := range ExtractVariableReferences(step.RemoteFile) {
			requiredVars[varName] = true
		}
		// Extract from step allowfailure
		for varName := range ExtractVariableReferences(step.AllowFailure) {
			requiredVars[varName] = true
		}
	}

	// Extract from task host (if it's a variable reference)
	if t.Spec.Host != "" && !strings.Contains(t.Spec.Host, "=") {
		// If host is not a label selector, it might be a variable reference
		for varName := range ExtractVariableReferences(t.Spec.Host) {
			requiredVars[varName] = true
		}
	}

	return requiredVars
}

// ResolvePathReference resolves path references like tasks.{taskName}.results.{resultKey}
// Returns the resolved value and true if the reference was found, empty string and false otherwise
func ResolvePathReference(pathRef string, taskResults map[string]map[string]string) (string, bool) {
	// Format: tasks.{taskName}.results.{resultKey}
	if !strings.HasPrefix(pathRef, "tasks.") {
		return "", false
	}

	parts := strings.Split(pathRef, ".")
	if len(parts) != 4 || parts[0] != "tasks" || parts[2] != "results" {
		return "", false
	}

	taskName := parts[1]
	resultKey := parts[3]

	if taskResults == nil {
		return "", false
	}

	if results, ok := taskResults[taskName]; ok {
		if value, ok := results[resultKey]; ok {
			return value, true
		}
	}

	return "", false
}

// RenderStringWithPathRefs renders string with both regular variables and path references
// taskResults: map[taskName]map[resultKey]value
func RenderStringWithPathRefs(target string, vars map[string]string, taskResults map[string]map[string]string) string {
	// First, resolve path references (tasks.{taskName}.results.{resultKey})
	// Use a loop to handle nested references
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		replaced := false
		// Find all path references in the format ${tasks.xxx.results.yyy}
		start := 0
		for {
			idx := strings.Index(target[start:], "${tasks.")
			if idx == -1 {
				break
			}
			idx += start
			// Find the closing }
			endIdx := strings.Index(target[idx:], "}")
			if endIdx == -1 {
				break
			}
			endIdx += idx

			// Extract the path reference
			pathRef := target[idx+2 : endIdx] // +2 to skip "${"
			if value, found := ResolvePathReference(pathRef, taskResults); found {
				target = target[:idx] + value + target[endIdx+1:]
				replaced = true
				start = idx + len(value)
			} else {
				start = endIdx + 1
			}
		}
		if !replaced {
			break
		}
	}

	// Then, resolve regular variables
	return RenderString(target, vars)
}

// ResolveStepReference resolves step references like steps.{stepName}.output
// Returns the resolved value and true if the reference was found, empty string and false otherwise
func ResolveStepReference(stepRef string, stepOutputs map[string]string) (string, bool) {
	// Format: steps.{stepName}.output
	if !strings.HasPrefix(stepRef, "steps.") {
		return "", false
	}

	parts := strings.Split(stepRef, ".")
	if len(parts) != 3 || parts[0] != "steps" || parts[2] != "output" {
		return "", false
	}

	stepName := parts[1]

	if stepOutputs == nil {
		return "", false
	}

	if output, ok := stepOutputs[stepName]; ok {
		return output, true
	}

	return "", false
}

// RenderStringWithStepRefs renders string with both regular variables and step references
// stepOutputs: map[stepName]output
func RenderStringWithStepRefs(target string, vars map[string]string, stepOutputs map[string]string) string {
	// First, resolve step references (steps.{stepName}.output)
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		replaced := false
		// Find all step references in the format ${steps.xxx.output}
		start := 0
		for {
			idx := strings.Index(target[start:], "${steps.")
			if idx == -1 {
				break
			}
			idx += start
			// Find the closing }
			endIdx := strings.Index(target[idx:], "}")
			if endIdx == -1 {
				break
			}
			endIdx += idx

			// Extract the step reference
			stepRef := target[idx+2 : endIdx] // +2 to skip "${"
			if value, found := ResolveStepReference(stepRef, stepOutputs); found {
				target = target[:idx] + value + target[endIdx+1:]
				replaced = true
				start = idx + len(value)
			} else {
				start = endIdx + 1
			}
		}
		if !replaced {
			break
		}
	}

	// Then, resolve regular variables
	return RenderString(target, vars)
}

// RenderStepVariablesWithStepRefs renders step variables with support for step references
// stepOutputs: map[stepName]output
func RenderStepVariablesWithStepRefs(step *opsv1.Step, vars map[string]string, stepOutputs map[string]string) *opsv1.Step {
	// replace all
	// but in the case, result=${message} will cause error
	// so we need to replace twice
	// this implementation is not good
	f := func() {
		// First resolve step references, then regular variables
		step.Name = RenderStringWithStepRefs(step.Name, vars, stepOutputs)
		step.Content = RenderStringWithStepRefs(step.Content, vars, stepOutputs)
		step.LocalFile = RenderStringWithStepRefs(step.LocalFile, vars, stepOutputs)
		step.RemoteFile = RenderStringWithStepRefs(step.RemoteFile, vars, stepOutputs)
	}
	f()
	f()

	return step
}
