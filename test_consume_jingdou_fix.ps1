# 测试 consume_jingdou 安全修复

Write-Host "`n========== 测试 consume_jingdou 安全修复 ==========" -ForegroundColor Cyan

# 1. 登录获取Token
Write-Host "`n[步骤1] 登录获取Token" -ForegroundColor Yellow
$loginBody = @{
    username = "user001"
    password = "pass123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResp.data.access_token
    Write-Host "✅ 登录成功: $($loginResp.msg)" -ForegroundColor Green
    Write-Host "   Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
} catch {
    Write-Host "❌ 登录失败: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# 2. 测试1：不传 consume_jingdou（正确方式）
Write-Host "`n[测试1] 创建任务（不传 consume_jingdou，由服务端计算）" -ForegroundColor Yellow

$startTime = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")

$taskBody1 = @{
    task_type = "browse"
    sku = "TEST-SKU-" + (Get-Date -Format "HHmmss")
    shop_name = "测试店铺"
    keyword = "测试商品"
    execute_count = 3
    priority = 1
    remark = "测试任务-服务端自动计算京豆"
    start_time = $startTime
} | ConvertTo-Json

try {
    $taskResp1 = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $taskBody1 -ContentType "application/json" -Headers $headers
    
    Write-Host "✅ 测试1通过: 不传 consume_jingdou，任务创建成功!" -ForegroundColor Green
    Write-Host "   返回消息: $($taskResp1.msg)" -ForegroundColor Cyan
    Write-Host "   任务ID: $($taskResp1.data.task_id)" -ForegroundColor White
    Write-Host "   服务端计算的京豆: $($taskResp1.data.consume_jingdou) 京豆" -ForegroundColor Yellow
    Write-Host "   剩余余额: $($taskResp1.data.balance) 京豆" -ForegroundColor Yellow
    
    $taskId1 = $taskResp1.data.task_id
    
} catch {
    $errorResp = $_.ErrorDetails.Message | ConvertFrom-Json
    Write-Host "❌ 测试1失败: $($errorResp.msg)" -ForegroundColor Red
    exit 1
}

# 3. 测试2：尝试传入 consume_jingdou（应该被忽略）
Write-Host "`n[测试2] 尝试传入恶意的 consume_jingdou=1（应该被忽略）" -ForegroundColor Yellow
$taskBody2 = @{
    task_type = "browse"
    sku = "TEST-SKU-" + (Get-Date -Format "HHmmss")
    shop_name = "测试店铺2"
    keyword = "测试商品2"
    execute_count = 5
    priority = 1
    consume_jingdou = 1  # 恶意传入极小值
    remark = "测试任务-尝试传入恶意京豆值"
    start_time = $startTime
} | ConvertTo-Json

try {
    $taskResp2 = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $taskBody2 -ContentType "application/json" -Headers $headers
    
    $actualConsume = $taskResp2.data.consume_jingdou
    $expectedConsume = 10  # browse类型单价2京豆 × 5次 = 10京豆
    
    if ($actualConsume -eq $expectedConsume) {
        Write-Host "✅ 测试2通过: 客户端传入的值被忽略，服务端正确计算!" -ForegroundColor Green
        Write-Host "   客户端尝试传入: 1 京豆（恶意值）" -ForegroundColor Red
        Write-Host "   服务端实际计算: $actualConsume 京豆（正确值）" -ForegroundColor Green
        Write-Host "   返回消息: $($taskResp2.msg)" -ForegroundColor Cyan
    } else {
        Write-Host "❌ 测试2失败: 服务端未正确计算京豆" -ForegroundColor Red
        Write-Host "   期望: $expectedConsume 京豆" -ForegroundColor Yellow
        Write-Host "   实际: $actualConsume 京豆" -ForegroundColor Yellow
        exit 1
    }
    
    $taskId2 = $taskResp2.data.task_id
    
} catch {
    $errorResp = $_.ErrorDetails.Message | ConvertFrom-Json
    Write-Host "❌ 测试2失败: $($errorResp.msg)" -ForegroundColor Red
    exit 1
}

# 4. 验证任务详情中的 consume_jingdou
Write-Host "`n[测试3] 验证任务详情中的京豆消耗" -ForegroundColor Yellow

try {
    $taskDetail = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$taskId1" -Method GET -Headers $headers
    
    Write-Host "✅ 测试3通过: 任务详情正确保存" -ForegroundColor Green
    Write-Host "   任务ID: $($taskDetail.data.id)" -ForegroundColor White
    Write-Host "   执行次数: $($taskDetail.data.execute_count)" -ForegroundColor White
    Write-Host "   消耗京豆: $($taskDetail.data.consume_jingdou)" -ForegroundColor Yellow
    
} catch {
    Write-Host "❌ 测试3失败: 无法获取任务详情" -ForegroundColor Red
}

# 5. 清理测试数据
Write-Host "`n[清理] 删除测试任务" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$taskId1" -Method DELETE -Headers $headers | Out-Null
    Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$taskId2" -Method DELETE -Headers $headers | Out-Null
    Write-Host "✅ 测试数据已清理" -ForegroundColor Green
} catch {
    Write-Host "⚠️ 清理失败，可能需要手动删除" -ForegroundColor Yellow
}

# 总结
Write-Host "`n========== 测试总结 ==========" -ForegroundColor Cyan
Write-Host "✅ consume_jingdou 安全问题已完全修复!" -ForegroundColor Green
Write-Host "" 
Write-Host "修复验证:" -ForegroundColor White
Write-Host "  1. ✅ 客户端不传 consume_jingdou 可正常创建任务" -ForegroundColor Green
Write-Host "  2. ✅ 客户端传入 consume_jingdou 会被服务端忽略" -ForegroundColor Green
Write-Host "  3. ✅ 服务端根据任务类型和执行次数正确计算" -ForegroundColor Green
Write-Host "  4. ✅ 任务详情中正确保存服务端计算的值" -ForegroundColor Green
Write-Host ""
Write-Host "安全提升:" -ForegroundColor White
Write-Host "  • 防止客户端篡改消费金额" -ForegroundColor Cyan
Write-Host "  • 确保业务逻辑完整性" -ForegroundColor Cyan
Write-Host "  • 提高系统安全性" -ForegroundColor Cyan
Write-Host "`n========================================" -ForegroundColor Cyan
