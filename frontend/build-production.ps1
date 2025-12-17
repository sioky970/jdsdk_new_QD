# JD任务平台 - 生产环境构建脚本

Write-Host "`n========== JD任务平台生产环境构建 ==========" -ForegroundColor Cyan

# 1. 检查环境配置
Write-Host "`n[1] 检查环境配置..." -ForegroundColor Yellow

if (Test-Path ".env.production") {
    $envContent = Get-Content ".env.production" -Raw
    if ($envContent -match "VITE_SERVICE_BASE_URL=https://your-production-domain.com") {
        Write-Host "⚠️  警告: 检测到默认后端地址，请先修改 .env.production" -ForegroundColor Red
        Write-Host "   当前: https://your-production-domain.com" -ForegroundColor Yellow
        Write-Host "   请修改为实际的生产环境后端地址" -ForegroundColor Yellow
        $continue = Read-Host "`n是否继续构建? (y/N)"
        if ($continue -ne "y" -and $continue -ne "Y") {
            Write-Host "构建已取消" -ForegroundColor Yellow
            exit 0
        }
    }
    
    # 显示当前配置
    if ($envContent -match "VITE_SERVICE_BASE_URL=(.+)") {
        Write-Host "✅ 后端地址: $($Matches[1])" -ForegroundColor Green
    }
} else {
    Write-Host "❌ 错误: 未找到 .env.production 文件" -ForegroundColor Red
    Write-Host "   请先复制 .env.production.example 为 .env.production 并修改配置" -ForegroundColor Yellow
    exit 1
}

# 2. 检查依赖
Write-Host "`n[2] 检查依赖..." -ForegroundColor Yellow
if (!(Test-Path "node_modules")) {
    Write-Host "⚠️  未找到 node_modules，正在安装依赖..." -ForegroundColor Yellow
    pnpm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 依赖安装失败" -ForegroundColor Red
        exit 1
    }
}
Write-Host "✅ 依赖检查完成" -ForegroundColor Green

# 3. 清理旧构建
Write-Host "`n[3] 清理旧构建..." -ForegroundColor Yellow
if (Test-Path "dist") {
    Remove-Item -Path "dist" -Recurse -Force
    Write-Host "✅ 已清理 dist 目录" -ForegroundColor Green
}

# 4. 执行构建
Write-Host "`n[4] 开始构建..." -ForegroundColor Yellow
Write-Host "⏳ 构建中，请稍候..." -ForegroundColor Cyan

$startTime = Get-Date
pnpm build

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ 构建失败，请检查错误信息" -ForegroundColor Red
    exit 1
}

$endTime = Get-Date
$duration = ($endTime - $startTime).TotalSeconds

Write-Host "`n✅ 构建成功！耗时: $([math]::Round($duration, 2)) 秒" -ForegroundColor Green

# 5. 显示构建结果
Write-Host "`n[5] 构建结果:" -ForegroundColor Yellow
if (Test-Path "dist") {
    $distSize = (Get-ChildItem -Path "dist" -Recurse | Measure-Object -Property Length -Sum).Sum / 1MB
    Write-Host "📦 构建目录: dist/" -ForegroundColor Cyan
    Write-Host "📊 总大小: $([math]::Round($distSize, 2)) MB" -ForegroundColor Cyan
    
    # 列出主要文件
    Write-Host "`n主要文件:" -ForegroundColor Cyan
    Get-ChildItem -Path "dist" -File | Select-Object Name, @{Name="Size(KB)";Expression={[math]::Round($_.Length/1KB, 2)}} | Format-Table -AutoSize
}

# 6. 部署建议
Write-Host "`n========== 部署建议 ==========" -ForegroundColor Cyan
Write-Host "📝 接下来的步骤:" -ForegroundColor Yellow
Write-Host "  1. 将 dist 目录上传到服务器" -ForegroundColor White
Write-Host "  2. 配置 Nginx/Apache (参考 DEPLOYMENT.md)" -ForegroundColor White
Write-Host "  3. 确保后端服务已启动" -ForegroundColor White
Write-Host "  4. 配置 SSL 证书（推荐）" -ForegroundColor White
Write-Host "  5. 测试访问前端页面" -ForegroundColor White

Write-Host "`n📖 详细部署说明请查看: DEPLOYMENT.md" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan
