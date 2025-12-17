#!/usr/bin/env pwsh
# JD任务平台后端启动脚本
# 自动检测并结束占用端口5001的进程，然后启动Go后端服务

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  JD任务平台后端启动脚本" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$BackendDir = "d:\工程\jd-task-platform-go"
$Port = 5001

Write-Host "[1/3] 切换到后端目录..." -ForegroundColor Yellow
if (-not (Test-Path $BackendDir)) {
    Write-Host "错误：后端目录不存在: $BackendDir" -ForegroundColor Red
    exit 1
}
Set-Location $BackendDir
Write-Host "当前目录: $BackendDir" -ForegroundColor Green

Write-Host ""
Write-Host "[2/3] 检查端口 $Port 占用情况..." -ForegroundColor Yellow
try {
    $connection = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
    if ($connection) {
        $processId = $connection.OwningProcess
        $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
        if ($process) {
            Write-Host "发现占用端口 $Port 的进程: PID=$processId Name=$($process.ProcessName)" -ForegroundColor Yellow
            Write-Host "正在结束进程..." -ForegroundColor Yellow
            Stop-Process -Id $processId -Force
            Start-Sleep -Seconds 1
            Write-Host "进程已结束" -ForegroundColor Green
        }
    } else {
        Write-Host "端口 $Port 当前未被占用" -ForegroundColor Green
    }
} catch {
    Write-Host "端口 $Port 当前未被占用" -ForegroundColor Green
}

Write-Host ""
Write-Host "[3/3] 启动Go后端服务..." -ForegroundColor Yellow
Write-Host "执行命令: go run main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  后端服务日志输出" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

go run main.go
