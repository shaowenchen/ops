### opscli file Command Usage

The `opscli file` command is used for transferring files between the local host, object storage, API servers, and clusters. Below are the details for various use cases.

#### 1. **Host - Local and Object Storage File Transfer**

- **Set AK/SK (Access Key / Secret Key)**

```bash
export ak=<your-ak>
export sk=<your-sk>
```

- **Upload a Local File to Object Storage**

To upload a file `./tmp.log` to the object storage at `s3://logs/tmp.log`:

```bash
/usr/local/bin/opscli file --direction upload --localfile ./tmp.log --remotefile s3://logs/tmp.log --bucket obs-test
```

Here:

- `--bucket` is the S3 bucket name.
- `--region` is the S3 bucket's region.
- `--endpoint` is the S3 bucket's endpoint.
- `--direction` specifies the upload direction.
- `--localfile` is the local file to be uploaded.
- `--remotefile` is the destination file in object storage.

- **Download a File from S3 to Local**

To download `s3://logs/tmp.log` to the local file `./tmp1.log`:

```bash
/usr/local/bin/opscli file --direction download --localfile ./tmp1.log --remotefile s3://logs/tmp.log --bucket obs-test
```

- **Unset AK/SK**

To clear the AK/SK environment variables:

```bash
unset ak
unset sk
```

#### 2. **Host - Local and API Server File Transfer**

> This option provides encryption/decryption for file transfers with the API server.

- **Upload to API Server**

```bash
/usr/local/bin/opscli file --direction upload --api https://gh-uploadapi.chenshaowen.com/api/v1/files --localfile ./tmp.log
```

If `aeskey` is `""`, a random encryption key is generated automatically. If not set, the file is uploaded without encryption.

- **Download from API Server**

```bash
/usr/local/bin/opscli file --api https://gh-uploadapi.chenshaowen.com/api/v1/files --aeskey xxx --direction download --remotefile https://download_url_link.com.aes
```

#### 3. **Cluster - Local and API Server File Transfer**

- **Upload to Cluster's API Server**

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename node1 --direction upload --api https://gh-uploadapi.chenshaowen.com/api/v1/files --aeskey "" --localfile /root/tmp.log --runtimeimage shaowenchen/ops-cli
```

- **Download from Cluster's API Server**

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --api https://gh-uploadapi.chenshaowen.com/api/v1/files --aeskey xxx --localfile /root/tmp1.log --remotefile https://gh-uploadapi.chenshaowen.com/uploadbases/cdn0/raw/1721621949-tmp.log.aes --runtimeimage shaowenchen/ops-cli
```

#### 4. **Cluster - Local and Object Storage File Transfer**

- **Upload to Object Storage from Cluster**

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction upload --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket multimodal --localfile /root/tmp.log --remotefile s3://logs/tmp.log --runtimeimage shaowenchen/ops-cli
```

- **Download from Object Storage to Cluster**

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket multimodal --localfile /root/tmp2.log --remotefile s3://logs/tmp.log --runtimeimage shaowenchen/ops-cli
```

#### 5. **Cluster - Copy Image File to Local**

To copy an image file from the cluster to the local machine:

```bash
/usr/local/bin/opscli file -i ~/.kube/config --nodename xxx --direction download --localfile /root/opscli-copy --remotefile shaowenchen/ops-cli:latest:///usr/local/bin/opscli
```

This command helps copy the executable or file from a container/image inside the cluster to the local machine.

#### 6. **Mount Host Paths to Container**

The `--mount` flag allows you to mount host paths into the container when transferring files. This is useful for accessing host files or directories from within the container.

- **Single Mount**

```bash
opscli file -i ~/.kube/config --nodename node1 \
    --direction upload \
    --localfile /root/tmp.log \
    --remotefile s3://logs/tmp.log \
    --mount /opt/data:/data
```

- **Multiple Mounts**

You can specify multiple mounts by using `--mount` multiple times:

```bash
opscli file -i ~/.kube/config --nodename node1 \
    --direction upload \
    --localfile /root/tmp.log \
    --remotefile s3://logs/tmp.log \
    --mount /opt/data:/data \
    --mount /opt/logs:/logs
```

- **Mount Format**

The mount format is: `hostPath:mountPath`
- `hostPath`: absolute path on the host (required)
- `mountPath`: absolute path in the container (required)

**Note**: If you need to access the host root filesystem for file operations, you can mount it explicitly:
```bash
opscli file --mount /:/host ...
```
