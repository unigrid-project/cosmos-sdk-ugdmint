/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom           		= []byte("MintDenom")
	KeySubsidyHalvingInterval 	= []byte("SubsidyHalvingInterval")
	KeyGoalBonded          		= []byte("GoalBonded")
	KeyBlocksPerYear       		= []byte("BlocksPerYear")
)

// Deprecated: ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs get the params.ParamSet
//
// Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeySubsidyHalvingInterval, &p.SubsidyHalvingInterval, validateSubsidyHalvingInterval),
		paramtypes.NewParamSetPair(KeyGoalBonded, &p.GoalBonded, validateGoalBonded),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
}

