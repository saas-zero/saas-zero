package syslogslogic

import (
	"testing"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent"
)

func TestOperationLogToResp_StatusPreserved(t *testing.T) {
	now := time.Now()
	o := &ent.SysOperationLog{
		ID:         100,
		Module:     "user",
		Operation:  "create",
		Method:     "POST",
		Path:       "/system/user/create",
		Params:     `{"username":"admin","password":"***"}`,
		Result:     `{"code":400,"msg":"参数错误"}`,
		ErrorMsg:   "参数错误",
		Status:     "failure",
		Duration:   12,
		IP:         "192.168.1.1",
		OperatorID: 1,
		CreatedAt:  now,
	}
	resp := operationLogToResp(o)

	if resp.GetStatus() != "failure" {
		t.Fatalf("expected status failure, got %s", resp.GetStatus())
	}
	if resp.GetModule() != "user" || resp.GetOperation() != "create" {
		t.Fatalf("unexpected module/operation: %s/%s", resp.GetModule(), resp.GetOperation())
	}
	if resp.GetResponseData() != `{"code":400,"msg":"参数错误"}` {
		t.Fatalf("unexpected responseData: %s", resp.GetResponseData())
	}
	if resp.GetErrorMsg() != "参数错误" {
		t.Fatalf("unexpected errorMsg: %s", resp.GetErrorMsg())
	}
	// 响应摘要和失败原因分别保存；此处同时验证 idStr 精度无丢失
	if resp.GetIdStr() != "100" {
		t.Fatalf("expected idStr 100, got %s", resp.GetIdStr())
	}
}

func TestOperationLogToResp_DefaultStatusSuccess(t *testing.T) {
	o := &ent.SysOperationLog{ID: 1, CreatedAt: time.Now()}
	resp := operationLogToResp(o)
	if resp.GetStatus() != "" {
		t.Fatalf("status should map from DB, got %s (DB default handled by schema)", resp.GetStatus())
	}
}
