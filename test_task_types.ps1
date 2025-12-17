# 测试任务类型预设系统

$baseUrl = "http://localhost:5001/api/v1"

Write-Host "========== 任务类型预设系统测试 ==========" -ForegroundColor Cyan

# 登录管理员
Write-Host "`n[1/6] 管理员登录..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body (@{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json) -ContentType "application/json"

$token = $loginResponse.data.token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "✓ 登录成功，Token: $($token.Substring(0, 20))..." -ForegroundColor Green

# 测试1：获取任务类型列表（查看预设类型和时间段）
Write-Host "`n[2/6] 获取任务类型列表..." -ForegroundColor Yellow
$typesResponse = Invoke-RestMethod -Uri "$baseUrl/tasks/types" -Method GET -Headers $headers

Write-Host "✓ 任务类型列表（共 $($typesResponse.data.task_types.Count) 个）:" -ForegroundColor Green
foreach ($type in $typesResponse.data.task_types) {
    $timeSlot = "无时间限制"
    if ($type.time_slot1_start -and $type.time_slot1_end) {
        $timeSlot = "{0}-{1}" -f $type.time_slot1_start, $type.time_slot1_end
        if ($type.time_slot2_start -and $type.time_slot2_end) {
            $timeSlot += ", {0}-{1}" -f $type.time_slot2_start, $type.time_slot2_end
        }
    }
    $preset = ""
    if ($type.is_system_preset) { 
        $preset = " [系统预设]" 
    }
    Write-Host "  [$($type.id)] $($type.type_name) ($($type.type_code))" -ForegroundColor Cyan
    Write-Host "      价格: $($type.jingdou_price)京豆 | 启用: $($type.is_active) | 时间段: $timeSlot$preset"
}

# 测试2：尝试创建新任务类型（应该失败）
Write-Host "`n[3/6] 测试创建新任务类型（应被禁止）..." -ForegroundColor Yellow
try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/tasks/types" -Method POST -Headers $headers -Body (@{
        type_code = "custom_task"
        type_name = "自定义任务"
        jingdou_price = 5
    } | ConvertTo-Json)
    Write-Host "✗ 创建成功（不应该允许）" -ForegroundColor Red
} catch {
    $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($errorResponse.message -like "*不允许创建*") {
        Write-Host "✓ 正确拦截：$($errorResponse.message)" -ForegroundColor Green
    } else {
        Write-Host "⚠ 错误信息: $($errorResponse.message)" -ForegroundColor Yellow
    }
}

# 测试3：更新任务类型（修改价格和时间段）
Write-Host "`n[4/6] 更新任务类型配置..." -ForegroundColor Yellow
$browseType = $typesResponse.data.task_types | Where-Object { $_.type_code -eq "browse" }
$updateResponse = Invoke-RestMethod -Uri "$baseUrl/tasks/types/$($browseType.id)" -Method PUT -Headers $headers -Body (@{
    jingdou_price = 3
    is_active = $true
    time_slot1_start = "09:00"
    time_slot1_end = "11:00"
    time_slot2_start = "15:00"
    time_slot2_end = "17:00"
} | ConvertTo-Json)

Write-Host "✓ 更新成功：$($updateResponse.message)" -ForegroundColor Green

# 测试4：验证更新
Write-Host "`n[5/6] 验证更新结果..." -ForegroundColor Yellow
$typesResponse2 = Invoke-RestMethod -Uri "$baseUrl/tasks/types" -Method GET -Headers $headers
$browseTypeUpdated = $typesResponse2.data.task_types | Where-Object { $_.type_code -eq "browse" }

Write-Host "✓ 浏览任务配置：" -ForegroundColor Green
Write-Host "  价格: $($browseTypeUpdated.jingdou_price)京豆 (应该是3)"
$ts1 = "{0}-{1}" -f $browseTypeUpdated.time_slot1_start, $browseTypeUpdated.time_slot1_end
$ts2 = "{0}-{1}" -f $browseTypeUpdated.time_slot2_start, $browseTypeUpdated.time_slot2_end
Write-Host "  时间段: $ts1, $ts2"

# 测试5：测试时间段限制（尝试在非允许时间段创建任务）
Write-Host "`n[6/6] 测试时间段限制（创建任务）..." -ForegroundColor Yellow

# 获取当前时间
$currentTime = Get-Date -Format "HH:mm"
Write-Host "  当前时间: $currentTime"
Write-Host "  允许时间段: 09:00-11:00, 15:00-17:00"

# 登录普通用户
$userLoginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body (@{
    username = "testuser"
    password = "test123"
} | ConvertTo-Json) -ContentType "application/json"

$userToken = $userLoginResponse.data.token
$userHeaders = @{
    "Authorization" = "Bearer $userToken"
    "Content-Type" = "application/json"
}

try {
    $taskResponse = Invoke-RestMethod -Uri "$baseUrl/tasks" -Method POST -Headers $userHeaders -Body (@{
        task_type = "browse"
        sku = "TEST123456"
        shop_name = "测试店铺"
        start_time = (Get-Date).AddMinutes(10).ToString("yyyy-MM-ddTHH:mm:ssZ")
        execute_count = 1
        priority = 1
        remark = "测试时间段限制"
    } | ConvertTo-Json)
    
    Write-Host "✓ 任务创建成功（当前在允许时间段内）" -ForegroundColor Green
    Write-Host "  任务ID: $($taskResponse.data.task_id)"
    Write-Host "  消耗京豆: $($taskResponse.data.consume_jingdou)京豆"
} catch {
    $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($errorResponse.message -like "*时间段*") {
        Write-Host "✓ 正确拦截：$($errorResponse.message)" -ForegroundColor Green
    } else {
        Write-Host "⚠ 错误: $($errorResponse.message)" -ForegroundColor Yellow
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
