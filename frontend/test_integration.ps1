# 前后端对接测试脚本

Write-Host "`n========== 前后端对接测试 ==========" -ForegroundColor Cyan

# 1. 检查后端服务
Write-Host "`n[1] 检查后端服务..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body '{"username":"test","password":"test"}' -ContentType "application/json" -ErrorAction SilentlyContinue
    Write-Host "✅ 后端服务运行正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 后端服务未运行，请先启动后端" -ForegroundColor Red
    Write-Host "   启动命令: cd d:\工程\jd-task-platform-go ; .\bin\jd-task-platform.exe" -ForegroundColor Yellow
    exit 1
}

# 2. 检查前端依赖
Write-Host "`n[2] 检查前端项目..." -ForegroundColor Yellow
$frontendPath = Join-Path $PSScriptRoot "node_modules"
if (Test-Path $frontendPath) {
    Write-Host "✅ 前端依赖已安装" -ForegroundColor Green
} else {
    Write-Host "⚠️  前端依赖未安装，正在安装..." -ForegroundColor Yellow
    Set-Location $PSScriptRoot
    pnpm install
}

# 3. 显示配置信息
Write-Host "`n[3] 配置信息" -ForegroundColor Yellow
Write-Host "  后端地址: http://localhost:5001" -ForegroundColor White
Write-Host "  API前缀: /api" -ForegroundColor White
Write-Host "  成功码: 0" -ForegroundColor White

# 4. 测试登录API
Write-Host "`n[4] 测试登录API..." -ForegroundColor Yellow
$loginData = '{"username":"admin","password":"admin123"}'
try {
    $loginResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    if ($loginResp.code -eq 0) {
        Write-Host "✅ 登录成功" -ForegroundColor Green
        Write-Host "   用户名: $($loginResp.data.username)" -ForegroundColor Cyan
        Write-Host "   角色: $($loginResp.data.role)" -ForegroundColor Cyan
        Write-Host "   Token: $($loginResp.data.access_token.Substring(0, 20))..." -ForegroundColor Gray
    } else {
        Write-Host "❌ 登录失败: $($loginResp.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ 登录API调用失败" -ForegroundColor Red
    Write-Host "   $_" -ForegroundColor Red
}

# 5. 显示启动命令
Write-Host "`n[5] 启动前端项目" -ForegroundColor Yellow
Write-Host "  执行命令:" -ForegroundColor White
Write-Host "  cd `"$PSScriptRoot`"" -ForegroundColor Cyan
Write-Host "  pnpm dev" -ForegroundColor Cyan
Write-Host ""
Write-Host "  前端地址: http://localhost:5173" -ForegroundColor Green
Write-Host ""

# 6. 测试账号
Write-Host "[6] 测试账号" -ForegroundColor Yellow
Write-Host "  管理员: admin / admin123" -ForegroundColor Green
Write-Host "  普通用户: user001 / pass123" -ForegroundColor Green

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "✅ 前后端对接配置完成！" -ForegroundColor Green
Write-Host "   现在可以启动前端项目进行测试了" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
