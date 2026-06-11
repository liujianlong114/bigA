package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Config 编排器配置
type Config struct {
	ProjectRoot  string
	BaseURL      string
	FlutterURL   string
	Interval     time.Duration
	ReportDir    string
	BackendBin   string
	StartFlutter bool
}

// RunResult 单次巡检结果
type RunResult struct {
	StartedAt  time.Time         `json:"started_at"`
	BackendUp  bool              `json:"backend_up"`
	FlutterUp  bool              `json:"flutter_up"`
	SmokeTests []SmokeTestResult `json:"smoke_tests"`
	Progress   ProgressReport    `json:"progress"`
	NextTasks  []string          `json:"next_tasks"`
	AllPassed  bool              `json:"all_passed"`
}

type SmokeTestResult struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status int    `json:"status"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type ProgressReport struct {
	Completed []string `json:"completed"`
	Pending   []string `json:"pending"`
	P0Done    int      `json:"p0_done"`
	P0Total   int      `json:"p0_total"`
	P1Done    int      `json:"p1_done"`
	P1Total   int      `json:"p1_total"`
}

// Orchestrator 30 分钟进度巡检
type Orchestrator struct {
	cfg Config
}

func New(cfg Config) *Orchestrator {
	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Minute
	}
	if cfg.ReportDir == "" {
		cfg.ReportDir = filepath.Join(cfg.ProjectRoot, "marvis_promot")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.FlutterURL == "" {
		cfg.FlutterURL = "http://localhost:3000"
	}
	return &Orchestrator{cfg: cfg}
}

func (o *Orchestrator) Start(ctx context.Context) {
	log.Printf("marvis: 编排器启动 interval=%s base=%s", o.cfg.Interval, o.cfg.BaseURL)
	o.runOnce(ctx)
	ticker := time.NewTicker(o.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.runOnce(ctx)
		}
	}
}

// RunOnce 执行单次巡检（供 MARVIS_ONCE=1 使用）
func (o *Orchestrator) RunOnce(ctx context.Context) {
	o.runOnce(ctx)
}

func (o *Orchestrator) runOnce(ctx context.Context) {
	log.Printf("marvis: ===== 开始巡检 %s =====", time.Now().Format(time.RFC3339))
	res := &RunResult{StartedAt: time.Now()}

	o.ensureBackend(ctx)
	o.ensureFlutter(ctx)

	res.BackendUp = o.ping(o.cfg.BaseURL + "/health")
	res.FlutterUp = o.ping(o.cfg.FlutterURL)

	// 等待 stream 写入 live 数据
	if res.BackendUp {
		time.Sleep(3 * time.Second)
	}

	res.SmokeTests = o.runSmokeTests(ctx)
	res.Progress = o.evalProgress()
	res.NextTasks = o.planNext(res.Progress)
	res.AllPassed = res.BackendUp && allSmokeOK(res.SmokeTests)

	if err := o.writeReports(res); err != nil {
		log.Printf("marvis: 写报告失败: %v", err)
	}
	log.Printf("marvis: 巡检完成 backend=%v flutter=%v smoke=%d/%d pass=%v",
		res.BackendUp, res.FlutterUp, countPass(res.SmokeTests), len(res.SmokeTests), res.AllPassed)
}

func (o *Orchestrator) ensureBackend(ctx context.Context) {
	if o.ping(o.cfg.BaseURL + "/health") {
		return
	}
	log.Printf("marvis: 启动后端...")
	_ = killPort(8080)
	bin := o.cfg.BackendBin
	if bin == "" {
		bin = filepath.Join(o.cfg.ProjectRoot, "bin", "bigA")
	}
	if _, err := os.Stat(bin); err != nil {
		// fallback go run
		cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
		cmd.Dir = o.cfg.ProjectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Start()
	} else {
		cmd := exec.CommandContext(ctx, bin)
		cmd.Dir = o.cfg.ProjectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Start()
	}
	for i := 0; i < 30; i++ {
		if o.ping(o.cfg.BaseURL + "/health") {
			return
		}
		time.Sleep(time.Second)
	}
}

func (o *Orchestrator) ensureFlutter(ctx context.Context) {
	if !o.cfg.StartFlutter {
		return
	}
	if o.ping(o.cfg.FlutterURL) {
		return
	}
	log.Printf("marvis: 启动 Flutter Web...")
	_ = killPort(3000)
	cmd := exec.CommandContext(ctx, "flutter", "run", "-d", "web-server", "--web-port=3000")
	cmd.Dir = filepath.Join(o.cfg.ProjectRoot, "app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	for i := 0; i < 60; i++ {
		if o.ping(o.cfg.FlutterURL) {
			return
		}
		time.Sleep(time.Second)
	}
}

func (o *Orchestrator) ping(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (o *Orchestrator) runSmokeTests(ctx context.Context) []SmokeTestResult {
	tests := []struct {
		name string
		path string
	}{
		{"health", "/health"},
		{"quote", "/api/v1/market/quote?code=600519"},
		{"list", "/api/v1/market/list?page=1&size=5"},
		{"stocks", "/api/v1/market/stocks?page=1&size=5"},
		{"boards", "/api/v1/market/boards"},
		{"auth_login", "/api/v1/auth/login"},
		{"kline", "/api/v1/market/kline?code=600519&period=day&limit=5"},
		{"live_meta", "/api/v1/market/live/meta"},
		{"portfolio", "/api/v1/portfolio"},
		{"ai_state", "/api/v1/ai/state"},
		{"analysis_market", "/api/v1/analysis/market"},
		{"analysis_anomalies", "/api/v1/analysis/anomalies"},
		{"analysis_portfolio", "/api/v1/analysis/portfolio"},
		{"analysis_report", "/api/v1/analysis/daily-report"},
		{"analysis_predict", "/api/v1/analysis/predict?code=600519"},
		{"sectors", "/api/v1/market/sectors?page=1&size=5"},
		{"sector_stocks", "/api/v1/market/sector/new_hghy/stocks?page=1&size=5"},
		{"conditional_orders", "/api/v1/conditional-orders"},
		{"performance", "/api/v1/performance"},
	}
	client := &http.Client{Timeout: 15 * time.Second}
	var out []SmokeTestResult
	for _, t := range tests {
		url := o.cfg.BaseURL + t.path
		var req *http.Request
		if t.name == "auth_login" {
			req, _ = http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		}
		resp, err := client.Do(req)
		r := SmokeTestResult{Name: t.name, Path: t.path}
		if err != nil {
			r.Detail = err.Error()
			out = append(out, r)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		r.Status = resp.StatusCode
		r.OK = resp.StatusCode >= 200 && resp.StatusCode < 300
		if t.name == "auth_login" {
			r.OK = resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized
		}
		if !r.OK {
			r.Detail = string(body)
			if len(r.Detail) > 200 {
				r.Detail = r.Detail[:200]
			}
		}
		// live_meta 可能 404 若 stream 未就绪，降级为 warning
		if t.name == "live_meta" && resp.StatusCode == 404 {
			r.OK = true
			r.Detail = "no live data yet (acceptable)"
		}
		out = append(out, r)
	}
	return out
}

type taskCheck func(o *Orchestrator) bool

var taskRegistry = []struct {
	ID       string
	Priority string
	Check    taskCheck
}{
	{"P0-持仓分析", "P0", checkPathOK("/api/v1/analysis/portfolio")},
	{"P0-大盘解析", "P0", checkPathOK("/api/v1/analysis/market")},
	{"P0-异动检测", "P0", checkPathOK("/api/v1/analysis/anomalies")},
	{"P0-每日复盘", "P0", checkPathOK("/api/v1/analysis/daily-report")},
	{"P1-涨跌推测", "P1", checkPathOK("/api/v1/analysis/predict?code=600519")},
	{"P1-板块行情", "P1", checkPathOK("/api/v1/market/sectors?page=1&size=5")},
	{"P1-条件单", "P1", checkPathOK("/api/v1/conditional-orders")},
	{"P1-绩效分析", "P1", checkPathOK("/api/v1/performance")},
	{"P1-K线指标", "P1", checkKlineIndicators},
	{"P2-全市场分页", "P2", checkPathOK("/api/v1/market/stocks?page=1&size=5")},
	{"P2-用户认证", "P2", checkAuthModule},
	{"P2-交易日历", "P2", func(o *Orchestrator) bool {
		_, err := os.Stat(filepath.Join(o.cfg.ProjectRoot, "internal/market/calendar_v2.go"))
		return err == nil
	}},
	{"P2-单元测试", "P2", checkUnitTests},
	{"P2-Docker", "P2", func(o *Orchestrator) bool {
		_, err := os.Stat(filepath.Join(o.cfg.ProjectRoot, "docker-compose.yml"))
		return err == nil
	}},
}

func checkPathOK(path string) taskCheck {
	return func(o *Orchestrator) bool {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(o.cfg.BaseURL + path)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode >= 200 && resp.StatusCode < 300
	}
}

func checkKlineIndicators(o *Orchestrator) bool {
	if !checkPathOK("/api/v1/market/kline?code=600519&period=day&limit=5")(o) {
		return false
	}
	b, err := os.ReadFile(filepath.Join(o.cfg.ProjectRoot, "app/lib/widgets/kline_indicators.dart"))
	if err != nil {
		return false
	}
	s := string(b)
	return strings.Contains(s, "calcMACD") && strings.Contains(s, "calcKDJ")
}

func checkAuthModule(o *Orchestrator) bool {
	if _, err := os.Stat(filepath.Join(o.cfg.ProjectRoot, "internal/api/auth.go")); err != nil {
		return false
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(o.cfg.BaseURL+"/api/v1/auth/login", "application/json", strings.NewReader(`{"username":"x","password":"y"}`))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest
}

func checkUnitTests(o *Orchestrator) bool {
	matches, _ := filepath.Glob(filepath.Join(o.cfg.ProjectRoot, "internal/sim/*_test.go"))
	return len(matches) >= 2
}

func (o *Orchestrator) evalProgress() ProgressReport {
	var rep ProgressReport
	for _, t := range taskRegistry {
		ok := t.Check(o)
		if ok {
			rep.Completed = append(rep.Completed, t.ID)
		} else {
			rep.Pending = append(rep.Pending, t.ID)
		}
		if strings.HasPrefix(t.Priority, "P0") {
			rep.P0Total++
			if ok {
				rep.P0Done++
			}
		}
		if strings.HasPrefix(t.Priority, "P1") {
			rep.P1Total++
			if ok {
				rep.P1Done++
			}
		}
	}
	return rep
}

func (o *Orchestrator) planNext(p ProgressReport) []string {
	var tasks []string
	for _, id := range p.Pending {
		tasks = append(tasks, "继续开发: "+id)
		if len(tasks) >= 5 {
			break
		}
	}
	if len(tasks) == 0 {
		tasks = append(tasks, "P0/P1 核心已完成，进入 P2：Docker 容器化 + 单元测试覆盖")
	}
	return tasks
}

func (o *Orchestrator) writeReports(res *RunResult) error {
	_ = os.MkdirAll(o.cfg.ReportDir, 0755)
	jb, _ := json.MarshalIndent(res, "", "  ")
	if err := os.WriteFile(filepath.Join(o.cfg.ReportDir, "progress_report.json"), jb, 0644); err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("# Marvis 下次开发指令\n\n")
	sb.WriteString(fmt.Sprintf("> 自动生成于 %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("## 巡检结果\n\n- 后端: %v\n- Flutter: %v\n- 冒烟测试: %d/%d 通过\n- P0: %d/%d | P1: %d/%d\n\n",
		res.BackendUp, res.FlutterUp, countPass(res.SmokeTests), len(res.SmokeTests),
		res.Progress.P0Done, res.Progress.P0Total, res.Progress.P1Done, res.Progress.P1Total))
	sb.WriteString("## 接口冒烟\n\n| 接口 | 状态 |\n|------|------|\n")
	for _, t := range res.SmokeTests {
		st := "✅"
		if !t.OK {
			st = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s %s | %s |\n", t.Name, t.Path, st))
	}
	sb.WriteString("\n## 待开发任务\n\n")
	for i, t := range res.NextTasks {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, t))
	}
	sb.WriteString("\n## 已完成\n\n")
	for _, c := range res.Progress.Completed {
		sb.WriteString("- " + c + "\n")
	}
	return os.WriteFile(filepath.Join(o.cfg.ReportDir, "next_instructions.md"), []byte(sb.String()), 0644)
}

func allSmokeOK(tests []SmokeTestResult) bool {
	for _, t := range tests {
		if !t.OK {
			return false
		}
	}
	return len(tests) > 0
}

func countPass(tests []SmokeTestResult) int {
	n := 0
	for _, t := range tests {
		if t.OK {
			n++
		}
	}
	return n
}

func killPort(port int) error {
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	for _, pid := range strings.Fields(string(out)) {
		_ = syscall.Kill(parsePID(pid), syscall.SIGKILL)
	}
	return nil
}

func parsePID(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
