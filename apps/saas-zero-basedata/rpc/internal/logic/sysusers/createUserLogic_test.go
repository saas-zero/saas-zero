package sysuserslogic

import (
	"context"
	stdsql "database/sql"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/migrate"
	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite"
)

// testCtx returns a context with test user/tenant info injected.
// testCtx returns a context with test user/tenant info injected.
// The tenant id comes from the freshly seeded client because BaseMixin
// overrides any manual ID with a snowflake ID.
func testCtx(tenantID int64) context.Context {
	return mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 1), "test-admin"),
		tenantID,
	)
}

// seedTestData creates the package + tenant rows required by the FK
// constraints of sys_users/sys_depts/sys_roles, and returns the tenant id.
func seedTestData(t *testing.T, client *ent.Client) int64 {
	t.Helper()
	ctx := mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 1), "system"),
		0,
	)
	pkg, err := client.SysPackage.Create().
		SetName("测试套餐").
		SetCode("test").
		SetSort(1).
		SetStatus(syspackage.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create test package: %v", err)
	}
	// The package_id FK column defaults to 0, which has no matching package
	// row, so an explicit reference is required when foreign keys are enabled.
	tenant, err := client.SysTenant.Create().
		SetName("测试租户").
		SetCode("test").
		SetAdminID(1).
		SetPackageID(pkg.ID).
		SetStatus(systenant.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create test tenant: %v", err)
	}
	return tenant.ID
}

// newTestClient creates an in-memory SQLite ent client for testing,
// seeded with a package + tenant so FK constraints are satisfied.
func newTestClient(t *testing.T) (*ent.Client, int64) {
	t.Helper()
	db, err := stdsql.Open("sqlite", "file:ent?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := client.Schema.Create(context.Background(), migrate.WithForeignKeys(true)); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	tenantID := seedTestData(t, client)
	return client, tenantID
}

func TestCreateUser_Success(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)

	svcCtx := &svc.ServiceContext{DB: client}
	logger := logx.WithContext(ctx)

	logic := &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logger,
	}

	in := &apps.UserReq{
		Username: proto.String("testuser"),
		Password: proto.String("password123"),
		Nickname: proto.String("测试用户"),
		Mobile:   proto.String("13800138000"),
		Email:    proto.String("test@example.com"),
		Status:   proto.String("active"),
		Remark:   proto.String("测试用"),
	}

	resp, err := logic.CreateUser(in)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if resp.GetCode() != 200 {
		t.Fatalf("expected code 200, got %d", resp.GetCode())
	}
	if resp.GetData().GetId() <= 0 {
		t.Fatalf("expected positive id, got %d", resp.GetData().GetId())
	}
	if resp.GetData().GetIdStr() == "" {
		t.Fatal("expected non-empty idStr")
	}
	if resp.GetData().GetStatus() != "active" {
		t.Fatalf("expected status active, got %s", resp.GetData().GetStatus())
	}

	// Verify user was actually saved
	user, err := client.SysUser.Get(ctx, resp.GetData().GetId())
	if err != nil {
		t.Fatalf("failed to fetch created user: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username testuser, got %s", user.Username)
	}
	if user.Nickname != "测试用户" {
		t.Fatalf("expected nickname 测试用户, got %s", user.Nickname)
	}
	if user.Mobile != "13800138000" {
		t.Fatalf("expected mobile 13800138000, got %s", user.Mobile)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", user.Email)
	}
	if user.Remark != "测试用" {
		t.Fatalf("expected remark 测试用, got %s", user.Remark)
	}
	if user.Status != sysuser.StatusActive {
		t.Fatalf("expected status active, got %s", user.Status)
	}

	// Verify audit fields were set by mixin hooks
	if user.TenantID != tenantID {
		t.Fatalf("expected tenantId %d, got %d", tenantID, user.TenantID)
	}
	if user.CreatedBy != "test-admin" {
		t.Fatalf("expected createdBy test-admin, got %s", user.CreatedBy)
	}
}

