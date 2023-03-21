#!/bin/sh
# Finalize the build

set -x
set -e

# Create git user for Gogs
addgroup --system git --gid 2000
adduser --ingroup git --uid 2000 --no-create-home --disabled-password --gecos 'Gogs Git User' git --home /data/git -shell /bin/bash && usermod --password '*' git && passwd --unlock git
echo "export GOGS_CUSTOM=${GOGS_CUSTOM}" >> /etc/profile

# Final cleaning
rm -rf /app/gogs/build
rm /app/gogs/docker/finalize.sh
rm /app/gogs/docker/finalize_ginfork.sh
rm /app/gogs/docker/start.sh
rm /app/gogs/docker/nsswitch.conf
rm /app/gogs/docker/README.md
rm -rf /app/gogs/docker/Dockefile*
# /app/gogs/docker/　配下のファイルの削除について検証する必要がある。
