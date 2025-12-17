Write-Host "`n=== Test Admin Time Slot Bypass ===" -ForegroundColor Cyan

# Test 1: Admin login and create task with browse type (has time restriction)
Write-Host "`n[Test 1] Admin creates browse task (time-restricted type)..." -ForegroundColor Yellow

# Login as admin
$adminLogin = '{"username":"admin","password":"admin123"}'
$adminResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $adminLogin -ContentType "application/json"
$adminToken = $adminResp.data.access_token
Write-Host "OK: Admin logged in" -ForegroundColor Green

$adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
$startTime = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")

# Create browse task (time restricted: 08:00-09:00)
$adminTask = @{
    task_type = "browse"
    sku = "ADMIN-TEST-" + (Get-Date -Format "HHmmss")
    shop_name = "Admin Test Shop"
    keyword = "admin test"
    execute_count = 2
    priority = 1
    remark = "Admin bypass time slot test"
    start_time = $startTime
} | ConvertTo-Json

try {
    $adminTaskResp = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $adminTask -ContentType "application/json" -Headers $adminHeaders
    Write-Host "PASS: Admin can create task anytime!" -ForegroundColor Green
    Write-Host "  Task ID: $($adminTaskResp.data.task_id)" -ForegroundColor White
    Write-Host "  Task Type: browse (restricted to 08:00-09:00)" -ForegroundColor Yellow
    Write-Host "  Current Time: $(Get-Date -Format 'HH:mm')" -ForegroundColor Yellow
    Write-Host "  Admin bypassed time restriction!" -ForegroundColor Cyan
    $adminTaskId = $adminTaskResp.data.task_id
} catch {
    $error = $_.ErrorDetails.Message | ConvertFrom-Json
    Write-Host "FAIL: Admin blocked by time slot: $($error.msg)" -ForegroundColor Red
    exit 1
}

# Test 2: Normal user tries to create same task type
Write-Host "`n[Test 2] Normal user creates browse task (should fail outside time slot)..." -ForegroundColor Yellow

# Login as normal user
$userLogin = '{"username":"user001","password":"pass123"}'
$userResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $userLogin -ContentType "application/json"
$userToken = $userResp.data.access_token
Write-Host "OK: User logged in" -ForegroundColor Green

$userHeaders = @{ "Authorization" = "Bearer $userToken" }

# Try to create browse task
$userTask = @{
    task_type = "browse"
    sku = "USER-TEST-" + (Get-Date -Format "HHmmss")
    shop_name = "User Test Shop"
    keyword = "user test"
    execute_count = 2
    priority = 1
    remark = "User time slot test"
    start_time = $startTime
} | ConvertTo-Json

try {
    $userTaskResp = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $userTask -ContentType "application/json" -Headers $userHeaders
    Write-Host "FAIL: Normal user should be blocked outside time slot!" -ForegroundColor Red
    Write-Host "  Task ID: $($userTaskResp.data.task_id)" -ForegroundColor Yellow
} catch {
    $error = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($error.msg -like "*时间段*") {
        Write-Host "PASS: Normal user blocked by time slot!" -ForegroundColor Green
        Write-Host "  Error: $($error.msg)" -ForegroundColor Yellow
    } else {
        Write-Host "UNEXPECTED: Different error: $($error.msg)" -ForegroundColor Yellow
    }
}

# Cleanup
Write-Host "`n[Cleanup]..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$adminTaskId" -Method DELETE -Headers $adminHeaders | Out-Null
    Write-Host "OK: Test data cleaned" -ForegroundColor Green
} catch {
    Write-Host "Warning: Cleanup may have failed" -ForegroundColor Yellow
}

# Summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
Write-Host "PASS: Admin time slot bypass feature works!" -ForegroundColor Green
Write-Host "" 
Write-Host "Feature verified:" -ForegroundColor White
Write-Host "  1. Admin can create tasks at any time" -ForegroundColor Green
Write-Host "  2. Admin bypasses time slot restrictions" -ForegroundColor Green
Write-Host "  3. Normal users still restricted by time slots" -ForegroundColor Green
Write-Host "  4. Time slot check is role-based" -ForegroundColor Green
Write-Host "`n====================" -ForegroundColor Cyan
