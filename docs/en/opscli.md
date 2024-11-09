### **opscli Overview**

`opscli` is a command-line interface (CLI) designed to facilitate batch remote execution of commands, file distribution, and management of Ops Controller CRD resources. It primarily deals with three types of CRDs: **Host**, **Cluster**, and **Task**.

#### **Main Features of opscli**

1. **Batch Remote Command Execution**  
   Allows you to run commands across multiple remote hosts or nodes in a cluster.

2. **Batch File Distribution**  
   Enables bulk file distribution to remote systems.

3. **Ops Controller CRD Resources**  
   Facilitates the creation and management of `Host`, `Cluster`, and `Task` resources in the Ops Controller.

### **Introducing Copilot**

**Ops Copilot** is a sub-command of the `opscli` that integrates Ops capabilities with LLM (Large Language Models). The main goal of Copilot is to improve the user experience and efficiency of managing tasks and their parameters.

#### **Problems Solved by Ops Copilot:**

- **Choosing the Right Task:** Helps in selecting the correct task to solve a particular problem.
- **Selecting Task Parameters:** Guides in determining the appropriate parameters for each task.
- **Unified Entry for Ops and LLM:** Consolidates operations with LLM and Ops into a single entry point, eliminating the need to switch between different tools.
- **Aligning Ops with LLM:** Enhances the alignment of LLM with Ops tasks, making operations smoother.

### **Ops Copilot Architecture**

Ops Copilot uses a **Pipeline** to define and describe workflows. It utilizes LLM to convert text input into executable pipelines with specific parameters. The process works as follows:

1. **Pipeline Definition:** A user describes the desired operation using text.
2. **LLM Processing:** The text is parsed and converted into a corresponding pipeline with the necessary parameters.
3. **Task Execution:** Ops Copilot interacts with the **Ops Server** and **Ops Controller** to create a `TaskRun` and execute it.
4. **Task Execution Monitoring:** The Ops Controller watches for the creation of the `TaskRun`, performs the specified tasks on the relevant `Cluster` or `Host`, and returns the result.

### **Supported Operating Systems**

- **Linux**
- **macOS**

This architecture ensures that Ops tasks can be efficiently automated, with the help of LLM, making the workflow more intelligent and streamlined.
