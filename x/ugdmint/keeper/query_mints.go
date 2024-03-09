package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) AllMintRecords(goCtx context.Context, req *types.QueryAllMintRecordsRequest) (*types.QueryAllMintRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), []byte("MintRecordPrefix"))
	iterator := store.Iterator(nil, nil) // Iterate over all records
	defer iterator.Close()

	var records []types.MintRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.MintRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		records = append(records, record)
	}
	return &types.QueryAllMintRecordsResponse{MintRecords: records}, nil
}
