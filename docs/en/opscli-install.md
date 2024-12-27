### opscli Installation Guide

#### 1. **Single Machine Installation**

- **For Domestic Users (China)**

```bash
curl -sfL https://cf.ghproxy.cc/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

- **For International Users (Outside China)**

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

#### 2. **Batch Installation**

To install `opscli` on multiple hosts, list the IP addresses of all hosts in a `hosts.txt` file, and use the `opscli shell` command to execute the installation. The default credential is the current user's `~/.ssh/id_rsa`.

- **For Domestic Users (China)**

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://cf.ghproxy.cc/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
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

#### 5. **More Information**

To see additional usage options:

```bash
/usr/local/bin/opscli --help
```
