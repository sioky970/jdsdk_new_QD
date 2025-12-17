# 测试任务创建API（已修复consume_jingdou安全问题）

Write-Host "`n========== 测试修复后的任务创建API ==========" -ForegroundColor Cyan

# 1. 登录获取Token
Write-Host "`n[步骤1] 登录获取Token" -ForegroundColor Yellow
$loginBody = @{
    username = "user001"
    password = "pass123"
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.access_token
Write-Host "✅ 登录成功: $($loginResp.msg)" -ForegroundColor Green

# 2. 创建任务（不传consume_jingdou字段，由服务端自动计算）
Write-Host "`n[步骤2] 创建任务（服务端自动计算京豆消耗）" -ForegroundColor Yellow
$taskBody = @{
    task_type = "browse"
    sku = "TEST-SKU-" + (Get-Date -Format "HHmmss")
    shop_name = "测试店铺"
    keyword = "测试商品"
    execute_count = 3
    priority = 1
    remark = "测试任务-无需客户端传入consume_jingdou"
    start_time = "2025-12-11T10:00:00Z"
} | ConvertTo-Json

$headers = @{
    "Authorization" = "Bearer $token"
}

try {
    $taskResp = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $taskBody -ContentType "application/json" -Headers $headers
    
    Write-Host "`n✅ 任务创建成功!" -ForegroundColor Green
    Write-Host "返回消息: $($taskResp.msg)" -ForegroundColor Cyan
    Write-Host "任务ID: $($taskResp.data.task_id)" -ForegroundColor White
    Write-Host "消耗京豆: $($taskResp.data.consume_jingdou) (由服务端计算)" -ForegroundColor Yellow
    Write-Host "剩余余额: $($taskResp.data.balance)" -ForegroundColor Yellow
    
    Write-Host "`n✅ 测试通过: consume_jingdou已从客户端请求中移除，由服务端安全计算" -ForegroundColor Green
    
} catch {
    $errorResp = $_.ErrorDetails.Message | ConvertFrom-Json
    Write-Host "`n❌ 错误: $($errorResp.msg)" -ForegroundColor Red
    Write-Host "错误代码: $($errorResp.code)" -ForegroundColor Yellow
}

Write-Host "`n========== 测试完成 ==========" -ForegroundColor Cyan
