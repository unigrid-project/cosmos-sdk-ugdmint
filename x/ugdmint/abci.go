package ugdmint

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/keeper"
	"github.com/unigrid-project/cosmos-sdk-ugdmint/x/ugdmint/types"
)

var (
	prevBlockTime = time.Now()
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)
	height := uint64(ctx.BlockHeight())
	bondedRatio := k.BondedRatio(ctx)

	minter.SubsidyHalvingInterval = params.SubsidyHalvingInterval
	k.SetMinter(ctx, minter)
	//var prevCtx sdk.Context
	//if ctx.BlockHeader().Height != 1 {
	//prevCtx.
	prevCtx := sdk.NewContext(ctx.MultiStore(), ctx.BlockHeader(), false, log.NewNopLogger()).WithBlockTime(prevBlockTime)
	prevBlockTime = ctx.BlockTime()
	//} else {
	//prevCtx = ctx
	//}

	// mint coins, uodate supply
	mintedCoins := minter.BlockProvision(params, height, ctx, prevCtx)
	ok, mintedCoin := mintedCoins.Find("ugd")

	if !ok {
		_, mintedCoin = mintedCoins.Find("fermi")
	}
	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUGDMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeySubsidyHalvingInterval, minter.SubsidyHalvingInterval.String()),
			//sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoins.String()),
		),
	)

	//Start the mint cache and minting of new tokens when thier are any in hedgehog.
	mc := types.GetCache()
	fmt.Printf("Heigth: %d\n", height)
	m, mErr := mc.Read(height)

	if mErr == nil {
		fmt.Println("thier where no errors when checking heigth. its time to mint to address!!")
		acc, aErr := types.ConvertStringToAcc(m.Address)
		if aErr != nil {
			fmt.Println("convert to account failed")
			panic("error!!!!")
		}
		coins := types.ConvertIntToCoin(params, m.Amount)
		fmt.Println("time to mint")
		k.MintCoins(ctx, coins)
		mErr := k.AddNewMint(ctx, coins, acc)
		fmt.Println(mErr.Error())
	}
}
