#!/bin/sh
# Setting gin-fork

set -x
set -e

# Replace Setting valuse on custom/conf/app.ini

sed -i "s/DB_HOST/$DB_HOST/g" /data/gogs/custom/conf/app.ini
sed -i "s/DB_NAME/$DB_NAME/g" /data/gogs/custom/conf/app.ini
sed -i "s/DB_USER/$DB_USER/g" /data/gogs/custom/conf/app.ini
sed -i "s/DB_PASSWORD/$DB_PASSWORD/g" /data/gogs/custom/conf/app.ini

sed -i "s/DG_GIT_API_TOKEN/$DG_GIT_API_TOKEN/g" /data/gogs/custom/conf/app.ini