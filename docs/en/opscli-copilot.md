### opscli copilot Command Usage

The `opscli copilot` command assists in managing your operations with LLM (Large Language Model) integration. Below are the details for the available flags, environment variables, and example commands.

#### Flags

```bash
Usage:
  opscli copilot [flags]

Flags:
  -e, --endpoint string   e.g. https://api.openai.com/v1
  -h, --help              help for copilot
      --history int        (default 5)
  -k, --key string        e.g. sk-xxx
  -m, --model string      e.g. gpt-3.5-turbo
  -s, --silence           Suppress output
  -v, --verbose string    Verbose mode
```

#### Environment Variables

`copilot` will automatically fetch the following environment variables if set. If not, it will use default values:

- `OPENAI_API_KEY` (API key for OpenAI)
- `OPENAI_API_HOST` (Host URL for OpenAI API)
- `OPENAI_API_BASE` (Base URL for OpenAI API)
- `OPENAI_API_MODEL` (Model type for OpenAI API)
- `OPS_SERVER` (Server URL for ops server)
- `OPS_TOKEN` (Token for authentication)

You can set these environment variables as follows:

```bash
export OPENAI_API_KEY=sk-xxxx
export OPENAI_API_HOST=https://llmapi.YOUR-OPENAI-SERVER.com/v1
export OPS_SERVER=http://1.1.1.1
export OPS_TOKEN=xxxx
```

#### Running Copilot

To run Copilot:

```bash
/usr/local/bin/opscli copilot
```

You will be prompted with:

```
Welcome to Opscli Copilot. Please type "exit" or "q" to quit.
Opscli>
```

#### Available Operations

To check available operations:

```bash
Opscli> What operations are available?
Here are the available operations and their descriptions:

1. list-cluster: Query the list of Kubernetes clusters.
2. list-task: Query the list of tasks.
3. list-pipeline: Query the list of pipelines.
4. restart-pod: Restart or delete a Pod. Variables: podname (one or more Pod names).
5. force-restart-pod: Force restart or delete a Pod. Variables: podname (one or more Pod names).
6. get-cluster-ip: Query the IP addresses of clusters. Variables: clusterip (one or more cluster IP addresses).
7. clear-disk: Clear disk space. Variables: nodeName (one or more node names).
```

#### Querying Clusters

To list clusters:

```bash
Opscli> What clusters are available?
The following clusters are available:
1. ops-system/xx-xx: A cluster deployed on xxx cloud for 88 inference cluster.
2. ops-system/xx-xx: A cluster deployed on xxx for the 119 cluster.
3. ops-system/xx-xx: A cluster deployed on xxx for integrated training and inference.
4. ops-system/xx-xx: A cluster deployed on xxx for NPU training.
```

#### Restarting a Pod

To force restart a pod:

```bash
Opscli> Force restart the pod ubuntu-8474647969-qszcj in the training-inference cluster
```

The process will look like this:

1. Check if the pod exists:

   - Output: "Pod ubuntu-8474647969-qszcj found in default namespace."

2. Delete the pod:
   - Output: "Warning: Force delete, no confirmation waiting for the pod to terminate. The resource may run indefinitely on the cluster."
   - Output: "Pod 'ubuntu-8474647969-qszcj' has been force deleted."

In your cluster, you can watch the pod status with:

```bash
kubectl get pod ubuntu-8474647969-qszcj -w
```

The output will show the pod terminating and restarting:

```
NAME                      READY   STATUS    RESTARTS   AGE
ubuntu-8474647969-qszcj   1/1     Running   0          20h
ubuntu-8474647969-qszcj   1/1     Terminating   0          20h
ubuntu-8474647969-qszcj   1/1     Terminating   0          20h
```
