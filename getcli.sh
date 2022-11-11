#!/bin/sh

# check dependence
if ! [ -x "$(command -v curl)" ]; then
  echo 'not found curl' >&2
  exit 1
fi

case "$(uname -m)" in
  x86_64)
    ARCH=amd64
    ;;
  *)
    echo "ARCH isn't supported"
    exit 1
    ;;
esac

case "$(uname)" in
  Linux)
    OSTYPE=linux
    ;;
  Darwin)
    OSTYPE=darwin
    ;;
  *)
    echo "OS isn't supported"
    exit 1
    ;;
esac

# get version
if [ "x${VERSION}" = "x" ]; then
  VERSION="latest"
fi

# download file
FILENAME="opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/shaowenchen/ops/releases/download/${VERSION}/opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"

http_code=$(curl --connect-timeout 3 -s -o temp.out -w '%{http_code}' https://github.com)
rm -rf temp.out || true

if [ $http_code -ne 302 ]; then
  DOWNLOAD_URL="https://ghproxy.com/${DOWNLOAD_URL}"
fi

rm -rf opscli* || true
curl -fsLO "$DOWNLOAD_URL"

if [ ! -f "${FILENAME}" ]; then
  echo "Download error."
  exit 1
fi

# install
tar -xzf "${FILENAME}"
chmod +x opscli

`pwd`/opscli version

if [ $? -ne 0 ]; then
    echo "Opscli file error"
    exit 1
fi

OPSDIR="${HOME}/.ops/"
if [ ! -d "${OPSDIR}" ]; then
  mkdir "${OPSDIR}"
fi

if [ -d "${OPSDIR}task" ]; then
  mv ${OPSDIR}task ${OPSDIR}.task_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv task ${OPSDIR}

if [ `id -u` -eq 0 ]; then
  mv -f opscli /usr/local/bin/
  echo "Congratulations! Opscli live in /usr/local/bin/opscli"
else
  echo "Congratulations! Please run 'sudo mv `pwd`/opscli /usr/local/bin/' to install."
fi

# clear
rm -rf "${FILENAME}"
