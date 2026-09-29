package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/protobuf/proto"
)

// sensitiveKeys 记录时需要脱敏的请求/响应字段（登录与改密场景）。
var sensitiveKeys = []string{"password", "newPassword", "oldPassword", "resetPassword", "captchaVal", "token", "authorization", "mobile", "phone"}

const maxLogPayloadBytes = 64 * 1024

// OperationLog 记录所有写操作（非 GET）的审计日志。跳过 GET 与 /init/* 路由。
// 采集请求参数（脱敏）、响应体、HTTP 状态码与失败原因。
// 操作人/租户信息来自 JwtAuth 已注入的 context，因此必须注册在 JwtAuth 之后。
func OperationLog(cli apps.SysLogsClient) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || strings.HasPrefix(r.URL.Path, "/init/") {
				next(w, r)
				return
			}

			start := time.Now()

			// 通过包装器边读边捕获，避免无限制读取请求体；原始请求仍完整传给业务处理。
			var requestCapture *captureReadCloser
			if r.Body != nil {
				requestCapture = &captureReadCloser{ReadCloser: r.Body}
				r.Body = requestCapture
			}

			// 包装 ResponseWriter 以捕获状态码与响应体
			rw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next(rw, r)
			duration := time.Since(start).Milliseconds()
			requestBody := ""
			if requestCapture != nil {
				requestBody = requestCapture.logValue()
			}

			ctx := r.Context()
			module, operation := parseModuleOperation(r.URL.Path)

			status := responseStatus(rw.status, rw.body.String())
			errorMsg := extractErrorMsg(rw.body.String())
			if status == "failure" && errorMsg == "" {
				errorMsg = fmt.Sprintf("HTTP %d", rw.status)
			}
			requestParam := safeLogPayload(requestBody, 2000)
			if r.URL.RawQuery != "" {
				queryJSON, _ := json.Marshal(r.URL.Query())
				queryLog := safeLogPayload(string(queryJSON), 2000)
				if requestParam == "" {
					requestParam = queryLog
				} else {
					requestParam = truncate(requestParam+"\nquery="+queryLog, 2000)
				}
			}
			log := &apps.OperationLog{
				Module:        proto.String(module),
				Operation:     proto.String(operation),
				RequestMethod: proto.String(r.Method),
				RequestUrl:    proto.String(r.URL.Path),
				RequestParam:  proto.String(requestParam),
				ResponseData:  proto.String(rw.logValue()),
				Status:        proto.String(status),
				ErrorMsg:      proto.String(errorMsg),
				Duration:      proto.Int64(duration),
				OperatorIp:    proto.String(httpx.GetRemoteAddr(r)),
				OperatorId:    proto.Int64(mixins.GetCurrentUserId(ctx)),
				OperatorName:  proto.String(mixins.GetCurrentUserName(ctx)),
				TenantId:      proto.Int64(mixins.GetCurrentTenantId(ctx)),
			}

			// 异步写入，失败仅记录日志，不影响主流程。使用独立 context 避免请求结束后被取消。
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, err := cli.CreateOperationLog(bgCtx, log); err != nil {
					logx.Errorf("write operation log error: %v", err)
				}
			}()
		}
	}
}

// loggingResponseWriter 捕获响应状态码与响应体，供审计日志使用。
type loggingResponseWriter struct {
	http.ResponseWriter
	status    int
	body      bytes.Buffer
	truncated bool
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) logValue() string {
	if w.truncated {
		return "[payload omitted: too large]"
	}
	return safeLogPayload(w.body.String(), 4000)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.body.Len() <= maxLogPayloadBytes {
		remaining := maxLogPayloadBytes + 1 - w.body.Len()
		if len(b) > remaining {
			w.truncated = true
			_, _ = w.body.Write(b[:remaining])
		} else {
			_, _ = w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

type captureReadCloser struct {
	io.ReadCloser
	buf       bytes.Buffer
	truncated bool
}

func (r *captureReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if r.buf.Len() <= maxLogPayloadBytes {
		remaining := maxLogPayloadBytes + 1 - r.buf.Len()
		if n > remaining {
			r.truncated = true
			_, _ = r.buf.Write(p[:remaining])
		} else {
			_, _ = r.buf.Write(p[:n])
		}
	}
	return n, err
}

func (r *captureReadCloser) logValue() string {
	if r.truncated || r.buf.Len() > maxLogPayloadBytes {
		return "[payload omitted: too large]"
	}
	return r.buf.String()
}

// maskSensitive 将 JSON 中敏感字段的值替换为 "***"。
// 支持 key 精确匹配与嵌套（key 含密码/password 前缀）。非 JSON 原样返回。
func maskSensitive(s string) string {
	if s == "" {
		return s
	}
	var obj any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return "[non-json payload omitted]"
	}
	maskJSON(obj)
	out, err := json.Marshal(obj)
	if err != nil {
		return s
	}
	return string(out)
}

func safeLogPayload(s string, max int) string {
	if s == "" || strings.HasPrefix(s, "[payload omitted:") {
		return s
	}
	return truncate(maskSensitive(s), max)
}

func maskJSON(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if isSensitiveKey(k) {
				t[k] = "***"
				continue
			}
			maskJSON(val)
		}
	case []any:
		for _, item := range t {
			maskJSON(item)
		}
	}
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	for _, s := range sensitiveKeys {
		l := strings.ToLower(s)
		if lower == l || strings.HasPrefix(lower, l) {
			return true
		}
	}
	return false
}

// extractErrorMsg 从响应体中提取失败信息（status>=400 时的错误详情）。
// 返回响应 JSON 的 msg 字段；非失败响应返回空串。
func responseStatus(httpStatus int, body string) string {
	if httpStatus >= http.StatusBadRequest {
		return "failure"
	}
	if body != "" {
		var obj struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal([]byte(body), &obj); err == nil && obj.Code != 0 && obj.Code != http.StatusOK {
			return "failure"
		}
	}
	return "success"
}

func extractErrorMsg(body string) string {
	if body == "" {
		return ""
	}
	var obj struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal([]byte(body), &obj); err != nil || obj.Code == 0 || obj.Code == http.StatusOK {
		return ""
	}
	return truncate(obj.Msg, 500)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// parseModuleOperation 从请求路径推导模块名与操作名。
// 例：/system/user/create → ("user", "create")；/system/dict/data/list → ("dict", "list")
// 非 /system 前缀（如 /oauth/login）→ ("oauth", "login")。
func parseModuleOperation(path string) (module, operation string) {
	segs := make([]string, 0, 4)
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	switch len(segs) {
	case 0:
		return "", ""
	case 1:
		return segs[0], ""
	default:
		// /system/xxx/yyy → (xxx, yyy)；其余 /a/b → (a, b)
		if segs[0] == "system" && len(segs) > 2 {
			return segs[1], segs[len(segs)-1]
		}
		return segs[0], segs[len(segs)-1]
	}
}
