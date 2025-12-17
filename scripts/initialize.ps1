# ========================================
# Database Initialization Script
# Initialize base data for JD Task Platform
# ========================================

param(
    [switch]$Force
)

Write-Host '========================================' -ForegroundColor Cyan
Write-Host '  JD Task Platform - Database Init' -ForegroundColor Cyan
Write-Host '========================================' -ForegroundColor Cyan
Write-Host ''

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Check MySQL container status
Write-Host '[1/4] Checking MySQL container...' -ForegroundColor Yellow
$containerStatus = docker ps --filter 'name=jd-task-mysql' --format '{{.Status}}'
if (-not $containerStatus) {
    Write-Host 'X MySQL container not running' -ForegroundColor Red
    exit 1
}
Write-Host 'OK MySQL container is running' -ForegroundColor Green
Write-Host ''

# Initialize task types
Write-Host '[2/4] Initializing task types...' -ForegroundColor Yellow
$taskTypesCount = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e 'SELECT COUNT(*) FROM jd.task_types;' 2>$null

if ($taskTypesCount -eq 0 -or $Force) {
    if ($Force -and $taskTypesCount -gt 0) {
        Write-Host '  - Clearing existing task types...' -ForegroundColor Gray
        docker exec -i jd-task-mysql mysql -uroot -p123456 jd -e 'DELETE FROM task_types;' 2>$null
    }
    
    Write-Host '  - Inserting 6 preset task types...' -ForegroundColor Gray
    docker exec -i jd-task-mysql mysql -uroot -p123456 jd -e \"INSERT INTO task_types (type_code, type_name, jingdou_price, is_active, time_slot1_start, time_slot1_end, time_slot2_start, time_slot2_end, is_system_preset, created_at, updated_at) VALUES ('browse', '浏览任务', 2, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()), ('search_browse', '关键词搜索浏览任务', 3, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()), ('add_to_cart', '加入购物车任务', 5, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()), ('follow_shop', '关注店铺任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()), ('follow_product', '收藏商品任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()), ('purchase', '购买商品任务', 10, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW());\" 2>$null
    
    $newCount = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e 'SELECT COUNT(*) FROM jd.task_types;' 2>$null
    Write-Host \"OK Task types initialized (count: $newCount)\" -ForegroundColor Green
} else {
    Write-Host \"OK Task types exist (count: $taskTypesCount), skipped\" -ForegroundColor Green
}
Write-Host ''

# Initialize system settings
Write-Host '[3/4] Initializing system settings...' -ForegroundColor Yellow
docker cp \"$scriptDir\init_data.sql\" jd-task-mysql:/tmp/init_data.sql 2>$null
docker exec -it jd-task-mysql mysql -uroot -p123456 jd -e 'source /tmp/init_data.sql' 2>$null | Out-Null

$settingsCount = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e 'SELECT COUNT(*) FROM jd.settings;' 2>$null
Write-Host \"OK System settings initialized (count: $settingsCount)\" -ForegroundColor Green
Write-Host ''

# Display summary
Write-Host '[4/4] Initialization Summary' -ForegroundColor Yellow
Write-Host '----------------------------------------' -ForegroundColor Gray

$taskTypesSummary = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e 'SELECT type_name, jingdou_price FROM jd.task_types ORDER BY id;' 2>$null
Write-Host 'Task Types:' -ForegroundColor Cyan
$taskTypesSummary -split \"\
\" | ForEach-Object {
    if ($_.Trim()) {
        $parts = $_ -split \"\	\"
        Write-Host \"  - $($parts[0]): $($parts[1]) Jingdou\" -ForegroundColor White
    }
}
Write-Host ''

Write-Host 'System Settings:' -ForegroundColor Cyan
$settingsSummary = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e 'SELECT param_key, param_value FROM jd.settings ORDER BY id;' 2>$null
$settingsSummary -split \"\
\" | ForEach-Object {
    if ($_.Trim()) {
        $parts = $_ -split \"\	\"
        Write-Host \"  - $($parts[0]): $($parts[1])\" -ForegroundColor White
    }
}
Write-Host ''

$adminCount = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e \"SELECT COUNT(*) FROM jd.users WHERE role='admin';\" 2>$null
$userCount = docker exec -i jd-task-mysql mysql -uroot -p123456 -sN -e \"SELECT COUNT(*) FROM jd.users WHERE role='user';\" 2>$null
Write-Host 'User Statistics:' -ForegroundColor Cyan
Write-Host \"  - Admin: $adminCount\" -ForegroundColor White
Write-Host \"  - User: $userCount\" -ForegroundColor White
Write-Host ''

Write-Host '========================================' -ForegroundColor Green
Write-Host '  OK Database Initialization Complete!' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
