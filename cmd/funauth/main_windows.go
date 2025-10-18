//go:build windows

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Yeah114/FunAuth/cmd/funauth/internal/router"
)

func main() {
	// 确保标准日志输出到 stdout（部分面板默认不抓取 stderr）
	log.SetOutput(os.Stdout)

	r := router.NewRouter()

	addr := os.Getenv("FUNAUTH_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Windows 下启动前尝试释放端口
	if p, ok := parsePort(addr); ok {
		log.Printf("[port] try free port %d before binding", p)
		freePortWindows(p)
	}

	log.Printf("[server] binding address: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func parsePort(addr string) (int, bool) {
	a := strings.TrimSpace(addr)
	if a == "" {
		return 0, false
	}
	if after, ok := strings.CutPrefix(a, ":"); ok {
		a = after
	} else if strings.Contains(a, ":") {
		idx := strings.LastIndex(a, ":")
		if idx >= 0 && idx+1 < len(a) {
			a = a[idx+1:]
		}
	}
	v, err := strconv.Atoi(a)
	if err != nil || v <= 0 || v > 65535 {
		return 0, false
	}
	return v, true
}

func freePortWindows(port int) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pids := findPidsWithNetstat(ctx, port)
	if len(pids) == 0 {
		log.Printf("[port] no owner found for %d", port)
		return
	}
	uniq := make(map[int]struct{})
	for _, p := range pids {
		uniq[p] = struct{}{}
	}
	for pid := range uniq {
		if pid <= 1 || pid == os.Getpid() {
			continue
		}
		log.Printf("[port] taskkill pid=%d for port %d", pid, port)
		// 直接强制结束进程及其子进程
		_ = exec.CommandContext(ctx, "taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
	}
}

func findPidsWithNetstat(ctx context.Context, port int) []int {
	cmd := exec.CommandContext(ctx, "netstat", "-ano")
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) == 0 {
		return nil
	}
	lines := strings.Split(string(out), "\n")
	var pids []int
	needle := ":" + strconv.Itoa(port)
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if !(strings.HasPrefix(ln, "TCP") || strings.HasPrefix(ln, "UDP")) {
			continue
		}
		if !strings.Contains(ln, needle) {
			continue
		}
		fields := strings.Fields(ln)
		if len(fields) < 4 {
			continue
		}
		pidField := fields[len(fields)-1]
		if v, err := strconv.Atoi(pidField); err == nil {
			pids = append(pids, v)
		}
	}
	return pids
}