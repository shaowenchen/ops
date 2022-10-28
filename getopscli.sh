#!/bin/sh

# check dependence
if ! command -v curl &> /dev/null
then
    echo "could't find curl"
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
DOWNLOAD_URL="https://github.com/shaowenchen/opscli/releases/download/${VERSION}/opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"

http_code=$(curl --connect-timeout 3 -s -o temp.out -w '%{http_code}' ${DOWNLOAD_URL})
rm -rf temp.out || true

if [ $http_code -ne 302 ]; then
    DOWNLOAD_URL="https://ghproxy.com/${DOWNLOAD_URL}"
fi

curl -fsLO "$DOWNLOAD_URL"

if [ ! -f "${FILENAME}" ]; then
   echo "Download error."
   exit 1
fi

# install
if [ -d "pipeline" ]; then
  mv pipeline .pipeline_$(date +%F_%R)
fi
tar -xzf "${FILENAME}"
chmod +x opscli

if [ `id -u` -ne 0 ]; then
  mv -f opscli /usr/local/bin/
  /usr/local/bin/opscli version
  echo "Congratulations! Opscli live in /usr/local/bin/opscli"
else
  `pwd`/opscli version
  echo "Congratulations! Opscli live in `pwd`opscli"
fi

# clear
rm -rf "${FILENAME}"
