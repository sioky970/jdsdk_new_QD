$baseUrl = "http://localhost:5001/api"

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "测试设备请求任务API（新逻辑）" -ForegroundColor Cyan
Write-Host "============================================`n" -ForegroundColor Cyan

# 设备请求任务
$deviceData = @{
    device_id = "TEST_DEVICE_" + (Get-Date -Format "HHmmss")
    device_name = "测试设备"
    device_type = "ios"
    device_model = "iPhone 14"
    os_version = "iOS 17"
    app_version = "1.0.0"
} | ConvertTo-Json -Depth 10

$headers = @{
    "Content-Type" = "application/json"
    "X-Device-Key" = "KKNN778899"
}

Write-Host "第1次请求任务..." -ForegroundColor Yellow
try {
    $response1 = Invoke-RestMethod -Uri "$baseUrl/devices/request-task" -Method POST -Headers $headers -Body $deviceData
    Write-Host "响应:" -ForegroundColor Green
    $response1 | ConvertTo-Json -Depth 10
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

Write-Host "`n等待2秒..." -ForegroundColor Gray
Start-Sleep -Seconds 2

Write-Host "`n第2次请求任务（测试是否能继续下发running状态的任务）..." -ForegroundColor Yellow
# 使用不同的设备ID
$deviceData2 = @{
    device_id = "TEST_DEVICE_" + (Get-Date -Format "HHmmss")
    device_name = "测试设备2"
    device_type = "ios"
    device_model = "iPhone 15"
    os_version = "iOS 17"
    app_version = "1.0.0"
} | ConvertTo-Json -Depth 10

try {
    $response2 = Invoke-RestMethod -Uri "$baseUrl/devices/request-task" -Method POST -Headers $headers -Body $deviceData2
    Write-Host "响应:" -ForegroundColor Green
    $response2 | ConvertTo-Json -Depth 10
} catch {
    Write-Host "错误: $_" -ForegroundColor Red
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "测试完成" -ForegroundColor Cyan
Write-Host "============================================`n" -ForegroundColor Cyan
