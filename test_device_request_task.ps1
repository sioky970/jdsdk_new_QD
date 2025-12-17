# 测试设备请求任务API
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  设备请求任务API测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:5001/api"
$deviceKey = "KKNN778899"
$deviceId = "TEST_DEVICE_" + (Get-Date -Format "HHmmss")

$deviceData = @{
    device_id = $deviceId
    device_name = "测试设备"
    device_type = "android"
    device_model = "Pixel 6"
    os_version = "Android 13"
    app_version = "1.0.0"
} | ConvertTo-Json

$headers = @{
    "Content-Type" = "application/json"
    "X-Device-Key" = $deviceKey
}

Write-Host "[1/4] 登录并创建测试任务..." -ForegroundColor Yellow
$loginResp = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -ContentType "application/json" -Body '{\"username\":\"admin\",\"password\":\"admin123\"}'
$token = $loginResp.data.access_token
Write-Host "登录成功" -ForegroundColor Green

$taskData = @{
    task_type = "browse"
    sku = "TEST_SKU_" + (Get-Date -Format "HHmmss")
    shop_name = "测试店铺"
    start_time = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    execute_count = 5
    priority = 1
    remark = "设备请求任务测试"
} | ConvertTo-Json

$taskResp = Invoke-RestMethod -Uri "$baseUrl/tasks" -Method POST -Headers @{Authorization="Bearer $token";ContentType="application/json"} -Body $taskData
$taskId = $taskResp.data.task_id
Write-Host "任务创建成功, ID: $taskId" -ForegroundColor Green

Write-Host ""
Write-Host "[2/4] 设备请求任务..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/devices/request-task" -Method POST -Headers $headers -Body $deviceData

Write-Host "请求成功！" -ForegroundColor Green
Write-Host ($response | ConvertTo-Json -Depth 3)

if ($response.data.has_task) {
    Write-Host ""
    Write-Host "成功获取任务！" -ForegroundColor Green
    Write-Host "  任务ID: $($response.data.task_id)" -ForegroundColor White
    Write-Host "  任务类型: $($response.data.task_type)" -ForegroundColor White
    Write-Host "  SKU: $($response.data.sku)" -ForegroundColor White
    
    $assignedTaskId = $response.data.task_id
    
    Write-Host ""
    Write-Host "[3/4] 提交任务反馈..." -ForegroundColor Yellow
    $feedbackData = @{
        device_id = $deviceId
        task_id = $assignedTaskId
        status = "success"
        message = "任务执行成功"
    } | ConvertTo-Json
    
    $feedbackResp = Invoke-RestMethod -Uri "$baseUrl/devices/task-feedback" -Method POST -Headers $headers -Body $feedbackData
    Write-Host "反馈提交成功: $($feedbackResp.msg)" -ForegroundColor Green
} else {
    Write-Host "暂无待执行任务: $($response.data.message)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[4/4] 验证任务状态..." -ForegroundColor Yellow
$taskDetailResp = Invoke-RestMethod -Uri "$baseUrl/tasks/$taskId" -Headers @{Authorization="Bearer $token"}
Write-Host "任务状态: $($taskDetailResp.data.status)" -ForegroundColor Green
Write-Host "已执行: $($taskDetailResp.data.executed_count)/$($taskDetailResp.data.execute_count)" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
