# 测试任务过期退款功能
$baseUrl = "http://localhost:5001/api"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  测试任务过期退款功能" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 1. 先登录获取token
Write-Host "`n[1] 管理员登录..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json; charset=utf-8"
    $token = $loginResp.data.access_token
    Write-Host "    登录成功!" -ForegroundColor Green
} catch {
    Write-Host "    登录失败: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json; charset=utf-8"
}

# 2. 获取testuser的信息
Write-Host "`n[2] 获取testuser信息..." -ForegroundColor Yellow
try {
    $usersResp = Invoke-RestMethod -Uri "$baseUrl/users?page=1`&page_size=100" -Headers $headers
    $testuser = $usersResp.data.users | Where-Object { $_.username -eq "testuser" } | Select-Object -First 1
    
    if ($testuser) {
        Write-Host "    用户ID: $($testuser.id)" -ForegroundColor Green
        Write-Host "    用户名: $($testuser.username)" -ForegroundColor Green
        Write-Host "    当前余额: $($testuser.jingdou_balance) 京豆" -ForegroundColor Green
    } else {
        Write-Host "    testuser不存在" -ForegroundColor Yellow
    }
} catch {
    Write-Host "    获取用户失败: $_" -ForegroundColor Red
}

# 3. 检查当前过期任务状态
Write-Host "`n[3] 检查过期任务状态..." -ForegroundColor Yellow
try {
    $tasksResp = Invoke-RestMethod -Uri "$baseUrl/tasks?page=1`&page_size=10`&status=partial_completed" -Headers $headers
    Write-Host "    partial_completed 状态任务数: $($tasksResp.data.total)" -ForegroundColor Green
    
    if ($tasksResp.data.tasks -and $tasksResp.data.tasks.Count -gt 0) {
        $task = $tasksResp.data.tasks[0]
        Write-Host "    示例任务:" -ForegroundColor Cyan
        Write-Host "      - 任务ID: $($task.id)" -ForegroundColor White
        Write-Host "      - SKU: $($task.sku)" -ForegroundColor White
        Write-Host "      - 状态: $($task.status)" -ForegroundColor White
        Write-Host "      - 完成: $($task.executed_count)/$($task.execute_count)" -ForegroundColor White
        Write-Host "      - 备注: $($task.remark)" -ForegroundColor White
    }
} catch {
    Write-Host "    查询任务失败: $_" -ForegroundColor Red
}

# 4. 手动触发过期检查
Write-Host "`n[4] 手动触发过期检查..." -ForegroundColor Yellow
try {
    $checkResp = Invoke-RestMethod -Uri "$baseUrl/admin/dashboard/trigger-expire-check" -Method POST -Headers $headers
    Write-Host "    $($checkResp.data.message)" -ForegroundColor Green
    Write-Host "    处理任务数: $($checkResp.data.processed_count)" -ForegroundColor Cyan
    
    if ($checkResp.data.processed_tasks -and $checkResp.data.processed_tasks.Count -gt 0) {
        Write-Host "    处理详情:" -ForegroundColor Cyan
        foreach ($task in $checkResp.data.processed_tasks | Select-Object -First 5) {
            Write-Host "      - 任务$($task.task_id): $($task.sku) [$($task.old_status) -> $($task.new_status)] 完成$($task.executed)/$($task.total) 退款$($task.refund_jingdou)京豆" -ForegroundColor White
        }
    }
} catch {
    Write-Host "    触发过期检查失败: $_" -ForegroundColor Red
}

# 5. 查看任务列表中的 partial_completed 状态
Write-Host "`n[5] 查看已过期任务列表..." -ForegroundColor Yellow
try {
    $tasksResp = Invoke-RestMethod -Uri "$baseUrl/tasks?page=1`&page_size=5`&status=partial_completed" -Headers $headers
    Write-Host "    共 $($tasksResp.data.total) 个过期任务" -ForegroundColor Green
    
    if ($tasksResp.data.tasks -and $tasksResp.data.tasks.Count -gt 0) {
        Write-Host "    前5个任务:" -ForegroundColor Cyan
        foreach ($task in $tasksResp.data.tasks | Select-Object -First 5) {
            Write-Host "      - 任务$($task.id): $($task.sku) 完成$($task.executed_count)/$($task.execute_count)" -ForegroundColor White
        }
    }
} catch {
    Write-Host "    查询失败: $_" -ForegroundColor Red
}

# 6. 测试设备请求任务（应该不会获取到过期任务）
Write-Host "`n[6] 测试设备请求任务（验证过期任务不会被下发）..." -ForegroundColor Yellow
try {
    $deviceBody = @{ device_id = "test_expire_device" } | ConvertTo-Json
    $deviceResp = Invoke-RestMethod -Uri "$baseUrl/devices/request-task" -Method POST -Headers $headers -Body $deviceBody
    
    if ($deviceResp.data.has_task) {
        Write-Host "    获取到任务ID: $($deviceResp.data.task_id)" -ForegroundColor Green
        Write-Host "    任务类型: $($deviceResp.data.task_type)" -ForegroundColor Cyan
        Write-Host "    SKU: $($deviceResp.data.sku)" -ForegroundColor Cyan
    } else {
        Write-Host "    $($deviceResp.data.message)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "    请求任务失败: $_" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  测试完成!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
