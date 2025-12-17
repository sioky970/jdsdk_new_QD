# API 返回消息审计脚本

$handlerFiles = Get-ChildItem -Path "d:\工程\jd-task-platform-go\internal\handlers" -Filter "*.go"

Write-Host "========== API 返回消息审计 ==========" -ForegroundColor Cyan
Write-Host ""

$allMessages = @()

foreach ($file in $handlerFiles) {
    Write-Host "检查文件: $($file.Name)" -ForegroundColor Yellow
    
    $content = Get-Content $file.FullName -Raw
    
    # 查找所有 response.Error 调用
    $errorMatches = [regex]::Matches($content, 'response\.Error\([^,]+,\s*[^,]+,\s*"([^"]+)"')
    foreach ($match in $errorMatches) {
        $msg = $match.Groups[1].Value
        $allMessages += [PSCustomObject]@{
            File = $file.Name
            Type = "Error"
            Message = $msg
        }
    }
    
    # 查找所有 response.SuccessWithMsg 调用
    $successMatches = [regex]::Matches($content, 'response\.SuccessWithMsg\([^,]+,\s*"([^"]+)"')
    foreach ($match in $successMatches) {
        $msg = $match.Groups[1].Value
        $allMessages += [PSCustomObject]@{
            File = $file.Name
            Type = "Success"
            Message = $msg
        }
    }
}

Write-Host "`n========== 汇总统计 ==========" -ForegroundColor Cyan
Write-Host "总消息数: $($allMessages.Count)" -ForegroundColor White
Write-Host "错误消息: $(($allMessages | Where-Object {$_.Type -eq 'Error'}).Count)" -ForegroundColor Red
Write-Host "成功消息: $(($allMessages | Where-Object {$_.Type -eq 'Success'}).Count)" -ForegroundColor Green

Write-Host "`n========== 所有消息 ==========" -ForegroundColor Cyan
$allMessages | Format-Table -AutoSize

# 导出到CSV
$allMessages | Export-Csv -Path "d:\工程\jd-task-platform-go\api_messages_audit.csv" -NoTypeInformation -Encoding UTF8
Write-Host "`n已导出到: api_messages_audit.csv" -ForegroundColor Green
