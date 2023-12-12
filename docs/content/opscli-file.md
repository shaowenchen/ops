## opscli file command

### `-i` 指定操作目标清单

- 指定主机

`-i 1.1.1.1`

通过 `--username` 指定用户名，`--password` 指定密码。

- 批量主机

`-i hosts.txt`

```bash
cat hosts.txt

1.1.1.1
2.2.2.2
```

opscli 会从每行中正则匹配 ip 地址，作为目标地址。

- 集群全部节点

```bash
-i ~/.kube/config --all
```

`-i` 默认值为 `~/.kube/config`。

- 集群指定节点

```bash
-i ~/.kube/config --nodename node1
```

node1 为节点名称。

### 本地文件与对象存储文件互传

- 设置 AK\SK

```bash
export ak=
export sk=
```

- 上传本地文件 `./tmp.log` 到对象存储 `s3://logs/tmp.log`

```bash
/usr/local/bin/opscli file --direction upload --localfile ./tmp.log --remotefile s3://logs/tmp.log --bucket obs-test
```

`--bucket` 为 S3 bucket 名称，`--region` 为 S3 bucket 所在区域，`--endpoint` 为 S3 bucket 的 endpoint，`--direction` 为上传方向，`--localfile` 为本地文件，`--remotefile` 为远程文件。

- 下载 S3 `s3://logs/tmp.log` 到本地文件 `./tmp1.log`

```bash
/usr/local/bin/opscli file --direction download --localfile ./tmp1.log --remotefile s3://logs/tmp.log  --bucket obs-test
```

- 清理 AK\SK

```bash
unset ak
unset sk
```

### 本地文件分发到远程主机上

- 上传本地文件 `./tmp.log` 到远程主机 `/tmp/tmp.log`

```bash
/usr/local/bin/opscli file --direction upload --localfile ./tmp.log --remotefile /tmp/tmp.log -i 1.2.3.4 --port 2222 --username root
```

- 下载远程主机 `/tmp/tmp.log` 到本地文件 `./tmp1.log`

```bash
/usr/local/bin/opscli file --direction download --localfile ./tmp1.log --remotefile /tmp/tmp.log -i 1.2.3.4 --port 2222 --username root
```

### 本地文件上传到 API Server，可加密

> 提供本地加解密，与服务器端进行文件传输

- 上传本地文件 `./tmp.log` 到 API Server

```bash
/usr/local/bin/opscli file --direction upload --api https://uploadapi.vinqi.info/api/v1/files --aeskey "" --localfile ./tmp.log

Please use the following command to download the file:
opscli file --api https://uploadapi.vinqi.info/api/v1/files --aeskey xxxxxxxxxxx --direction download --remotefile https://download_url_link.com.aes
```

这里的 api 提供上传服务，aeskey 为空字符串时自动生成一个随机秘钥，如果不设置 aeskey 默认为 unset 将不会进行文件加密。

### 从镜像中提取文件到集群主机
