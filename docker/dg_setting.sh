#!/bin/sh
# Setting gin-fork

set -x
set -e

# Replace Setting valuse on custom/conf/app.ini
sed -i "s/DG_GIT_API_TOKEN/$DG_GIT_API_TOKEN/g" /data/gogs/custom/conf/app.ini
sed -i "s/DATABESE_PASSWORD/$DATABESE_PASSWORD/g" /data/gogs/custom/conf/app.ini