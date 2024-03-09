package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
)

type (
	Keeper struct {
		cdc codec.BinaryCodec
		// storeKey         storetypes.StoreKey
		storeService     store.KVStoreService
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
	//key storetypes.StoreKey,
	storeService store.KVStoreService,
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
		cdc: cdc,
		//storeKey:         key,
		storeService:     storeService,
		stakingKeeper:    sk,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		authKeeper:       ak,
	}
}

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
func (k Keeper) GetMinter(ctx context.Context) (minter types.Minter) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx context.Context, minter types.Minter) {

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
// func (k Keeper) StakingTokenSupply(ctx sdk.Context) math.Int {
// 	return k.stakingKeeper.StakingTokenSupply(ctx)
// }

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx context.Context) (math.LegacyDec, error) {
	return k.stakingKeeper.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
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

func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return k.authKeeper.GetAccount(ctx, addr)
}

func (k Keeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) error {
	k.authKeeper.SetAccount(ctx, acc)
	return nil
}

func (k Keeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, addr)
}

func (k Keeper) GetNextAccountNumber(ctx context.Context) (uint64, error) {
	// Use the authKeeper to get the next account number.
	// This assumes that the authKeeper has a NextAccountNumber method.
	return k.authKeeper.NextAccountNumber(ctx), nil
}

func (k Keeper) SetMintRecord(ctx sdk.Context, record types.MintRecord) error {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), []byte("MintRecordPrefix"))
	key := []byte(fmt.Sprintf("mintRecord:%d", record.BlockHeight))
	value := k.cdc.MustMarshal(&record)
	store.Set(key, value)
	return nil
}

func (k Keeper) GetMintRecord(ctx sdk.Context, blockHeight int64) (types.MintRecord, bool) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), []byte("MintRecordPrefix"))
	key := []byte(fmt.Sprintf("mintRecord:%d", blockHeight))
	value := store.Get(key)
	if value == nil {
		return types.MintRecord{}, false
	}

	var record types.MintRecord
	k.cdc.MustUnmarshal(value, &record)
	return record, true
}
