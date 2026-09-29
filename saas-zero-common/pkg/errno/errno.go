package errno

import (
	"errors"
	"fmt"
	"net/http"
)

type Errno struct {
	Code int
	Msg  string
}

func New(code int, msg string) *Errno {
	return &Errno{Code: code, Msg: msg}
}

func (e *Errno) Error() string {
	return e.Msg
}

// JSON returns {"code":N,"msg":"..."} for HTTP middleware responses.
func (e *Errno) JSON() string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, e.Code, e.Msg)
}

// ErrHandler 是统一的 HTTP 错误处理器（配合 httpx.SetErrorHandler 使用）。
// 所有响应 code 均取自本包错误码：
//   - errno.Errno 错误：使用其自身 code/msg
//   - 其他意外错误（DB/RPC/参数解析等）：统一映射为 InternalError
func ErrHandler(err error) (int, any) {
	var e *Errno
	if errors.As(err, &e) {
		return e.Code, map[string]any{"code": e.Code, "msg": e.Msg}
	}
	return http.StatusInternalServerError, map[string]any{
		"code": InternalError.Code,
		"msg":  InternalError.Msg,
	}
}

var (
	// 通用
	Success       = New(200, "success")
	InternalError = New(500, "内部服务器错误")
	InvalidParam  = New(400, "参数错误")
	Forbidden     = New(403, "无权限")

	// 认证 1001+
	TenantNotFound   = New(1001, "租户不存在")
	UserNotFound     = New(1002, "用户不存在")
	WrongPassword    = New(1003, "密码错误")
	TokenExpired     = New(1004, "token无效或已过期")
	OldPasswordWrong = New(1005, "旧密码错误")
	UserIdFormatErr  = New(1006, "用户ID格式错误")
	AccountLocked    = New(1007, "账号已锁定，请稍后再试")
	AccountDisabled  = New(1008, "账号已被禁用")

	// HTTP 状态码
	Unauthorized         = New(401, "未授权")
	MissingAuthHeader    = New(401, "missing authorization header")
	InvalidAuthHeader    = New(401, "invalid authorization header")
	InvalidToken         = New(401, "invalid token")
	TokenInvalidated     = New(401, "token invalidated")
	TokenVersionMismatch = New(401, "token version mismatch, please re-login")
	NoRoles              = New(403, "no roles")
	ForbiddenOperation   = New(403, "forbidden")
	AuthServiceUnavailable = New(500, "auth service unavailable, please try again later")

	// 业务校验 4xx
	HasSubDept     = New(400, "该部门下存在子部门，无法删除")
	HasSubMenu     = New(400, "该菜单下存在子菜单，无法删除")
	PackageInUse   = New(400, "该套餐已被租户使用，无法删除")
	UsernameExists = New(400, "用户名已存在")
)
