#!/bin/bash

APP_NAME="async_account_check"
TAG="latest"

echo "Building Docker image..."
docker build -t ${APP_NAME}:${TAG} .

if [ $? -eq 0 ]; then
    echo "Build successful"
else
    echo "Build failed"
    exit 1
fi

echo "Starting container..."
docker run -d --name ${APP_NAME} \
    --network dev_pay_net \
    -e TZ=Asia/Shanghai \
    ${APP_NAME}:${TAG}

echo "Container started"
docker ps | grep ${APP_NAME}
