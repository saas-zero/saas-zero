package syspackageslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreatePackageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePackageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePackageLogic {
	return &CreatePackageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePackageLogic) CreatePackage(in *apps.PackageReq) (*apps.PackageResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	create := l.svcCtx.DB.SysPackage.Create().
		SetName(in.GetName()).
		SetCode(in.GetCode()).
		SetStatus(syspackage.Status(in.GetStatus())).
		SetSort(uint32(in.GetSort()))

	if in.GetRemark() != "" {
		create.SetRemark(in.GetRemark())
	}

	//logx.Infof("CreatePackage req:\n%s", prototext.Format(in))

	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.PackageResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.Package{
			Id:    proto.Int64(result.ID),
			IdStr: proto.String(id.ToString(result.ID)),
		},
	}, nil
}
