package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Yeah114/FunAuth/internal/router"
)

func main() {
	// 确保标准日志输出到 stdout（部分面板默认不抓取 stderr）
	log.SetOutput(os.Stdout)

	r := router.NewRouter()

	addr := os.Getenv("FUNAUTH_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Linux 下启动前尝试释放端口
	if runtime.GOOS == "linux" {
		if p, ok := parsePort(addr); ok {
			log.Printf("[port] try free port %d before binding", p)
			freePortLinux(p)
		}
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

func freePortLinux(port int) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pids := findPidsWithSS(ctx, port)
	if len(pids) == 0 {
		pids = findPidsWithLsof(ctx, port)
	}
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
		log.Printf("[port] sending SIGTERM to pid=%d for port %d", pid, port)
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	time.Sleep(1200 * time.Millisecond)
	for pid := range uniq {
		if pid <= 1 || pid == os.Getpid() {
			continue
		}
		if alive(pid) {
			log.Printf("[port] sending SIGKILL to pid=%d for port %d", pid, port)
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}

func alive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

func findPidsWithSS(ctx context.Context, port int) []int {
	cmd := exec.CommandContext(ctx, "ss", "-lntp")
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) == 0 {
		return nil
	}
	lines := strings.Split(string(out), "\n")
	var pids []int
	needle := ":" + strconv.Itoa(port) + " "
	for _, ln := range lines {
		if !strings.Contains(ln, needle) {
			continue
		}
		s := ln
		for {
			idx := strings.Index(s, "pid=")
			if idx < 0 {
				break
			}
			s = s[idx+4:]
			j := 0
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			if j > 0 {
				if v, err := strconv.Atoi(s[:j]); err == nil {
					pids = append(pids, v)
				}
				s = s[j:]
			} else {
				break
			}
		}
	}
	return pids
}

func findPidsWithLsof(ctx context.Context, port int) []int {
	argsSets := [][]string{
		{"-t", "-iTCP:" + strconv.Itoa(port), "-sTCP:LISTEN"},
		{"-t", "-i:" + strconv.Itoa(port)},
	}
	for _, args := range argsSets {
		cmd := exec.CommandContext(ctx, "lsof", args...)
		out, err := cmd.CombinedOutput()
		if err != nil || len(out) == 0 {
			continue
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var pids []int
		for _, ln := range lines {
			if v, err := strconv.Atoi(strings.TrimSpace(ln)); err == nil {
				pids = append(pids, v)
			}
		}
		if len(pids) > 0 {
			return pids
		}
	}
	return nil
}
