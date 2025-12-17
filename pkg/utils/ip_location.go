package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IPLocationInfo IP地理位置信息
type IPLocationInfo struct {
	Province string `json:"province"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
}

// IPQueryResponse IP查询API响应 (使用ip.taobao.com API)
type IPQueryResponse struct {
	Code int `json:"code"`
	Data struct {
		Country  string `json:"country"`
		Province string `json:"region"`
		City     string `json:"city"`
		ISP      string `json:"isp"`
	} `json:"data"`
}

// QueryIPLocation 查询IP地理位置
// 使用淘宝IP库API (免费)
func QueryIPLocation(ip string) (*IPLocationInfo, error) {
	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 构建请求URL
	url := fmt.Sprintf("https://ip.taobao.com/outGetIpInfo?ip=%s&accessKey=alibaba-inc", ip)

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("IP查询请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var apiResp IPQueryResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("IP查询失败，错误码: %d", apiResp.Code)
	}

	// 构建返回结果
	location := &IPLocationInfo{
		Province: strings.TrimSpace(apiResp.Data.Province),
		City:     strings.TrimSpace(apiResp.Data.City),
		ISP:      strings.TrimSpace(apiResp.Data.ISP),
	}

	return location, nil
}

// QueryIPLocationWithFallback 带降级的IP查询（如果淘宝API失败，使用备用方案）
func QueryIPLocationWithFallback(ip string) (*IPLocationInfo, error) {
	// 尝试淘宝API
	location, err := QueryIPLocation(ip)
	if err == nil {
		return location, nil
	}

	// 如果失败，返回默认值而不是错误
	return &IPLocationInfo{
		Province: "未知",
		City:     "未知",
		ISP:      "未知",
	}, nil
}

// ParseProxyLine 解析代理行格式：IP|Port|Username|Password
func ParseProxyLine(line string) (ip string, port int, username string, password string, err error) {
	parts := strings.Split(strings.TrimSpace(line), "|")
	if len(parts) != 4 {
		return "", 0, "", "", fmt.Errorf("代理格式错误，应为：IP|Port|Username|Password")
	}

	ip = strings.TrimSpace(parts[0])
	username = strings.TrimSpace(parts[2])
	password = strings.TrimSpace(parts[3])

	// 解析端口
	var portNum int
	if _, err := fmt.Sscanf(parts[1], "%d", &portNum); err != nil {
		return "", 0, "", "", fmt.Errorf("端口格式错误: %v", err)
	}

	if portNum < 1 || portNum > 65535 {
		return "", 0, "", "", fmt.Errorf("端口范围错误，应在1-65535之间")
	}

	return ip, portNum, username, password, nil
}

// GetLocationByIP 获取IP地理位置(兼容旧代码)
func GetLocationByIP(ip string) string {
	location, err := QueryIPLocationWithFallback(ip)
	if err != nil {
		return "未知位置"
	}

	// 构建位置字符串
	if location.Province != "未知" {
		if location.City != "" && location.City != location.Province {
			return location.Province + " " + location.City
		}
		return location.Province
	}

	return "未知位置"
}
