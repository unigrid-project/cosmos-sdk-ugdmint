package ugdmint

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/keeper"
	"github.com/unigrid-project/cosmos-ugdmint/x/ugdmint/types"
)

var (
	prevBlockTime = time.Now()
	account       authtypes.BaseAccount
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
	bondedRatio, _ := k.BondedRatio(ctx)

	minter.SubsidyHalvingInterval = params.SubsidyHalvingInterval
	k.SetMinter(ctx, minter)

	// get the previous block time from the context
	prevCtx := sdk.NewContext(ctx.MultiStore(), ctx.BlockHeader(), false, log.NewNopLogger()).WithBlockHeight(ctx.BlockHeight() - 1)
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
	//fmt.Printf("height: %d\n", height)
	m, mErr := mc.Read(height)
	if mErr == nil {
		//fmt.Println("There were no errors when checking height. its time to mint to address!!")
		acc, aErr := types.ConvertStringToAcc(m.Address)

		if aErr != nil {
			fmt.Println("convert to account failed")
			panic("error!!!!")
		}
		// get the actual account from the account keeper
		account := k.GetAccount(ctx, acc)
		//fmt.Println("Acc:", acc)

		if account == nil {
			// Create a new BaseAccount with the address
			baseAcc := authtypes.NewBaseAccountWithAddress(acc)
			//fmt.Println("BaseAccount:", baseAcc)
			// Set the initial balance for the account (if you have any initial balance to set)
			// baseAcc.SetCoins(initialBalance)
			//fmt.Println("baseAcc.PubKey:", baseAcc.PubKey)
			// Convert the BaseAccount to a DelayedVestingAccount
			endTime := ctx.BlockTime().Add(10 * 365 * 24 * time.Hour) // 10 years from now
			vestingAcc, _ := vestingtypes.NewDelayedVestingAccount(baseAcc, sdk.Coins{}, endTime.Unix())
			//fmt.Println("Vesting Account:", vestingAcc)
			// Set this new account in the keeper
			k.SetAccount(ctx, vestingAcc)
		} else if baseAcc, ok := account.(*authtypes.BaseAccount); ok {
			endTime := ctx.BlockTime().Add(10 * 365 * 24 * time.Hour) // 10 years from now
			currentBalances := k.GetAllBalances(ctx, baseAcc.GetAddress())
			//fmt.Println("baseAcc.PubKey:", baseAcc.PubKey)
			vestingAcc, _ := vestingtypes.NewDelayedVestingAccount(baseAcc, currentBalances, endTime.Unix())
			k.SetAccount(ctx, vestingAcc)
		} else if baseAcc, ok := account.(*vestingtypes.DelayedVestingAccount); ok {
			currentBalances := k.GetAllBalances(ctx, baseAcc.GetAddress())

			startTime := ctx.BlockTime().Unix() // Current block time as start time

			// Calculate the amount for each vesting period for each coin in currentBalances
			amountPerPeriod := sdk.Coins{}
			for _, coin := range currentBalances {
				amount := coin.Amount.Quo(math.NewInt(10))
				amountPerPeriod = append(amountPerPeriod, sdk.NewCoin(coin.Denom, amount))
			}

			// Create 10 vesting periods, each 1 minute apart
			periods := vestingtypes.Periods{}
			for i := 0; i < 10; i++ {
				period := vestingtypes.Period{
					Length: 60, // 60 seconds = 1 minute
					Amount: amountPerPeriod,
				}
				periods = append(periods, period)
			}
			fmt.Println("baseAcc.PubKey:", baseAcc.PubKey)
			baseAccount := &authtypes.BaseAccount{
				Address:       baseAcc.Address,
				PubKey:        baseAcc.PubKey,
				AccountNumber: baseAcc.AccountNumber,
				Sequence:      baseAcc.Sequence,
			}

			// Create the PeriodicVestingAccount
			vestingAcc, _ := vestingtypes.NewPeriodicVestingAccount(baseAccount, currentBalances, startTime, periods)
			k.SetAccount(ctx, vestingAcc)
		} //else if baseAcc, ok := account.(*vestingtypes.PeriodicVestingAccount); ok {
		//}

		coins := types.ConvertIntToCoin(params, m.Amount)
		//fmt.Println("time to mint")
		k.MintCoins(ctx, coins)
		//fmt.Printf("Coins are minted to address = %s\n", acc.String())
		mErr := k.AddNewMint(ctx, coins, acc)
		if mErr != nil {
			fmt.Println(mErr.Error())
		}
		//fmt.Println("Coins have been minted")
	}
}
