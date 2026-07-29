$APP_NAME = "async_account_check"
$TAG = "latest"

Write-Host "Building Docker image..."
docker build -t "${APP_NAME}:${TAG}" .

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
    "${APP_NAME}:${TAG}"

Write-Host "Container started"
docker ps | Select-String $APP_NAME
