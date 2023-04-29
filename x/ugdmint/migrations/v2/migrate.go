package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/exported"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/types"
)

const (
	ModuleName = "ugdmint"
)

var ParamsKey = []byte{0x01}

// Migrate migrates the x/mint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/ugdmint
// module state.
func Migrate(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	var currParams types.Params
	legacySubspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&currParams)
	store.Set(ParamsKey, bz)

	return nil
}
