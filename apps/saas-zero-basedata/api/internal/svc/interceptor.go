package svc

import (
	"context"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// authClientInterceptor reads current user/tenant info from the Go context
// and attaches them as gRPC outgoing metadata headers.  The RPC server's
// authInterceptor reads these headers and sets them back as Go context values.
//
// Only attaches metadata when values are non-zero, so that init endpoints
// (which set metadata manually in their logic) are not overridden.
func authClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	uid := mixins.GetCurrentUserId(ctx)
	uname := mixins.GetCurrentUserName(ctx)
	tid := mixins.GetCurrentTenantId(ctx)

	if uid > 0 || uname != "" || tid > 0 {
		md := metadata.Pairs(
			"x-user-id", strconv.FormatInt(uid, 10),
			"x-user-name", uname,
			"x-tenant-id", strconv.FormatInt(tid, 10),
		)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
