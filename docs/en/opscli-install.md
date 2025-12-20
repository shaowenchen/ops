### opscli Installation Guide

#### 1. **Single Machine Installation**

- **For Domestic Users (China)**

```bash
PROXY=https://ghfast.top/
curl -sfL $PROXY/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest PROXY=$PROXY sh -
```

- **For International Users (Outside China)**

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

#### 2. **Batch Installation**

To install `opscli` on multiple hosts, list the IP addresses of all hosts in a `hosts.txt` file, and use the `opscli shell` command to execute the installation. The default credential is the current user's `~/.ssh/id_rsa`.

- **For Domestic Users (China)**

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://ghproxy.chenshaowen.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

- **For International Users (Outside China)**

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

#### 3. **Version Upgrade**

- **For Single Machine**

```bash
sudo /usr/local/bin/opscli upgrade
```

- **For Batch Upgrade**

```bash
/usr/local/bin/opscli shell --content "sudo /usr/local/bin/opscli upgrade" -i hosts.txt
```

#### 4. **Auto-completion Setup**

- **For bash**

```bash
echo 'source <(opscli completion bash)' >>~/.bashrc
```

- **For zsh**

```bash
echo 'source <(opscli completion zsh)' >>~/.zshrc
```

#### 5. **Configuration**

`opscli` supports configuration management through the `config` command, allowing you to set and manage configuration values that are used across all CLI commands.

**Configuration File Location**

Configuration is stored in `~/.ops/opscli/config` (YAML format).

**Supported Configuration Keys**

- **proxy**: Proxy URL for network requests (e.g., `https://ghfast.top/`)
- **runtimeimage**: Default runtime image for Kubernetes tasks (e.g., `ubuntu:22.04`)

**Configuration Commands**

- **Set configuration**: `opscli config set <key> <value>`
  ```bash
  opscli config set proxy https://ghfast.top/
  opscli config set runtimeimage ubuntu:22.04
  ```

- **Get configuration**: `opscli config get <key>`
  ```bash
  opscli config get proxy
  ```

- **List all configurations**: `opscli config list`
  ```bash
  opscli config list
  # Output:
  # proxy = https://ghfast.top/
  # runtimeimage = (not set)
  ```

- **Unset configuration**: `opscli config unset <key>`
  ```bash
  opscli config unset proxy
  ```

**Configuration Priority**

Configuration values follow a priority order (highest to lowest):

1. **CLI Arguments** (highest priority)
   - Command-line flags like `--proxy` or `--runtimeimage`
   - Example: `opscli task --filepath task.yaml --proxy https://cli-proxy.com`

2. **Environment Variables**
   - `PROXY`: Proxy URL
   - `DEFAULT_RUNTIME_IMAGE`: Default runtime image
   - Example: `export PROXY=https://env-proxy.com`

3. **Configuration File** (`~/.ops/opscli/config`)
   - Values set via `opscli config set`
   - Example: `opscli config set proxy https://config-proxy.com`

4. **Default Values** (lowest priority)
   - Built-in defaults
   - Proxy: `https://ghproxy.chenshaowen.com/`
   - Runtime Image: `ubuntu:22.04`

**Usage Examples**

```bash
# Example 1: Using configuration file
opscli config set proxy https://ghfast.top/
opscli upgrade --manifests  # Automatically uses proxy from config

# Example 2: Override with environment variable
export PROXY=https://env-proxy.com
opscli upgrade --manifests  # Uses environment variable

# Example 3: Override with CLI argument (highest priority)
opscli upgrade --proxy https://cli-proxy.com  # Uses CLI argument
```

#### 6. **More Information**

To see additional usage options:

```bash
/usr/local/bin/opscli --help
```
