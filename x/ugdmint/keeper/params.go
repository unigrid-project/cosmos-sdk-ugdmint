package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
)

// // GetParams get all parameters as types.Params
// func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
// 	store := ctx.KVStore(k.storeKey)
// 	bz := store.Get(types.ParamsKey)
// 	if bz == nil {
// 		return p
// 	}

// 	k.cdc.MustUnmarshal(bz, &p)
// 	return p
// }

// // SetParams set the params
// func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
// 	if err := p.Validate(); err != nil {
// 		return err
// 	}

// 	store := ctx.KVStore(k.storeKey)
// 	bz := k.cdc.MustMarshal(&p)
// 	store.Set(types.ParamsKey, bz)

// 	return nil
// }

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}
