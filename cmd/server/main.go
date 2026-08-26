package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/web"
)

func main() {
	addr := flag.String("addr", "", "监听地址")
	dataDir := flag.String("data-dir", "./data", "数据目录")
	selfcheck := flag.Bool("selfcheck", false, "运行端到端自检后退出")
	selfcheckTimeout := flag.Duration("selfcheck-timeout", 20*time.Second, "自检超时")
	flag.Parse()
	listenAddr, err := resolveAddress(*addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *selfcheck {
		temporary, err := os.MkdirTemp("", "retirement-selfcheck-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer os.RemoveAll(temporary)
		*dataDir = temporary
	}
	store, err := persistence.Open(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	service := application.NewService(store)
	server := &http.Server{Addr: listenAddr, Handler: web.NewServer(service), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	if *selfcheck {
		if err := runSelfcheck(server, listenAddr, *selfcheckTimeout); err != nil {
			fmt.Fprintln(os.Stderr, "自检失败:", err)
			os.Exit(1)
		}
		return
	}
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "监听失败:", err)
		os.Exit(1)
	}
	defer listener.Close()
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, "服务退出:", err)
		os.Exit(1)
	}
}

func resolveAddress(flagValue string) (string, error) {
	if flagValue != "" {
		return validateAddress(flagValue)
	}
	if port := os.Getenv("PORT"); port != "" {
		return validateAddress("127.0.0.1:" + port)
	}
	return "127.0.0.1:19081", nil
}

func validateAddress(value string) (string, error) {
	host, port, err := net.SplitHostPort(value)
	if err != nil || host == "" || port == "" {
		return "", fmt.Errorf("监听地址必须包含明确主机和端口: %s", value)
	}
	if host == "0.0.0.0" || host == "::" || host == "" {
		return "", errors.New("禁止绑定通配地址")
	}
	if strings.HasPrefix(host, "127.") == false && host != "localhost" {
		return "", errors.New("监听地址必须使用回环主机")
	}
	var portNumber int
	if _, err := fmt.Sscanf(port, "%d", &portNumber); err != nil || portNumber < 1024 || portNumber > 65535 {
		return "", errors.New("端口必须是 1024-65535")
	}
	return value, nil
}

func runSelfcheck(server *http.Server, addr string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	client := &http.Client{Timeout: timeout / 3}
	base := "http://" + addr
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var created struct {
		Case struct {
			ID       string `json:"id"`
			Revision int64  `json:"revision"`
		} `json:"case"`
	}
	createPayload := map[string]any{"actor_id": "admin-01", "role": "safety_admin", "request_id": "self-create", "expected_revision": 0, "site": "化学楼 B203", "owner_id": "owner-01", "reason": "试剂过期退役", "planned_date": "2030-01-02", "materials": []map[string]any{{"material_code": "MAT-001", "display_name": "示例溶剂", "hazard_class": "general", "declared_quantity": 2, "unit": "L", "package_condition": "intact", "disposal_method": "中和"}}}
	if err := callJSON(ctx, client, base+"/api/cases", http.MethodPost, createPayload, &created); err != nil {
		return err
	}
	id, revision := created.Case.ID, created.Case.Revision
	count := func(request, counter, role string) error {
		payload := map[string]any{"actor_id": counter, "role": role, "request_id": request, "expected_revision": revision, "counter_id": counter, "observations": []map[string]any{{"material_code": "MAT-001", "quantity": 2, "package_condition": "intact"}}}
		var result struct {
			Case struct {
				Revision int64 `json:"revision"`
			} `json:"case"`
		}
		if err := callJSON(ctx, client, base+"/api/cases/"+id+"/count", http.MethodPost, payload, &result); err != nil {
			return err
		}
		revision = result.Case.Revision
		return nil
	}
	if err := count("self-count-1", "counter-01", "material_owner"); err != nil {
		return err
	}
	if err := count("self-count-2", "counter-02", "witness"); err != nil {
		return err
	}
	post := func(action, request, role, actor string, payload map[string]any) error {
		payload["actor_id"], payload["role"], payload["request_id"], payload["expected_revision"] = actor, role, request, revision
		var result struct {
			Case struct {
				Revision int64 `json:"revision"`
			} `json:"case"`
		}
		if err := callJSON(ctx, client, base+"/api/cases/"+id+"/"+action, http.MethodPost, payload, &result); err != nil {
			return err
		}
		revision = result.Case.Revision
		return nil
	}
	if err := post("risk", "self-risk", "safety_admin", "admin-01", map[string]any{"site_conditions": []string{}, "protective_measures": []string{}}); err != nil {
		return err
	}
	if err := post("review", "self-review", "compliance_reviewer", "reviewer-01", map[string]any{"reviewer_id": "reviewer-01", "approved": true}); err != nil {
		return err
	}
	start := time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339)
	end := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	if err := post("destruction", "self-destroy", "witness", "witness-01", map[string]any{"method": "中和", "started_at": start, "completed_at": end, "witness_ids": []string{"witness-01", "witness-02"}, "evidence_digest": "sha256:self-check"}); err != nil {
		return err
	}
	if err := post("verification", "self-verify", "compliance_reviewer", "reviewer-01", map[string]any{"check_name": "残留检测", "threshold": 1, "measured_value": 0}); err != nil {
		return err
	}
	var preview struct {
		PreviewDigest string `json:"preview_digest"`
		ChainHead     string `json:"chain_head"`
	}
	if err := callJSON(ctx, client, base+"/api/cases/"+id+"/archive-preview", http.MethodGet, nil, &preview); err != nil {
		return err
	}
	if err := post("archive", "self-archive", "compliance_reviewer", "reviewer-01", map[string]any{"preview_digest": preview.PreviewDigest, "expected_chain_head": preview.ChainHead}); err != nil {
		return err
	}
	var detail struct {
		Case struct {
			Status string `json:"status"`
		} `json:"case"`
		ChainOK bool `json:"chain_ok"`
	}
	if err := callJSON(ctx, client, base+"/api/cases/"+id, http.MethodGet, nil, &detail); err != nil {
		return err
	}
	if detail.Case.Status != "archived" || !detail.ChainOK {
		return fmt.Errorf("归档状态或审计链不正确")
	}
	var immutablePayload = map[string]any{"actor_id": "reviewer-01", "role": "compliance_reviewer", "request_id": "self-immutable", "expected_revision": revision}
	if err := callJSON(ctx, client, base+"/api/cases/"+id+"/archive", http.MethodPost, immutablePayload, nil); err == nil {
		return fmt.Errorf("归档批次仍可修改")
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
	select {
	case serveErr := <-serveDone:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func callJSON(ctx context.Context, client *http.Client, url, method string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(data))
	}
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	if response.StatusCode >= 300 {
		return fmt.Errorf("%s %s: %s", method, url, strings.TrimSpace(string(data)))
	}
	if target != nil && len(data) > 0 {
		return json.Unmarshal(data, target)
	}
	return nil
}
