package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
	"jd-task-platform-go/pkg/utils"
)

type ProxyHandler struct {
	db *gorm.DB
}

func NewProxyHandler(db *gorm.DB) *ProxyHandler {
	return &ProxyHandler{db: db}
}

// generateClashQRCodeURL 生成 SOCKS5 分享链接
// 返回格式：socks://base64(用户名:密码@IP:端口)#节点名称
func generateClashQRCodeURL(proxyID uint, backendHost string) string {
	// 这个函数现在不需要使用 proxyID 和 backendHost
	// 因为在调用时会传入完整的代理信息
	// 这里保持函数签名不变，实际实现在调用处
	return ""
}

// generateSocksShareURL 生成 SOCKS5 分享链接
// 格式：socks://base64(用户名:密码@IP:端口)#节点名称
func generateSocksShareURL(ip string, port int, username, password string) string {
	// 构造认证信息：用户名:密码@IP:端口
	authInfo := fmt.Sprintf("%s:%s@%s:%d", username, password, ip, port)

	// Base64 编码
	base64Auth := base64.StdEncoding.EncodeToString([]byte(authInfo))

	// 构造节点名称
	nodeName := fmt.Sprintf("SOCKS5-%s", ip)

	// 构造最终的 socks:// URL
	return fmt.Sprintf("socks://%s#%s", base64Auth, url.QueryEscape(nodeName))
}

