#!/bin/sh
if [ `id -u` -ne 0 ]; then
  echo "please run with root"
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
    OS=$(cat /etc/os-release 2>/dev/null | grep ^ID= | awk -F= '{print $2}')
    case "$OS" in
      alpine)
          apk add coreutils
          ;;
      *)
          ;;
    esac
    ;;
  Darwin)
    OSTYPE=darwin
    ;;
  *)
    echo "OS isn't supported"
    exit 1
    ;;
esac

if [ "x${VERSION}" = "x" ]; then
  VERSION="$(curl -sL https://api.github.com/repos/shaowenchen/opscli/releases |
    grep -o 'download/v[0-9]*.[0-9]*.[0-9]*/' |
    sort --version-sort |
    tail -1 | awk -F'/' '{ print $2}')"
  VERSION="${VERSION##*/}"
fi

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
fi

if [ -d "pipeline"]; then
  mv pipeline .pipeline_$(date +%F_%R)
fi

tar -xzf "${FILENAME}"
chmod +x opscli
mv -f opscli /usr/local/bin/
rm -rf "${FILENAME}"
echo "Congratulations! Opscli live in /usr/local/bin/opscli"
/usr/local/bin/opscli version