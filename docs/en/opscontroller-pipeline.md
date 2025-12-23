### **Ops-controller-manager Pipeline Object**

The `Pipeline` object defines a sequence of tasks executed in a specific order. Tasks in a pipeline can reference results from previous tasks using path references.

#### **Create Pipeline Using YAML File**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Pipeline
metadata:
  name: app-deploy
  namespace: ops-system
spec:
  desc: "Deploy application to production"
  variables:
    namespace:
      value: "production"
  tasks:
    - name: prepare-config
      taskRef: prepare-config-task
      results:
        config: generate-step
        version: generate-step
    - name: deploy-app
      taskRef: deploy-app-task
      variables:
        config: ${config}
        version: ${version}
    - name: verify-deploy
      taskRef: verify-deploy-task
      variables:
        version: ${version}
```

In this YAML:

- **`desc`**: A description of the pipeline.
- **`variables`**: Pipeline-level variables.
- **`tasks`**: List of tasks to execute in order:
  - **`name`**: Task name in the pipeline (used for path references).
  - **`taskRef`**: Reference to the Task object.
  - **`results`**: Map of result keys to step names, defining which step outputs to export.
  - **`variables`**: Task-specific variables (automatically filled by controller). If a task variable has a default value and the pipeline has the same variable, it will be automatically filled here. Results from previous tasks are also automatically available as variables.

**Automatic Variable Filling:**

When a Pipeline is created or updated, the controller automatically fills `variables` for each task:
- If a task variable has a default value and the pipeline has the same variable, the pipeline value is filled into `TaskRef.variables`
- This makes all variable values visible in the Pipeline spec

#### **Variable References**

Results from previous tasks are automatically added to Pipeline variables, so you can reference them directly:

```yaml
variables:
  config: ${config}      # Direct reference (recommended)
  version: ${version}    # Direct reference (recommended)
```

You can also use path syntax for explicit references:

```yaml
variables:
  config: ${tasks.prepare-config.results.config}   # Path reference
  version: ${tasks.prepare-config.results.version} # Path reference
```

#### **View Pipeline Object**

```bash
kubectl get pipeline -n ops-system
```