// getBackendHost 获取后端服务器地址
// 从环境变量或请求头中获取
func getBackendHost(c *gin.Context) string {
	// 优先从环境变量读取
	// host := os.Getenv("BACKEND_HOST")
	// if host != "" {
	//     return host
	// }

	// 从请求构造（支持内网和外网）
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

// GetProxies 获取代理列表
// @Summary 获取代理列表
// @Description 获取SK5代理池列表（分页）
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param is_active query bool false "是否激活"
// @Param keyword query string false "关键词搜索(IP/备注)"
// @Success 200 {object} response.Response{data=object}
// @Router /proxies [get]
func (h *ProxyHandler) GetProxies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	isActiveStr := c.Query("is_active")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := h.db.Model(&models.Proxy{})

	// 激活状态过滤
	if isActiveStr != "" {
		isActive, _ := strconv.ParseBool(isActiveStr)
		query = query.Where("is_active = ?", isActive)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("ip LIKE ? OR remark LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取列表
	var proxies []models.Proxy
	query.Order("usage_count ASC, id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&proxies)

	response.Success(c, gin.H{
		"proxies":   proxies,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetProxyByID 获取代理详情
// @Summary 获取代理详情
// @Description 根据ID获取代理详细信息
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Success 200 {object} response.Response{data=models.Proxy}
// @Router /proxies/{id} [get]
func (h *ProxyHandler) GetProxyByID(c *gin.Context) {
	id := c.Param("id")

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "代理不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 查询使用记录
	var usageLogs []models.ProxyUsageLog
	h.db.Where("proxy_id = ?", proxy.ID).
		Order("assigned_at DESC").
		Limit(10).
		Find(&usageLogs)

	response.Success(c, gin.H{
		"proxy":       proxy,
		"usage_logs":  usageLogs,
		"usage_count": len(usageLogs),
	})
}

// CreateProxy 创建代理
// @Summary 创建代理
// @Description 创建新的SK5代理
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ProxyCreateRequest true "代理信息"
// @Success 201 {object} response.Response{data=models.Proxy}
// @Router /proxies [post]
func (h *ProxyHandler) CreateProxy(c *gin.Context) {
	var req models.ProxyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 检查IP+Port是否已存在
	var existing models.Proxy
	if err := h.db.Where("ip = ? AND port = ?", req.IP, req.Port).First(&existing).Error; err == nil {
		response.Error(c, http.StatusBadRequest, "该代理已存在")
		return
	}

	// 查询IP地理位置
	location, _ := utils.QueryIPLocationWithFallback(req.IP)

	// 创建代理（先保存以获取ID）
	proxy := models.Proxy{
		IP:         req.IP,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		Province:   location.Province,
		City:       location.City,
		ISP:        location.ISP,
		Remark:     req.Remark,
		QRCodeURL:  "", // 先留空
		UsageCount: 0,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.db.Create(&proxy).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}

	// 生成 SOCKS5 分享链接
	proxy.QRCodeURL = generateSocksShareURL(proxy.IP, proxy.Port, proxy.Username, proxy.Password)

	// 更新二维码 URL
	h.db.Model(&proxy).Update("qr_code_url", proxy.QRCodeURL)

	response.SuccessWithMsg(c, "代理创建成功", proxy)
}

// UpdateProxy 更新代理
// @Summary 更新代理
// @Description 更新代理信息
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Param request body models.ProxyUpdateRequest true "代理信息"
// @Success 200 {object} response.Response
// @Router /proxies/{id} [put]
func (h *ProxyHandler) UpdateProxy(c *gin.Context) {
	id := c.Param("id")

	var req models.ProxyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "代理不存在")
		return
	}

	// 更新字段
	needRegenerateQRCode := false
	if req.IP != nil {
		proxy.IP = *req.IP
		needRegenerateQRCode = true
		// 重新查询地理位置
		location, _ := utils.QueryIPLocationWithFallback(*req.IP)
		proxy.Province = location.Province
		proxy.City = location.City
		proxy.ISP = location.ISP
	}
	if req.Port != nil {
		proxy.Port = *req.Port
		needRegenerateQRCode = true
	}
	if req.Username != nil {
		proxy.Username = *req.Username
		needRegenerateQRCode = true
	}
	if req.Password != nil {
		proxy.Password = *req.Password
		needRegenerateQRCode = true
	}
	if req.Remark != nil {
		proxy.Remark = *req.Remark
	}
	if req.IsActive != nil {
		proxy.IsActive = *req.IsActive
	}

	// 如果代理信息变更，重新生成二维码 URL
	if needRegenerateQRCode {
		proxy.QRCodeURL = generateSocksShareURL(proxy.IP, proxy.Port, proxy.Username, proxy.Password)
	}

	proxy.UpdatedAt = time.Now()

	if err := h.db.Save(&proxy).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	response.SuccessWithMsg(c, "代理更新成功", proxy)
}

// DeleteProxy 删除代理
// @Summary 删除代理
// @Description 删除代理（同时删除使用记录）
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "代理ID"
// @Success 200 {object} response.Response
// @Router /proxies/{id} [delete]
func (h *ProxyHandler) DeleteProxy(c *gin.Context) {
	id := c.Param("id")

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "代理不存在")
		return
	}

	// 删除使用记录
	h.db.Where("proxy_id = ?", id).Delete(&models.ProxyUsageLog{})

	// 删除代理
	if err := h.db.Delete(&proxy).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}

	response.SuccessWithMsg(c, "代理删除成功", nil)
}

// BatchDeleteProxies 批量删除代理
// @Summary 批量删除代理
// @Description 批量删除多个代理（同时删除使用记录）
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ProxyBatchDeleteRequest true "代理ID列表"
// @Success 200 {object} response.Response
// @Router /proxies/batch-delete [post]
func (h *ProxyHandler) BatchDeleteProxies(c *gin.Context) {
	var req models.ProxyBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		response.Error(c, http.StatusBadRequest, "请选择要删除的代理")
		return
	}

	// 删除使用记录
	h.db.Where("proxy_id IN ?", req.IDs).Delete(&models.ProxyUsageLog{})

	// 批量删除代理
	result := h.db.Where("id IN ?", req.IDs).Delete(&models.Proxy{})
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "批量删除失败")
		return
	}

	response.SuccessWithMsg(c, fmt.Sprintf("成功删除 %d 个代理", result.RowsAffected), nil)
}

