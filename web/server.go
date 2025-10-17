package web

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// Server Web服务器结构
type Server struct {
	db            *sql.DB
	mux           *http.ServeMux
	port          int
	configManager *ConfigManager
	csvParser     *CSVParser
}

// NewServer 创建新的Web服务器实例
func NewServer(db *sql.DB) *Server {
	server := &Server{
		db:            db,
		mux:           http.NewServeMux(),
		configManager: NewConfigManager(db),
		csvParser:     NewCSVParser(db),
	}
	server.setupRoutes()
	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件服务
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API 路由 - 配置管理
	s.mux.HandleFunc("/api/config", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleGetConfig,
		"PUT": s.handleSetConfig,
	}))
	s.mux.HandleFunc("/api/config/reset", s.methodRouter(map[string]http.HandlerFunc{
		"POST": s.handleResetConfig,
	}))

	// API 路由 - 文件上传
	s.mux.HandleFunc("/api/upload", s.methodRouter(map[string]http.HandlerFunc{
		"POST": s.handleUpload,
	}))
	s.mux.HandleFunc("/api/upload/status", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleUploadStatus,
	}))

	// API 路由 - 统计查询
	s.mux.HandleFunc("/api/stats/overview", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleStatsOverview,
	}))
	s.mux.HandleFunc("/api/stats/monthly/", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleMonthlyStats,
	}))
	s.mux.HandleFunc("/api/stats/categories", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleCategoryStats,
	}))

	// API 路由 - 交易管理
	s.mux.HandleFunc("/api/transactions", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleGetTransactions,
	}))
	s.mux.HandleFunc("/api/transactions/", s.methodRouter(map[string]http.HandlerFunc{
		"DELETE": s.handleDeleteTransaction,
		"PUT":    s.handleUpdateTransaction,
	}))

	// 页面路由
	s.mux.HandleFunc("/", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleIndex,
	}))
	s.mux.HandleFunc("/config", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleConfigPage,
	}))
	s.mux.HandleFunc("/upload", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleUploadPage,
	}))
	s.mux.HandleFunc("/stats", s.methodRouter(map[string]http.HandlerFunc{
		"GET": s.handleStatsPage,
	}))
}

// methodRouter 根据HTTP方法路由到不同的处理函数
func (s *Server) methodRouter(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := methods[r.Method]; ok {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// Start 启动Web服务器
func (s *Server) Start() error {
	// 获取配置的端口
	portStr, err := s.configManager.GetConfig("server.port")
	if err != nil {
		portStr = "0" // 默认随机端口
	}
	
	configPort, _ := strconv.Atoi(portStr)
	
	// 如果配置为0，则使用随机端口
	if configPort == 0 {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("failed to get random port: %w", err)
		}
		s.port = listener.Addr().(*net.TCPAddr).Port
		listener.Close()
	} else {
		s.port = configPort
	}

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Web服务器启动在端口 %d", s.port)
	log.Printf("访问地址: http://localhost:%d", s.port)

	// 自动打开浏览器
	go s.openBrowser()

	// 启动HTTP服务器
	server := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return server.ListenAndServe()
}

// GetPort 获取服务器端口
func (s *Server) GetPort() int {
	return s.port
}

// openBrowser 自动打开浏览器
func (s *Server) openBrowser() {
	// 检查配置是否启用自动打开浏览器
	autoOpen, err := s.configManager.GetConfig("auto.open.browser")
	if err != nil || autoOpen != "true" {
		return
	}

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	url := fmt.Sprintf("http://localhost:%d", s.port)
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("无法自动打开浏览器: %v", err)
		log.Printf("请手动访问: %s", url)
	} else {
		log.Printf("已自动打开浏览器: %s", url)
	}
}

// 这里需要引用主包中的配置函数
// 为了避免循环依赖，我们需要重新定义或者通过接口传递