func TestCreateUser_WithDepartment(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)

	// First create a department
	dept, err := client.SysDept.Create().
		SetName("技术部").
		SetStatus("active").
		SetSort(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create dept: %v", err)
	}

	svcCtx := &svc.ServiceContext{DB: client}
	logic := &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}

	in := &apps.UserReq{
		Username: proto.String("deptuser"),
		Password: proto.String("pass123"),
		Nickname: proto.String("部门用户"),
		Status:   proto.String("active"),
		DeptId:   proto.Int64(dept.ID),
	}

	resp, err := logic.CreateUser(in)
	if err != nil {
		t.Fatalf("CreateUser with dept failed: %v", err)
	}

	user, err := client.SysUser.Get(ctx, resp.GetData().GetId())
	if err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}
	if user.DeptID != dept.ID {
		t.Fatalf("expected deptId %d, got %d", dept.ID, user.DeptID)
	}
}

func TestCreateUser_WithRoles(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)

	// First create roles
	role, err := client.SysRole.Create().
		SetName("管理员").
		SetCode("admin").
		SetSort(1).
		SetStatus(sysrole.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	svcCtx := &svc.ServiceContext{DB: client}
	logic := &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}

	in := &apps.UserReq{
		Username: proto.String("roleuser"),
		Password: proto.String("pass123"),
		Nickname: proto.String("角色用户"),
		Status:   proto.String("active"),
		RoleIds:  []int64{role.ID},
	}

	resp, err := logic.CreateUser(in)
	if err != nil {
		t.Fatalf("CreateUser with roles failed: %v", err)
	}

	user, err := client.SysUser.Query().
		Where(sysuser.IDEQ(resp.GetData().GetId())).
		WithRoles().
		Only(ctx)
	if err != nil {
		t.Fatalf("failed to fetch user with roles: %v", err)
	}
	if len(user.Edges.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(user.Edges.Roles))
	}
	if user.Edges.Roles[0].Code != "admin" {
		t.Fatalf("expected role code admin, got %s", user.Edges.Roles[0].Code)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)

	svcCtx := &svc.ServiceContext{DB: client}
	logic := &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}

	in := &apps.UserReq{
		Username: proto.String("duplicate"),
		Password: proto.String("pass123"),
		Nickname: proto.String("重复用户1"),
		Status:   proto.String("active"),
	}

	_, err := logic.CreateUser(in)
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	// Second create with same username should fail with a friendly business error
	_, err = logic.CreateUser(in)
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
	if e, ok := err.(*errno.Errno); !ok || e.Code != errno.UsernameExists.Code {
		t.Fatalf("expected username-exists business error (code %d), got %v", errno.UsernameExists.Code, err)
	}
}

func TestCreateUser_SameUsernameDifferentTenant(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	// Create user in tenant A
	ctxA := testCtx(tenantID)
	svcCtx := &svc.ServiceContext{DB: client}
	logic := &CreateUserLogic{
		ctx:    ctxA,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctxA),
	}
	in := &apps.UserReq{
		Username: proto.String("cross-tenant"),
		Password: proto.String("pass123"),
		Nickname: proto.String("租户A用户"),
		Status:   proto.String("active"),
	}
	if _, err := logic.CreateUser(in); err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	// Seed a second tenant so the tenant_id FK is satisfied
	sysCtx := mixins.SetCurrentTenantId(
		mixins.SetCurrentUserName(mixins.SetCurrentUserId(context.Background(), 1), "system"),
		0,
	)
	pkg, err := client.SysPackage.Query().First(sysCtx)
	if err != nil {
		t.Fatalf("failed to load test package: %v", err)
	}
	tenantB, err := client.SysTenant.Create().
		SetName("测试租户B").
		SetCode("test-b").
		SetAdminID(1).
		SetPackageID(pkg.ID).
		SetStatus(systenant.StatusActive).
		Save(sysCtx)
	if err != nil {
		t.Fatalf("failed to create test tenant B: %v", err)
	}

	// Same username in another tenant must be allowed (uniqueness is per-tenant)
	ctxB := testCtx(tenantB.ID)
	logicB := &CreateUserLogic{
		ctx:    ctxB,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctxB),
	}
	if _, err := logicB.CreateUser(in); err != nil {
		t.Fatalf("same username in another tenant should succeed: %v", err)
	}
}

func TestCreateUser_EmptyUsername(t *testing.T) {
	client, tenantID := newTestClient(t)
	defer client.Close()

	ctx := testCtx(tenantID)
	svcCtx := &svc.ServiceContext{DB: client}
	logic := &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}

	in := &apps.UserReq{
		Username: proto.String(""),
		Password: proto.String("pass123"),
		Status:   proto.String("active"),
	}

	_, err := logic.CreateUser(in)
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
}