// BatchImportProxies 批量导入代理
// @Summary 批量导入代理
// @Description 批量导入SK5代理，格式：IP|Port|Username|Password (每行一条)
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ProxyBatchImportRequest true "代理列表"
// @Success 200 {object} response.Response{data=object}
// @Router /proxies/batch-import [post]
func (h *ProxyHandler) BatchImportProxies(c *gin.Context) {
	var req models.ProxyBatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	lines := strings.Split(req.ProxyList, "\n")
	successCount := 0
	failCount := 0
	var errors []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析代理行
		ip, port, username, password, err := utils.ParseProxyLine(line)
		if err != nil {
			failCount++
			errors = append(errors, "第"+strconv.Itoa(i+1)+"行: "+err.Error())
			continue
		}

		// 检查是否已存在
		var existing models.Proxy
		if err := h.db.Where("ip = ? AND port = ?", ip, port).First(&existing).Error; err == nil {
			failCount++
			errors = append(errors, "第"+strconv.Itoa(i+1)+"行: 代理已存在")
			continue
		}

		// 查询IP地理位置（不阻塞）
		location, _ := utils.QueryIPLocationWithFallback(ip)

		// 创建代理（先保存以获取ID）
		proxy := models.Proxy{
			IP:         ip,
			Port:       port,
			Username:   username,
			Password:   password,
			Province:   location.Province,
			City:       location.City,
			ISP:        location.ISP,
			QRCodeURL:  "", // 先留空
			UsageCount: 0,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := h.db.Create(&proxy).Error; err != nil {
			failCount++
			errors = append(errors, "第"+strconv.Itoa(i+1)+"行: 创建失败")
			continue
		}

		// 生成 SOCKS5 分享链接
		proxy.QRCodeURL = generateSocksShareURL(proxy.IP, proxy.Port, proxy.Username, proxy.Password)

		// 更新二维码 URL
		h.db.Model(&proxy).Update("qr_code_url", proxy.QRCodeURL)

		successCount++
	}

	response.Success(c, gin.H{
		"success_count": successCount,
		"fail_count":    failCount,
		"errors":        errors,
		"message":       "批量导入完成",
	})
}

// AssignProxy 为设备分配代理（均衡分配算法）
// @Summary 为设备分配代理
// @Description 为设备分配SK5代理，采用最少使用次数优先的均衡分配算法
// @Tags 代理管理
// @Accept json
// @Produce json
// @Param request body models.ProxyAssignRequest true "设备信息"
// @Success 200 {object} response.Response{data=models.ProxyAssignResponse}
// @Router /proxies/assign [post]
func (h *ProxyHandler) AssignProxy(c *gin.Context) {
	var req models.ProxyAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 查找使用次数最少的激活代理（均衡分配）
	var proxy models.Proxy
	if err := h.db.Where("is_active = ?", true).
		Order("usage_count ASC, id ASC").
		First(&proxy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "暂无可用代理")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 更新使用次数
	proxy.UsageCount++
	proxy.UpdatedAt = time.Now()
	h.db.Save(&proxy)

	// 记录使用日志
	usageLog := models.ProxyUsageLog{
		ProxyID:    proxy.ID,
		DeviceID:   req.DeviceID,
		DeviceSN:   req.DeviceSN,
		IP:         proxy.IP,
		Port:       proxy.Port,
		AssignedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
	h.db.Create(&usageLog)

	// 返回代理信息
	assignResp := models.ProxyAssignResponse{
		IP:       proxy.IP,
		Port:     proxy.Port,
		Username: proxy.Username,
		Password: proxy.Password,
		ProxyID:  proxy.ID,
	}

	response.Success(c, assignResp)
}

// GetProxyStatistics 获取代理统计信息
// @Summary 获取代理统计信息
// @Description 获取代理池的统计数据
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.ProxyStatistics}
// @Router /proxies/statistics [get]
func (h *ProxyHandler) GetProxyStatistics(c *gin.Context) {
	var stats models.ProxyStatistics

	// 总数
	h.db.Model(&models.Proxy{}).Count(&stats.TotalCount)

	// 激活数量
	h.db.Model(&models.Proxy{}).Where("is_active = ?", true).Count(&stats.ActiveCount)

	// 未激活数量
	stats.InactiveCount = stats.TotalCount - stats.ActiveCount

	// 总使用次数
	h.db.Model(&models.Proxy{}).Select("COALESCE(SUM(usage_count), 0)").Scan(&stats.TotalUsage)

	// 平均使用次数
	if stats.TotalCount > 0 {
		stats.AvgUsage = float64(stats.TotalUsage) / float64(stats.TotalCount)
	}

	response.Success(c, stats)
}

