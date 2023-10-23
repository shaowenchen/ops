## opscli copilot command

### 参数说明

```bash
/usr/local/bin/opscli copilot --help

use llm to assist ops

Usage:
  opscli copilot [flags]

Flags:
  -e, --endpoint string   e.g. https://api.openai.com/v1
  -h, --help              help for copilot
      --history int        (default 5)
  -k, --key string        e.g. sk-xxx
  -m, --model string      e.g. gpt-3.5-turbo
  -s, --silence
  -v, --verbose string
```

copilot 会默认从环境变量获取 `OPENAI_API_HOST`、`OPENAI_API_BASE`、`OPENAI_API_MODEL`、`OPENAI_API_KEY`，如果环境变量不存在，则使用默认值。

-s 设置时，如果涉及代码执行，不需要用户授权二次确认输入 `y`。

### 使用

```bash
/usr/local/bin/opscli copilot

Welcome to Opscli Copilot. Please type "exit" or "q" to quit.
Opscli>
```

- 打开浏览器

```bash
Opscli> 打开浏览器
Open a browser and navigate to 'https://www.google.com'.
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
import webbrowser

webbrowser.open('https://www.google.com')
↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑

Can I run this code? (y/n)
y
```

- 获取 K8s 节点信息

```bash
Opscli> 获取 K8s 节点信息
Retrieve information about Kubernetes nodes
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
kubectl get nodes
↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑

Can I run this code? (y/n)
> y
NAME    STATUS   ROLES                         AGE    VERSION
node1   Ready    control-plane,master,worker   407d   v1.21.0
```
