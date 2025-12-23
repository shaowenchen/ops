## Ops-controller-manager Pipeline 对象

`Pipeline` 对象定义了一系列按顺序执行的任务。Pipeline 中的任务可以通过路径引用使用前面任务的结果。

### 使用 yaml 文件创建

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Pipeline
metadata:
  name: app-deploy
  namespace: ops-system
spec:
  desc: "部署应用到生产环境"
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

说明：

- **`desc`**: Pipeline 描述
- **`variables`**: Pipeline 级别的变量
- **`tasks`**: 按顺序执行的任务列表：
  - **`name`**: 任务在 Pipeline 中的名称（用于路径引用）
  - **`taskRef`**: 引用的 Task 对象
  - **`results`**: 结果键到 step 名称的映射，定义要导出的 step 输出
  - **`variables`**: 任务特定变量（由 controller 自动填充）。如果任务变量有默认值且 Pipeline 中有同名变量，会自动填充到这里。前面任务的结果也会自动作为变量可用

**自动变量填充：**

当 Pipeline 被创建或更新时，controller 会自动为每个任务填充 `variables`：
- 如果任务变量有默认值且 Pipeline 中有同名变量，会将 Pipeline 的值填充到 `TaskRef.variables` 中
- 这样所有变量值都会在 Pipeline spec 中明文展示

### 变量引用

前面任务的结果会自动添加到 Pipeline 变量中，可以直接引用：

```yaml
variables:
  config: ${config}      # 直接引用（推荐）
  version: ${version}    # 直接引用（推荐）
```

也可以使用路径语法进行显式引用：

```yaml
variables:
  config: ${tasks.prepare-config.results.config}   # 路径引用
  version: ${tasks.prepare-config.results.version} # 路径引用
```

### 查看对象

```bash
kubectl get pipeline -n ops-system
```

