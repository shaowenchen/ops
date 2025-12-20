## opscli 安装

### 安装

- 单机安装

国内使用:

```bash
PROXY=https://ghfast.top/
curl -sfL $PROXY/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest PROXY=$PROXY sh -
```

国外使用:

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

- 批量安装

将需要安装的全部主机 ip 都写入到 `hosts.txt` 文件中，然后使用 `opscli shell` 命令批量安装，凭证默认为当前用户的 `~/.ssh/id_rsa`。

国内使用:

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://ghproxy.chenshaowen.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

国外使用:

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

### 版本升级

- 单机

```bash
sudo /usr/local/bin/opscli upgrade
```

- 批量

```bash
/usr/local/bin/opscli shell --content "sudo /usr/local/bin/opscli upgrade" -i hosts.txt
```

## 自动补全

### bash

```bash
echo 'source <(opscli completion bash)' >>~/.bashrc
```

### zsh

```bash
echo 'source <(opscli completion zsh)' >>~/.zshrc
```

## 配置

`opscli` 支持通过 `config` 命令进行配置管理，允许您设置和管理在所有 CLI 命令中使用的配置值。

**配置文件位置**

配置文件存储在 `~/.ops/opscli/config`（YAML 格式）。

**支持的配置项**

- **proxy**: 网络请求的代理 URL（例如：`https://ghfast.top/`）
- **runtimeimage**: Kubernetes 任务的默认运行时镜像（例如：`ubuntu:22.04`）

**配置命令**

- **设置配置**: `opscli config set <key> <value>`
  ```bash
  opscli config set proxy https://ghfast.top/
  opscli config set runtimeimage ubuntu:22.04
  ```

- **获取配置**: `opscli config get <key>`
  ```bash
  opscli config get proxy
  ```

- **列出所有配置**: `opscli config list`
  ```bash
  opscli config list
  # 输出：
  # proxy = https://ghfast.top/
  # runtimeimage = (not set)
  ```

- **删除配置**: `opscli config unset <key>`
  ```bash
  opscli config unset proxy
  ```

**配置优先级**

配置值遵循以下优先级顺序（从高到低）：

1. **CLI 参数**（最高优先级）
   - 命令行标志，如 `--proxy` 或 `--runtimeimage`
   - 示例：`opscli task --filepath task.yaml --proxy https://cli-proxy.com`

2. **环境变量**
   - `PROXY`: 代理 URL
   - `DEFAULT_RUNTIME_IMAGE`: 默认运行时镜像
   - 示例：`export PROXY=https://env-proxy.com`

3. **配置文件**（`~/.ops/opscli/config`）
   - 通过 `opscli config set` 设置的值
   - 示例：`opscli config set proxy https://config-proxy.com`

4. **默认值**（最低优先级）
   - 内置默认值
   - Proxy: `https://ghproxy.chenshaowen.com/`
   - Runtime Image: `ubuntu:22.04`

**使用示例**

```bash
# 示例 1: 使用配置文件
opscli config set proxy https://ghfast.top/
opscli upgrade --manifests  # 自动使用配置文件中的 proxy

# 示例 2: 使用环境变量覆盖
export PROXY=https://env-proxy.com
opscli upgrade --manifests  # 使用环境变量

# 示例 3: 使用 CLI 参数覆盖（最高优先级）
opscli upgrade --proxy https://cli-proxy.com  # 使用 CLI 参数
```

## 更多

```bash
/usr/local/bin/opscli --help
```
