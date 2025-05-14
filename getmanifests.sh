#!/bin/sh

# check dependence
if ! [ -x "$(command -v curl)" ]; then
  echo 'not found curl' >&2
  exit 1
fi

if [ -z "$PROXY" ]; then
  PROXY="https://ghproxy.chenshaowen.com/"
fi

echo "Using Proxy: $PROXY"

DOWNLOAD_URL="https://github.com/shaowenchen/ops-manifests/archive/refs/tags/latest.tar.gz"
FILENAME="ops-manifests-latest.tar.gz"
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


OPSMANIFESTSDIR=$OPSTEMPDIR/ops-manifests-latest

OPSDIR="${HOME}/.ops/"
if [ ! -d "${OPSDIR}" ]; then
  mkdir "${OPSDIR}"
fi

if [ -d "${OPSDIR}tasks" ]; then
  mv ${OPSDIR}tasks ${OPSDIR}.tasks_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSMANIFESTSDIR/tasks" ${OPSDIR}

if [ -d "${OPSDIR}taskruns" ]; then
  mv ${OPSDIR}taskruns ${OPSDIR}.taskruns_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSMANIFESTSDIR/taskruns" ${OPSDIR}

if [ -d "${OPSDIR}pipelines" ]; then
  mv ${OPSDIR}pipelines ${OPSDIR}.pipelines_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSMANIFESTSDIR/pipelines" ${OPSDIR}

if [ -d "${OPSDIR}eventhooks" ]; then
  mv ${OPSDIR}eventhooks ${OPSDIR}.eventhooks_upgrade_$(date +%Y-%m-%d-%H-%M-%S)
fi

mv "$OPSMANIFESTSDIR/eventhooks" ${OPSDIR}

echo "Congratulations! Ops manifests has been upgraded in ${OPSDIR}"

# clear
rm -rf "$OPSTEMPDIR"
