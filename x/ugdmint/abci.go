package ugdmint

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type StatusResponse struct {
	Result struct {
		SyncInfo struct {
			CatchingUp bool `json:"catching_up"`
		} `json:"sync_info"`
	} `json:"result"`
}

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

	prevCtx := sdk.NewContext(ctx.MultiStore(), ctx.BlockHeader(), false, log.NewNopLogger()).WithBlockTime(prevBlockTime)
	prevBlockTime = ctx.BlockTime()

	// mint coins, update supply
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
	//if isNodeSyncing() {
	//	fmt.Println("Node is syncing. Skipping the minting process.")
	//} else {
	if mErr == nil {
		fmt.Println("There were no errors when checking height. its time to mint to address!!")
		acc, aErr := types.ConvertStringToAcc(m.Address)

		if aErr != nil {
			fmt.Println("convert to account failed")
			panic("error!!!!")
		}
		coins := types.ConvertIntToCoin(params, m.Amount)
		fmt.Println("time to mint")
		k.MintCoins(ctx, coins)
		fmt.Printf("Coins are minted to address = %s\n", acc.String())
		mErr := k.AddNewMint(ctx, coins, acc)
		if mErr != nil {
			fmt.Println(mErr.Error())
		}
		fmt.Println("Coins have been minted")
	}
	//}
}

func isNodeSyncing() bool {
	resp, err := http.Get("http://localhost:26657/status")
	if err != nil {
		// Handle error or return true to be safe
		return true
	}
	defer resp.Body.Close()

	var statusResponse StatusResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&statusResponse)
	if err != nil {
		// Handle error or return true to be safe
		return true
	}

	return statusResponse.Result.SyncInfo.CatchingUp
}
