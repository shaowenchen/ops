### **Ops Overview**

`Ops` is an operations tool designed to provide a simple and efficient platform for system administrators to complete maintenance tasks quickly. It aims to streamline operations, with a focus on automation and task management across multiple systems and clusters.

### **Production Use Cases**

- **CICD Clusters**: Building 2k+ daily for CICD clusters.
- **Global Clusters**: Managing 40+ clusters overseas.
- **AI Computing Clusters**: Managing 5+ AI computing clusters.
- **Architecture Support**: Supporting both ARM and X86 architectures.

### **Design Overview**

The core components of Ops are built around objects that represent hosts, clusters, and tasks. The design allows for efficient orchestration of operations in a Kubernetes-native environment.

#### **Key Objects:**

- **Host**: Represents machines (cloud-based or bare-metal) that can be accessed via SSH.
- **Cluster**: Represents Kubernetes clusters that can be accessed via `kubectl`.
- **Task**: Represents a combination of multiple files and shell commands.
- **Pipeline**: Represents a sequence of tasks executed in a specific order.

#### **Core Operations:**

- **File**: Uploading and distributing files to hosts and clusters.
- **Shell**: Executing shell scripts on remote hosts or clusters.

### **Components**

1. **ops-cli**: A command-line tool that assists system administrators with automation tasks. It includes a `copilot` subcommand that utilizes LLM (Large Language Models) to automatically trigger Ops tasks and solve issues.
2. **ops-server**: An HTTP service that provides RESTful APIs and a Dashboard interface for managing and monitoring tasks and resources.
3. **ops-controller**: A Kubernetes Operator that manages hosts, clusters, tasks, pipelines, and other resources in the Kubernetes environment.

### **Multi-Cluster Support**

In practice, it is recommended to:

- **Host Creation**: Create hosts based on the current cluster’s machines.
- **Cluster Creation**: Multiple clusters can be added, each cluster will be treated as a managed cluster.

**Task** and **Pipeline** objects are automatically synchronized across all clusters under management, removing the need for manual triggering.

When deploying a pipeline, a **PipelineRun** object is created, which can span multiple clusters. Unlike TaskRuns, PipelineRuns can cross clusters. The `ops-controller` watches the `PipelineRun` object, and based on the `cluster` field, it dispatches the pipeline to the corresponding cluster's controller, which executes the tasks and updates the status of the PipelineRun.

### **Event-Driven Architecture**

Ops adopts an event-driven approach to manage operations:

- **Heartbeat Events**: These events are triggered periodically to check the health of hosts and clusters.
- **Task Execution Events**: Events that are triggered when a TaskRun or PipelineRun task is executed.
- **Inspection Events**: Triggered during scheduled tasks or inspections to monitor the system.
- **Webhook Events**: Custom events like alerts and notifications for maintenance activities.

#### **Event Aggregation in Multi-Cluster Setup**:

- In a multi-cluster environment, each cluster is recommended to install a **Nats** component to collect events from edge clusters.
- Events are then aggregated into one or more clusters to centralize the monitoring and management of operational tasks.

### **System Architecture**

- The overall architecture is built to handle large-scale, multi-cluster environments, ensuring that tasks and pipelines can be seamlessly distributed and monitored across multiple regions and systems.
- The **ops-server** acts as a central hub for managing tasks and pipelines, while the **ops-controller** manages resources in the Kubernetes clusters.

This system is designed to provide comprehensive and automated operations management, facilitating the handling of complex tasks across diverse infrastructures.
