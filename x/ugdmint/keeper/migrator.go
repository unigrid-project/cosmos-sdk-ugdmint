package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/exported"
	v2 "github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/migrations/v2"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace exported.Subspace
}

// NewMigrator returns Migrator instance for the state migration.
func NewMigrator(k Keeper, ss exported.Subspace) Migrator {
	return Migrator{
		keeper:         k,
		legacySubspace: ss,
	}
}

// Migrate1to2 migrates the x/ugdmint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/ugdmint
// module state.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}
