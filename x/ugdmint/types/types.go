package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintRecord defines the structure for minting records
type MintRecord struct {
	BlockHeight int64     `json:"block_height"`
	Account     string    `json:"account"`
	Amount      sdk.Coins `json:"amount"`
}
