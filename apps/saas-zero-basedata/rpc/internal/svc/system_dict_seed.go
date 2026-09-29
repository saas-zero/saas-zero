package svc

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
)

type SystemDictSeed struct {
	Name   string
	Key    string
	Remark string
	Items  []SystemDictDataSeed
}

type SystemDictDataSeed struct {
	Name  string
	Key   string
	Value string
}

// SystemDictSeeds contains the system-default dictionaries used by the UI.
// They are stored with tenant_id=0 and are inherited by every tenant.
var SystemDictSeeds = []SystemDictSeed{
	{
		Name:   "状态",
		Key:    "status",
		Remark: "系统状态选项",
		Items: []SystemDictDataSeed{
			{Name: "启用", Key: "active", Value: "active"},
			{Name: "禁用", Key: "inactive", Value: "inactive"},
			{Name: "暂停", Key: "suspended", Value: "suspended"},
			{Name: "冻结", Key: "frozen", Value: "frozen"},
			{Name: "过期", Key: "expired", Value: "expired"},
			{Name: "成功", Key: "success", Value: "success"},
			{Name: "失败", Key: "fail", Value: "fail"},
			{Name: "失败", Key: "failure", Value: "failure"},
		},
	},
	{
		Name:   "菜单类型",
		Key:    "menu_type",
		Remark: "菜单节点类型",
		Items: []SystemDictDataSeed{
			{Name: "目录", Key: "directory", Value: "directory"},
			{Name: "菜单", Key: "menu", Value: "menu"},
			{Name: "按钮", Key: "button", Value: "button"},
		},
	},
	{
		Name:   "API类型",
		Key:    "api_type",
		Remark: "API资源类型",
		Items: []SystemDictDataSeed{
			{Name: "分组", Key: "group", Value: "group"},
			{Name: "接口", Key: "api", Value: "api"},
		},
	},
	{
		Name:   "任务分组",
		Key:    "job_group",
		Remark: "定时任务分组",
		Items: []SystemDictDataSeed{
			{Name: "默认分组", Key: "default", Value: "default"},
			{Name: "报表分组", Key: "report", Value: "report"},
			{Name: "同步分组", Key: "sync", Value: "sync"},
			{Name: "清理分组", Key: "cleanup", Value: "cleanup"},
		},
	},
	{
		Name:   "错过执行策略",
		Key:    "misfire_policy",
		Remark: "任务错过执行时的处理策略",
		Items: []SystemDictDataSeed{
			{Name: "放弃", Key: "skip", Value: "skip"},
			{Name: "补跑一次", Key: "fire_once", Value: "fire_once"},
			{Name: "全部补跑", Key: "fire_all", Value: "fire_all"},
		},
	},
	{
		Name:   "触发类型",
		Key:    "trigger_type",
		Remark: "任务触发来源",
		Items: []SystemDictDataSeed{
			{Name: "定时触发", Key: "cron", Value: "cron"},
			{Name: "手动触发", Key: "manual", Value: "manual"},
		},
	},
	{
		Name:   "任务执行状态",
		Key:    "job_exec_status",
		Remark: "任务执行结果状态",
		Items: []SystemDictDataSeed{
			{Name: "成功", Key: "success", Value: "success"},
			{Name: "失败", Key: "fail", Value: "fail"},
			{Name: "超时", Key: "timeout", Value: "timeout"},
			{Name: "跳过", Key: "skipped", Value: "skipped"},
			{Name: "无", Key: "none", Value: "none"},
		},
	},
}

// SeedSystemDicts inserts missing system dictionaries and items in one transaction.
func SeedSystemDicts(ctx context.Context, db *ent.Client) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}
	if err := SeedSystemDictsTx(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// SeedSystemDictsTx is used by InitAll so dictionary rows participate in its transaction.
func SeedSystemDictsTx(ctx context.Context, tx *ent.Tx) error {
	systemCtx := mixins.SetCurrentTenantId(ctx, 0)
	if mixins.GetCurrentUserId(systemCtx) <= 0 {
		systemCtx = mixins.SetCurrentUserId(systemCtx, 1)
	}
	if mixins.GetCurrentUserName(systemCtx) == "" {
		systemCtx = mixins.SetCurrentUserName(systemCtx, "system")
	}

	for _, seed := range SystemDictSeeds {
		dict, err := tx.SysDict.Query().
			Where(sysdict.KeyEQ(seed.Key), sysdict.TenantIDEQ(0), sysdict.DeletedAtIsNil()).
			Only(systemCtx)
		if ent.IsNotFound(err) {
			dict, err = tx.SysDict.Create().
				SetName(seed.Name).
				SetKey(seed.Key).
				SetStatus(sysdict.StatusActive).
				SetRemark(seed.Remark).
				Save(systemCtx)
		} else if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		for _, item := range seed.Items {
			_, err := tx.SysDictData.Query().
				Where(
					sysdictdata.DictIDEQ(dict.ID),
					sysdictdata.KeyEQ(item.Key),
					sysdictdata.TenantIDEQ(0),
					sysdictdata.DeletedAtIsNil(),
				).
				Only(systemCtx)
			if ent.IsNotFound(err) {
				_, err = tx.SysDictData.Create().
					SetDictID(dict.ID).
					SetName(item.Name).
					SetKey(item.Key).
					SetValue(item.Value).
					SetStatus(sysdictdata.StatusActive).
					Save(systemCtx)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
