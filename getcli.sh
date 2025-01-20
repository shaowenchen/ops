#!/bin/sh

# check dependence
if ! [ -x "$(command -v curl)" ]; then
  echo 'not found curl' >&2
  exit 1
fi

if [ -z "$PROXY" ]; then
  PROXY="https://ghproxy.chenshaowen.com/"
fi

echo "Using PROXY: $PROXY"

# Detect architecture
case "$(uname -m)" in
x86_64)
  ARCH=amd64
  ;;
aarch64)
  ARCH=arm64
  ;;
i386 | i686)
  ARCH=386
  ;;
armv7l)
  ARCH=arm
  GOARM=7
  ;;
*)
  echo "ARCH isn't supported"
  exit 1
  ;;
esac

# Detect OS type
case "$(uname)" in
Linux)
  OSTYPE=linux
  ;;
Darwin)
  OSTYPE=darwin
  ;;
CYGWIN* | MINGW* | MSYS*)
  OSTYPE=windows
  ;;
*)
  echo "OS isn't supported"
  exit 1
  ;;
esac

# remove version prefix 'v'
VERSION=$(echo "$VERSION" | sed 's/^[vV]//')
FILENAME="opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"

# get version
if [ "x${VERSION}" = "x" ] || [ "${VERSION}" = "latest" ]; then
  VERSION="latest"
  DOWNLOAD_URL="https://github.com/shaowenchen/ops/releases/download/${VERSION}/opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
else
  DOWNLOAD_URL="https://github.com/shaowenchen/ops/releases/download/v${VERSION}/opscli-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
  VERSION=v${VERSION}
fi

# download file

http_code=$(curl --connect-timeout 2 -s -o temp.out -w '%{http_code}' https://github.com)
rm -rf temp.out || true

if [ $http_code -ne 302 ]; then
  DOWNLOAD_URL="${PROXY}${DOWNLOAD_URL}"
fi

OPSTEMPDIR=$(mktemp -d)
curl -fsL "$DOWNLOAD_URL" -o "$OPSTEMPDIR/$FILENAME"

if [ ! -f "$OPSTEMPDIR/$FILENAME" ]; then
  echo "Download error."
  exit 1
fi

# install
tar -xzf "$OPSTEMPDIR/$FILENAME" -C "$OPSTEMPDIR"
chmod +x "$OPSTEMPDIR/opscli"

"$OPSTEMPDIR/opscli" version

if [ $? -ne 0 ]; then
  echo "Opscli file error"
  exit 1
fi

OPSDIR="${HOME}/.ops/"
if [ ! -d "${OPSDIR}" ]; then
  mkdir "${OPSDIR}"
fi

if [ -d "${OPSDIR}tasks" ]; then
  mv ${OPSDIR}tasks ${OPSDIR}.tasks_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSTEMPDIR/tasks" ${OPSDIR}

if [ -d "${OPSDIR}taskruns" ]; then
  mv ${OPSDIR}taskruns ${OPSDIR}.taskruns_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSTEMPDIR/taskruns" ${OPSDIR}

if [ -d "${OPSDIR}pipelines" ]; then
  mv ${OPSDIR}pipelines ${OPSDIR}.pipelines_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSTEMPDIR/pipelines" ${OPSDIR}

if [ -d "${OPSDIR}eventhooks" ]; then
  mv ${OPSDIR}eventhooks ${OPSDIR}.eventhooks_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSTEMPDIR/eventhooks" ${OPSDIR}

if [ $(id -u) -eq 0 ]; then
  mv -f "$OPSTEMPDIR/opscli" /usr/local/bin/
  echo "Congratulations! Opscli live in /usr/local/bin/opscli"
else
  mv -f "$OPSTEMPDIR/opscli" $(pwd)
  echo "Congratulations! Please run 'sudo mv $(pwd)/opscli /usr/local/bin/' to install."
fi

# clear
rm -rf "$OPSTEMPDIR/$FILENAME"
