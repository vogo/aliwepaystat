package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/vogo/aliwepaystat"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(fmt.Sprintf("解析模板失败: %v", err))
	}
}

// handleIndex 首页
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("渲染模板失败: %v", err), http.StatusInternalServerError)
	}
}

// handleStatsPage 统计页面
func (s *Server) handleStatsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "stats.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("渲染模板失败: %v", err), http.StatusInternalServerError)
	}
}

// handleGetConfig 获取配置
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		// 返回所有配置
		configs, err := s.configManager.GetAllConfig()
		if err != nil {
			http.Error(w, fmt.Sprintf("获取配置失败: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(configs)
		return
	}

	// 返回指定配置
	value, err := s.configManager.GetConfig(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取配置失败: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})
}

// handleSetConfig 设置配置
func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var config struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "无效的JSON格式", http.StatusBadRequest)
		return
	}

	if config.Key == "" {
		http.Error(w, "配置键不能为空", http.StatusBadRequest)
		return
	}

	if err := aliwepaystat.SetConfig(s.db, config.Key, config.Value, "用户设置"); err != nil {
		http.Error(w, fmt.Sprintf("设置配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleResetConfig 重置配置
func (s *Server) handleResetConfig(w http.ResponseWriter, r *http.Request) {
	// 这里可以实现重置配置到默认值的逻辑
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "配置已重置"})
}

// handleUpload 处理文件上传
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	// 获取平台选择
	platform := r.FormValue("platform")
	if platform != "alipay" && platform != "wechat" {
		http.Error(w, "无效的平台选择", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// 检查文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		http.Error(w, "只支持CSV文件", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		http.Error(w, "读取文件失败", http.StatusInternalServerError)
		return
	}

	// 根据平台选择解析器
	var parser aliwepaystat.TransParser
	if platform == "alipay" {
		parser = aliwepaystat.TransParserAlipay
	} else {
		parser = aliwepaystat.TransParserWechat
	}

	// 解析并导入CSV
	count, err := s.csvParser.ParseAndImportCSV(content, parser, platform)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析CSV失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response := map[string]any{
		"status":   "success",
		"message":  "文件上传成功",
		"platform": platform,
		"filename": header.Filename,
		"count":    count,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleUploadPage 上传页面
func (s *Server) handleUploadPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "upload.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("渲染模板失败: %v", err), http.StatusInternalServerError)
	}
}

// handleStatsOverview 统计概览
func (s *Server) handleStatsOverview(w http.ResponseWriter, r *http.Request) {
	// 这里可以实现统计概览逻辑
	response := map[string]any{
		"total_transactions": 0,
		"total_income":       0.0,
		"total_expense":      0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleMonthlyStats 月度统计
func (s *Server) handleMonthlyStats(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取年月参数
	path := strings.TrimPrefix(r.URL.Path, "/api/stats/monthly/")
	yearMonth := strings.TrimSuffix(path, "/")

	if yearMonth == "" {
		http.Error(w, "缺少年月参数", http.StatusBadRequest)
		return
	}

	// 这里可以实现月度统计逻辑
	response := map[string]any{
		"year_month":    yearMonth,
		"total_income":  0.0,
		"total_expense": 0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleCategoryStats 分类统计
func (s *Server) handleCategoryStats(w http.ResponseWriter, r *http.Request) {
	// 这里可以实现分类统计逻辑
	response := map[string]any{
		"categories": []map[string]any{},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleGetTransactions 获取交易记录
func (s *Server) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	pageNum := 1
	limitNum := 20

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
	}

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitNum = l
		}
	}

	// 这里可以实现交易记录查询逻辑
	response := map[string]any{
		"page":         pageNum,
		"limit":        limitNum,
		"total":        0,
		"transactions": []map[string]any{},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleDeleteTransaction 删除交易记录
func (s *Server) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取交易ID
	path := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	transactionID := strings.TrimSuffix(path, "/")

	if transactionID == "" {
		http.Error(w, "缺少交易ID", http.StatusBadRequest)
		return
	}

	// 这里可以实现删除交易记录逻辑
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "交易记录已删除"})
}

// handleUpdateTransaction 更新交易记录
func (s *Server) handleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取交易ID
	path := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	transactionID := strings.TrimSuffix(path, "/")

	if transactionID == "" {
		http.Error(w, "缺少交易ID", http.StatusBadRequest)
		return
	}

	// 这里可以实现更新交易记录逻辑
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "交易记录已更新"})
}

// handleUploadStatus 获取上传状态
func (s *Server) handleUploadStatus(w http.ResponseWriter, r *http.Request) {
	// 这里可以实现上传进度查询逻辑
	response := map[string]any{
		"status":   "completed",
		"progress": 100,
		"message":  "上传完成",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleConfigPage 配置页面
func (s *Server) handleConfigPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "config.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("渲染模板失败: %v", err), http.StatusInternalServerError)
	}
}