// GetProxyUsageLogs 获取代理使用记录
// @Summary 获取代理使用记录
// @Description 获取代理分配使用记录
// @Tags 代理管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param proxy_id query int false "代理ID"
// @Param device_id query int false "设备ID"
// @Success 200 {object} response.Response{data=object}
// @Router /proxies/usage-logs [get]
func (h *ProxyHandler) GetProxyUsageLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	proxyID := c.Query("proxy_id")
	deviceID := c.Query("device_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := h.db.Model(&models.ProxyUsageLog{})

	if proxyID != "" {
		query = query.Where("proxy_id = ?", proxyID)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var total int64
	query.Count(&total)

	var logs []models.ProxyUsageLog
	query.Order("assigned_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs)

	response.Success(c, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetV2rayConfig 获取代理的 v2rayN JSON 配置
// @Summary 获取 v2rayN JSON 配置
// @Description 返回指定代理的 v2rayN JSON 配置文件
// @Tags 代理管理
// @Produce json
// @Param id path int true "代理ID"
// @Success 200 {object} object "v2rayN JSON 配置"
// @Router /proxies/{id}/v2ray-config [get]
func (h *ProxyHandler) GetV2rayConfig(c *gin.Context) {
	id := c.Param("id")

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构造 v2rayN 支持的 JSON 配置
	// 参考：https://github.com/2dust/v2rayN
	// v2rayN 支持 HTTP/SOCKS 代理配置
	config := gin.H{
		"log": gin.H{
			"loglevel": "warning",
		},
		"inbounds": []gin.H{
			{
				"port":     10808,
				"protocol": "socks",
				"settings": gin.H{
					"auth": "noauth",
					"udp":  true,
				},
			},
			{
				"port":     10809,
				"protocol": "http",
				"settings": gin.H{},
			},
		},
		"outbounds": []gin.H{
			{
				"protocol": "socks",
				"settings": gin.H{
					"servers": []gin.H{
						{
							"address": proxy.IP,
							"port":    proxy.Port,
							"users": []gin.H{
								{
									"user": proxy.Username,
									"pass": proxy.Password,
								},
							},
						},
					},
				},
				"tag": fmt.Sprintf("SOCKS5代理-%s", proxy.IP),
			},
		},
	}

	// 返回 JSON 配置
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"proxy-%s.json\"", proxy.IP))
	c.JSON(http.StatusOK, config)
}

// GetClashConfig 获取代理的 Clash YAML 配置
// @Summary 获取 Clash YAML 配置
// @Description 返回指定代理的 Clash/Mihomo YAML 配置文件
// @Tags 代理管理
// @Produce plain
// @Param id path int true "代理ID"
// @Success 200 {string} string "YAML 配置文件"
// @Router /proxies/{id}/clash-config [get]
func (h *ProxyHandler) GetClashConfig(c *gin.Context) {
	id := c.Param("id")

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "代理不存在")
			return
		}
		c.String(http.StatusInternalServerError, "查询失败")
		return
	}

	// 构造 Clash YAML 配置
	// 参考 Mihomo 文档: https://wiki.metacubex.one/en/config/proxy-providers/content/
	yamlConfig := fmt.Sprintf(`# Clash/Mihomo SOCKS5 Proxy Configuration
# Generated by JD Task Platform
# Proxy: %s:%d

proxies:
  - name: "SOCKS5代理-%s"
    type: socks5
    server: %s
    port: %d
    username: %s
    password: "%s"`,
		proxy.IP, proxy.Port,
		proxy.IP,
		proxy.IP,
		proxy.Port,
		proxy.Username,
		proxy.Password,
	)

	// 返回 YAML 文件
	c.Header("Content-Type", "text/yaml; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"proxy-%s.yaml\"", proxy.IP))
	c.String(http.StatusOK, yamlConfig)
}
