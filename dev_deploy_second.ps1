# Second-level: check every 30 seconds, time window last 60 seconds
$APP_NAME = "async_account_check-second"
$IMAGE_NAME = "async_account_check"
$TAG = "second"
$CONFIG_FILE = "etc/config.yaml"

docker rm -f $APP_NAME
docker rmi -f "${IMAGE_NAME}:${TAG}"

Write-Host "Building Docker image for second-level..."
docker build -t "${IMAGE_NAME}:${TAG}" --build-arg CONFIG_FILE=$CONFIG_FILE .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful"
} else {
    Write-Host "Build failed"
    exit 1
}

Write-Host "Starting container..."
docker run -d --name $APP_NAME `
    --network dev_pay_net `
    -e TZ=Asia/Shanghai `
    "${IMAGE_NAME}:${TAG}"

Write-Host "Container started"
docker ps | Select-String $APP_NAME
