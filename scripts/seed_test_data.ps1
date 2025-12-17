# 创建测试数据脚本

$baseUrl = "http://localhost:5001"

# 登录获取token
$login = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method POST -Body '{"username":"admin","password":"admin123"}' -ContentType "application/json"
$token = $login.data.access_token
$headers = @{ "Authorization" = "Bearer $token" }

Write-Host "=== 创建过去3天的已完成任务 ===" -ForegroundColor Cyan

for ($day = 1; $day -le 3; $day++) {
    $pastDate = (Get-Date).AddDays(-$day).ToString("yyyy-MM-dd") + "T10:00:00Z"
    $createdCount = 0
    
    for ($i = 1; $i -le 10; $i++) {
        $sku = "HIST" + $day + "_" + $i + "_" + (Get-Random -Maximum 9999)
        $taskJson = @{
            task_type = "browse"
            sku = $sku
            shop_name = "HistoryShop Day$day #$i"
            keyword = "history"
            start_time = $pastDate
            execute_count = 5
            priority = 1
        } | ConvertTo-Json
        
        try {
            $resp = Invoke-RestMethod -Uri "$baseUrl/api/tasks" -Method POST -Body $taskJson -ContentType "application/json" -Headers $headers
            
            if ($resp.code -eq 0 -and $resp.data.task_id) {
                $taskId = $resp.data.task_id
                # 设置为已完成
                $updateBody = '{"status":"completed"}'
                Invoke-RestMethod -Uri "$baseUrl/api/tasks/$taskId" -Method PUT -Body $updateBody -ContentType "application/json" -Headers $headers | Out-Null
                $createdCount++
            }
        } catch {
            Write-Host "Error creating task: $_" -ForegroundColor Red
        }
    }
    
    Write-Host "Day -$day ($pastDate): Created $createdCount completed tasks" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== 验证压力统计 ===" -ForegroundColor Cyan
$pressure = Invoke-RestMethod -Uri "$baseUrl/api/admin/dashboard/task-pressure" -Method GET -Headers $headers
Write-Host "昨日完成: $($pressure.data.yesterday_completed)"
Write-Host "3日平均: $($pressure.data.avg_3days_completed)"
Write-Host "压力等级: $($pressure.data.pressure_level)"
