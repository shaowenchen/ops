#!/bin/sh

VERSION="v1.0.0"
filename="opscli-${VERSION}-linux-amd64.tar.gz"

DOWNLOAD_URL="https://github.com/shaowenchen/opscli/releases/download/${VERSION}/opscli-${VERSION}-linux-amd64.tar.gz"
curl -fsLO "$DOWNLOAD_URL"

tar -xzf "${filename}"
mv -f opscli /usr/local/bin/
rm -rf "${filename}"