package keeper

import (
	"encoding/binary"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/types"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         storetypes.StoreKey
		stakingKeeper    types.StakingKeeper
		bankKeeper       types.BankKeeper
		feeCollectorName string
		hedgehogUrl      string
		authKeeper       types.AccountKeeper
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		stakingKeeper:    sk,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		authKeeper:       ak,
	}
}

const PreviousBlockTimeKey = "previousBlockTime"

// SetHedgehogUrl sets the module's hedgehog url.
func (k *Keeper) SetHedgehogUrl(url string) {
	k.hedgehogUrl = url
}

// GetHedgehogUrl returns the module's hedgehog url.
func (k *Keeper) GetHedgehogUrl() string {
	return k.hedgehogUrl
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.stakingKeeper.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.stakingKeeper.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

// Send coins to new mint
func (k Keeper) AddNewMint(ctx sdk.Context, coins sdk.Coins, reciver sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, reciver, coins)
}

func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return k.authKeeper.GetAccount(ctx, addr)
}

func (k Keeper) SetAccount(ctx sdk.Context, acc authtypes.AccountI) {
	k.authKeeper.SetAccount(ctx, acc)
}

func (k Keeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, addr)
}

func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) time.Time {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(PreviousBlockTimeKey))
	if b == nil {
		return time.Time{} // or some default value
	}
	return time.Unix(int64(binary.BigEndian.Uint64(b)), 0)
}

func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t.Unix()))
	store.Set([]byte(PreviousBlockTimeKey), b)
}
