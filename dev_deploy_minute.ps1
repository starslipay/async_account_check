# Minute-level: check every 60 seconds, time window last 120 seconds
$APP_NAME = "async_account_check-minute"
$IMAGE_NAME = "async_account_check"
$TAG = "minute"
$CONFIG_FILE = "etc/config_minute.yaml"

docker rm -f $APP_NAME
docker rmi -f "${IMAGE_NAME}:${TAG}"

Write-Host "Building Docker image for minute-level..."
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
