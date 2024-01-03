package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the ugdmint module.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

// SubsidyHalvingInterval returns params.SubsidyHalvingInterval of the ugdmint module.
func (k Keeper) SubsidyHalvingInterval(goCtx context.Context, req *types.QuerySubsidyHalvingIntervalRequest) (*types.QuerySubsidyHalvingIntervalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	return &types.QuerySubsidyHalvingIntervalResponse{SubsidyHalvingInterval: params.SubsidyHalvingInterval}, nil
}
