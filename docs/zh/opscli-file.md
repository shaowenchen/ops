## opscli file command

### 主机 - 本地与对象存储互传文件

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

### 主机 - 本地与 API Server 互传文件

> 提供本地加解密，与服务器端进行文件传输

- 上传

```bash
/usr/local/bin/opscli file --direction upload --api https://uploadapi.vinqi.com/api/v1/files --aeskey "" --localfile ./tmp.log

Please use the following command to download the file:
opscli file --api https://uploadapi.vinqi.com/api/v1/files --aeskey xxx --direction download --remotefile https://download_url_link.com.aes
```

这里的 api 提供上传服务，aeskey 为空字符串时自动生成一个随机秘钥，如果不设置 aeskey 默认为 unset 将不会进行文件加密。

- 下载

```bash
/usr/local/bin/opscli file --api https://uploadapi.vinqi.com/api/v1/files --aeskey xxx --direction download --remotefile https://download_url_link.com.aes
```

### 集群 - 本地与 API Server 互传文件

- 上传

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename node1 --direction upload --api https://uploadapi.vinqi.com/api/v1/files --aeskey "" --localfile /root/tmp.log --runtimeimage shaowenchen/ops-cli
```

- 下载

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --api https://uploadapi.vinqi.com/api/v1/files --aeskey xxx --localfile /root/tmp1.log --remotefile https://uploadapi.vinqi.com/uploadbases/cdn0/raw/1721621949-tmp.log.aes --runtimeimage shaowenchen/ops-cli
```

### 集群 - 本地与对象存储互传文件

- 上传

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction upload --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket multimodal  --localfile /root/tmp.log --remotefile s3://logs/tmp.log --runtimeimage shaowenchen/ops-cli
```

- 下载

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket multimodal  --localfile /root/tmp2.log --remotefile s3://logs/tmp.log --runtimeimage shaowenchen/ops-cli
```

### 集群 - 镜像文件拷贝到本地

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --localfile /root/opscli-copy --remotefile shaowenchen/ops-cli:latest:///usr/local/bin/opscli
```
