package sysuserslogic

import (
	"context"
	"testing"

	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

func systemCtx() context.Context {
	return mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 1), "system"),
		0,
	)
}

// TestResetPassword_CrossTenantTarget 验证：A 租户管理员不能重置 B 租户用户的密码。
func TestResetPassword_CrossTenantTarget(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)

	// 在租户 A 创建一个普通用户
	createLogic := &CreateUserLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	resp, err := createLogic.CreateUser(&apps.UserReq{
		Username: proto.String("tenant-a-user"),
		Password: proto.String("pass123"),
		Nickname: proto.String("租户A用户"),
		Status:   proto.String("active"),
	})
	if err != nil {
		t.Fatalf("create tenant A user failed: %v", err)
	}
	targetA := resp.GetData().GetId()

	// 创建租户 B 及其用户
	sysCtx := systemCtx()
	pkg, err := client.SysPackage.Query().First(sysCtx)
	if err != nil {
		t.Fatalf("load package failed: %v", err)
	}
	tenantB, err := client.SysTenant.Create().
		SetName("测试租户B").
		SetCode("test-b-iso").
		SetAdminID(1).
		SetPackageID(pkg.ID).
		SetStatus(systenant.StatusActive).
		Save(sysCtx)
	if err != nil {
		t.Fatalf("create tenant B failed: %v", err)
	}
	ctxB := testCtx(tenantB.ID)
	createLogicB := &CreateUserLogic{ctx: ctxB, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctxB)}
	// 租户 B 管理员的 user id 沿用 context 中的 1，避免与 B 内真实用户混淆——这里直接创建"目标用户"
	targetB, err := createLogicB.CreateUser(&apps.UserReq{
		Username: proto.String("tenant-b-user"),
		Password: proto.String("pass456"),
		Nickname: proto.String("租户B用户"),
		Status:   proto.String("active"),
	})
	if err != nil {
		t.Fatalf("create tenant B user failed: %v", err)
	}

	// 租户 A 的操作者尝试重置租户 B 用户 → 必须失败（UserNotFound，因为按 A 租户查询不到）
	resetLogic := &ResetPasswordLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	_, err = resetLogic.ResetPassword(&apps.UserReq{
		Id:       proto.Int64(targetB.GetData().GetId()),
		Password: proto.String("hacked123"),
	})
	if err == nil {
		t.Fatal("cross-tenant reset must fail")
	}
	if e, ok := err.(*errno.Errno); !ok || e.Code != errno.UserNotFound.Code {
		t.Fatalf("expected UserNotFound(%d), got %v", errno.UserNotFound.Code, err)
	}

	// 同租户内正常重置仍应成功
	_, err = resetLogic.ResetPassword(&apps.UserReq{
		Id:       proto.Int64(targetA),
		Password: proto.String("newpass123"),
	})
	if err != nil {
		t.Fatalf("same-tenant reset should succeed: %v", err)
	}
	u, err := client.SysUser.Get(ctx, targetA)
	if err != nil {
		t.Fatalf("fetch user failed: %v", err)
	}
	if u.Password == "" || u.Password == "newpass123" {
		t.Fatal("password should have been hashed")
	}
}

// TestUpdateUser_CrossTenantTarget 验证：不能跨租户按全局 ID 更新用户。
func TestUpdateUser_CrossTenantTarget(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)
	createLogic := &CreateUserLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	resp, err := createLogic.CreateUser(&apps.UserReq{
		Username: proto.String("iso-user-a"),
		Password: proto.String("pass123"),
		Nickname: proto.String("租户A"),
		Status:   proto.String("active"),
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	u := resp.GetData()

	// 伪造 B 租户操作者 + 目标 A 租户的 user id
	ctxB := mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 99), "tenant-b-admin"),
		tenantID+1000,
	)
	updateLogic := &UpdateUserLogic{ctx: ctxB, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctxB)}
	_, err = updateLogic.UpdateUser(&apps.UserReq{
		Id:       proto.Int64(u.GetId()),
		Nickname: proto.String("被跨租户改"),
	})
	// 因为目标 id 在 B 租户 grep 不到，应报 UserNotFound
	if err == nil {
		t.Fatal("cross-tenant update must fail")
	}
	if e, ok := err.(*errno.Errno); !ok || e.Code != errno.UserNotFound.Code {
		t.Fatalf("expected UserNotFound(%d), got %v", errno.UserNotFound.Code, err)
	}
}

// TestDeleteUser_CrossTenantTarget 验证：软删除也必须按租户过滤，跨租户 ID 不会误删。
func TestDeleteUser_CrossTenantTarget(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)
	createLogic := &CreateUserLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	resp, err := createLogic.CreateUser(&apps.UserReq{
		Username: proto.String("del-user"),
		Password: proto.String("pass123"),
		Nickname: proto.String("待删用户"),
		Status:   proto.String("active"),
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	uid := resp.GetData().GetId()

	ctxB := mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 99), "tenant-b-admin"),
		tenantID+1000,
	)
	deleteLogic := &DeleteUserLogic{ctx: ctxB, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctxB)}
	if _, err := deleteLogic.DeleteUser(&apps.IdsReq{Ids: []int64{uid}}); err == nil {
		t.Fatal("cross-tenant delete must be rejected")
	}
	// 用户应仍然存在且未删除
	u, err := client.SysUser.TenantQuery(tenantID).Where(sysuser.IDEQ(uid)).Only(ctx)
	if err != nil {
		t.Fatalf("tenant A user should still exist: %v", err)
	}
	if !u.DeletedAt.IsZero() {
		t.Fatal("cross-tenant delete must NOT soft-delete the user")
	}
}

// TestAssignRoles_CrossTenantRole 验证：不能把其他租户的角色分配给当前租户用户。
func TestAssignRoles_CrossTenantRole(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)
	createLogic := &CreateUserLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	resp, err := createLogic.CreateUser(&apps.UserReq{
		Username: proto.String("assign-target"),
		Password: proto.String("pass123"),
		Nickname: proto.String("目标用户"),
		Status:   proto.String("active"),
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	uid := resp.GetData().GetId()

	// 创建租户 B 与 B 角色、B 用户（管理员角色）
	sysCtx := systemCtx()
	pkg, err := client.SysPackage.Query().First(sysCtx)
	if err != nil {
		t.Fatalf("load package failed: %v", err)
	}
	tenantB, err := client.SysTenant.Create().
		SetName("测试租户B").
		SetCode("test-b-iso2").
		SetAdminID(1).
		SetPackageID(pkg.ID).
		SetStatus(systenant.StatusActive).
		Save(sysCtx)
	if err != nil {
		t.Fatalf("create tenant B failed: %v", err)
	}
	bCtx := testCtx(tenantB.ID)
	roleB, err := client.SysRole.Create().
		SetName("B租户角色").
		SetCode("role-b").
		SetSort(1).
		SetStatus(sysrole.StatusActive).
		Save(bCtx)
	if err != nil {
		t.Fatalf("create role B failed: %v", err)
	}

	assignLogic := &AssignRolesLogic{ctx: ctx, svcCtx: &svc.ServiceContext{DB: client}, Logger: logx.WithContext(ctx)}
	_, err = assignLogic.AssignRoles(&apps.UserReq{
		Id:      proto.Int64(uid),
		RoleIds: []int64{roleB.ID},
	})
	if err == nil {
		t.Fatal("assigning cross-tenant role must fail")
	}
}
