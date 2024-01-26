package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
)

// InitGenesis new ugdmint genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)

	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
	goCtx := sdk.UnwrapSDKContext(ctx)
	ak.GetModuleAccount(goCtx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	return types.NewGenesisState(minter, params)
}
