# JD任务平台 Go版本 - 一键启动脚本

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  JD任务平台 Go版本 启动脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 切换到项目目录
$projectDir = "d:\工程\jd-task-platform-go"
Set-Location $projectDir

# 1. 检查 Go 环境
Write-Host "[1/5] 检查 Go 环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Go 已安装: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go 未安装，请先安装 Go" -ForegroundColor Red
    exit 1
}

# 2. 安装依赖
Write-Host ""
Write-Host "[2/5] 安装项目依赖..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 依赖安装成功" -ForegroundColor Green
} else {
    Write-Host "✗ 依赖安装失败" -ForegroundColor Red
    exit 1
}

# 3. 安装 Swag 工具
Write-Host ""
Write-Host "[3/5] 检查 Swag 工具..." -ForegroundColor Yellow
$swagPath = Join-Path $env:GOPATH "bin\swag.exe"
if (-not (Test-Path $swagPath)) {
    Write-Host "  正在安装 Swag..." -ForegroundColor Cyan
    go install github.com/swaggo/swag/cmd/swag@latest
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Swag 安装成功" -ForegroundColor Green
    } else {
        Write-Host "✗ Swag 安装失败" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✓ Swag 已安装" -ForegroundColor Green
}

# 4. 生成 Swagger 文档
Write-Host ""
Write-Host "[4/5] 生成 Swagger API 文档..." -ForegroundColor Yellow
$env:PATH += ";$env:GOPATH\bin"
& swag init
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Swagger 文档生成成功" -ForegroundColor Green
} else {
    Write-Host "✗ Swagger 文档生成失败" -ForegroundColor Red
    Write-Host "  请手动运行: swag init" -ForegroundColor Yellow
}

# 5. 检查 MySQL 容器
Write-Host ""
Write-Host "[5/5] 检查 MySQL 数据库..." -ForegroundColor Yellow
$mysqlContainer = docker ps --filter "name=jd-task-mysql" --format "{{.Status}}" 2>$null
if ($mysqlContainer -match "Up") {
    Write-Host "✓ MySQL 容器运行中" -ForegroundColor Green
} else {
    Write-Host "⚠ MySQL 容器未运行，正在启动..." -ForegroundColor Yellow
    cd "$projectDir\deploy\database"
    docker-compose up -d
    cd $projectDir
    Start-Sleep -Seconds 3
    Write-Host "✓ MySQL 容器已启动" -ForegroundColor Green
}

# 显示配置信息
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  配置信息" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  项目目录: $projectDir" -ForegroundColor White
Write-Host "  数据库: MySQL (Docker)" -ForegroundColor White
Write-Host "  数据库地址: localhost:3306" -ForegroundColor White
Write-Host "  数据库名: jd" -ForegroundColor White
Write-Host "  服务端口: 5001" -ForegroundColor White
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 启动选择
Write-Host "选择启动方式:" -ForegroundColor Yellow
Write-Host "1. 前台运行 (开发调试，查看实时日志)"
Write-Host "2. 编译并运行"
Write-Host "0. 取消"
Write-Host ""
$choice = Read-Host "请选择 (0-2)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "========================================" -ForegroundColor Green
        Write-Host "  正在启动服务..." -ForegroundColor Green
        Write-Host "========================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "  服务地址: http://localhost:5001" -ForegroundColor Cyan
        Write-Host "  API文档: http://localhost:5001/swagger/index.html" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "提示: 按 Ctrl+C 停止服务" -ForegroundColor Yellow
        Write-Host ""
        go run main.go
    }
    "2" {
        Write-Host ""
        Write-Host "正在编译..." -ForegroundColor Yellow
        go build -o jd-task-platform.exe main.go
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 编译成功" -ForegroundColor Green
            Write-Host ""
            Write-Host "========================================" -ForegroundColor Green
            Write-Host "  正在启动服务..." -ForegroundColor Green
            Write-Host "========================================" -ForegroundColor Green
            Write-Host ""
            Write-Host "  服务地址: http://localhost:5001" -ForegroundColor Cyan
            Write-Host "  API文档: http://localhost:5001/swagger/index.html" -ForegroundColor Cyan
            Write-Host ""
            .\jd-task-platform.exe
        } else {
            Write-Host "✗ 编译失败" -ForegroundColor Red
        }
    }
    "0" {
        Write-Host "已取消" -ForegroundColor Yellow
        exit 0
    }
    default {
        Write-Host "无效选项" -ForegroundColor Red
        exit 1
    }
}
