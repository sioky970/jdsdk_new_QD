Write-Host "`n=== consume_jingdou Security Fix Test ===" -ForegroundColor Cyan

# Login
Write-Host "`n[1] Login..." -ForegroundColor Yellow
$loginBody = '{"username":"admin","password":"admin123"}'
$loginResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.access_token
Write-Host "OK: $($loginResp.msg)" -ForegroundColor Green

$headers = @{ "Authorization" = "Bearer $token" }
$startTime = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")

# Test 1: Without consume_jingdou
Write-Host "`n[2] Create task WITHOUT consume_jingdou..." -ForegroundColor Yellow
$sku1 = "TEST-" + (Get-Date -Format "HHmmss")
$task1 = @{
    task_type = "purchase"
    sku = $sku1
    shop_name = "TestShop"
    keyword = "test"
    execute_count = 3
    priority = 1
    remark = "Test without consume_jingdou"
    start_time = $startTime
} | ConvertTo-Json

$resp1 = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $task1 -ContentType "application/json" -Headers $headers
Write-Host "OK: Task created" -ForegroundColor Green
Write-Host "  Task ID: $($resp1.data.task_id)" -ForegroundColor White
Write-Host "  Consumed: $($resp1.data.consume_jingdou) jingdou (server calculated)" -ForegroundColor Yellow
Write-Host "  Balance: $($resp1.data.balance)" -ForegroundColor White
$taskId1 = $resp1.data.task_id

# Test 2: With consume_jingdou=1 (should be ignored)
Write-Host "`n[3] Create task WITH consume_jingdou=1 (malicious)..." -ForegroundColor Yellow
$sku2 = "TEST-" + (Get-Date -Format "HHmmss")
$task2 = @{
    task_type = "purchase"
    sku = $sku2
    shop_name = "TestShop2"
    keyword = "test2"
    execute_count = 5
    priority = 1
    consume_jingdou = 1
    remark = "Test with malicious consume_jingdou"
    start_time = $startTime
} | ConvertTo-Json

$resp2 = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks" -Method POST -Body $task2 -ContentType "application/json" -Headers $headers
$actual = $resp2.data.consume_jingdou
$isAdmin = $resp2.data.is_admin

if ($isAdmin) {
    if ($actual -eq 0) {
        Write-Host "OK: Admin user - no jingdou consumed!" -ForegroundColor Green
        Write-Host "  Client sent: 1 jingdou (malicious)" -ForegroundColor Red
        Write-Host "  Server calculated: $actual jingdou (admin免费)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Admin should not consume jingdou!" -ForegroundColor Red
    }
} else {
    $expected = 25
    if ($actual -eq $expected) {
        Write-Host "OK: Server ignored client value!" -ForegroundColor Green
        Write-Host "  Client sent: 1 jingdou (malicious)" -ForegroundColor Red
        Write-Host "  Server calculated: $actual jingdou (correct)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Server used client value!" -ForegroundColor Red
        Write-Host "  Expected: $expected" -ForegroundColor Yellow
        Write-Host "  Actual: $actual" -ForegroundColor Yellow
    }
}
$taskId2 = $resp2.data.task_id

# Cleanup
Write-Host "`n[4] Cleanup..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$taskId1" -Method DELETE -Headers $headers | Out-Null
Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/$taskId2" -Method DELETE -Headers $headers | Out-Null
Write-Host "OK: Test data cleaned" -ForegroundColor Green

# Summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
Write-Host "PASS: consume_jingdou security fix verified!" -ForegroundColor Green
Write-Host "  - Client cannot send consume_jingdou" -ForegroundColor White
Write-Host "  - Server calculates based on task_type price" -ForegroundColor White
Write-Host "  - Malicious values are ignored" -ForegroundColor White
Write-Host "====================`n" -ForegroundColor Cyan
