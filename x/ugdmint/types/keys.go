package types

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey = []byte{0x00}
	ParamsKey = []byte{0x01}
)

const (
	// ModuleName defines the module name
	ModuleName = "ugdmint"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)
