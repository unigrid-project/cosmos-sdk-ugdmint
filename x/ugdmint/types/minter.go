package types

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	math "math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
)

type Mint struct {
	Address string
	Amount  int
	height  string
}

type Mints struct {
	Mints map[string]int
}

type HedgehogData struct {
	Timestamp         string `json:"timestamp"`
	PreviousTimeStamp string `json:"previousTimeStamp"`
	Flags             int    `json:"flags"`
	Hedgehogtype      string `json:"type"`
	Data              Mints  `json:"data"`
	PreviousData      Mints  `json:"previousData"`
	Signature         string `json:"signature"`
}

type MintCache struct {
	stop chan struct{}

	wg    sync.WaitGroup
	mu    sync.RWMutex
	mints map[uint64]Mint
	first bool
	//mints *cache.Cache
}

type ErrorWhenGettingCache struct{}

const PreviousBlockTimeKey = "previousBlockTime"

const (
	//defaultExperation   = 1 * time.Minute
	cacheUpdateInterval = 15 * time.Second
)

var (
	c          = &MintCache{}
	once       sync.Once
	currheight = uint64(1)
)

func (e *ErrorWhenGettingCache) Error() string {
	return "Faild to get address from cashe, cashe is probebly empty"
}

func (mc *MintCache) cleanupCache() {
	t := time.NewTicker(cacheUpdateInterval)
	defer t.Stop()
	if mc.first { // Use mc.first instead of global first
		hedgehogUrl := viper.GetString("hedgehog.hedgehog_url")
		//fmt.Println("hedgehogUrl in ugdmint 1:", hedgehogUrl)
		mc.callHedgehog(hedgehogUrl + "/gridspork/mint-storage")
		mc.first = false
	}
	for {
		select {
		case <-mc.stop:
			return
		case <-t.C:
			mc.mu.Lock()
			hedgehogUrl := viper.GetString("hedgehog.hedgehog_url")
			//fmt.Println("hedgehogUrl in ugdmint 2:", hedgehogUrl)
			mc.callHedgehog(hedgehogUrl + "/gridspork/mint-storage")
			mc.mu.Unlock()
		}
	}
}

func GetCache() *MintCache {
	fmt.Println("Getting cache")
	once.Do(func() {
		c = NewCache()
	})
	return c
}

func NewCache() *MintCache {
	mc := &MintCache{
		mints: make(map[uint64]Mint),
		stop:  make(chan struct{}),
		first: true, // Initialize it here
	}

	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		mc.cleanupCache()
	}()

	return mc
}

func (mc *MintCache) Read(height uint64) (Mint, error) {

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	cm, ok := mc.mints[height]
	if !ok {
		return Mint{}, &ErrorWhenGettingCache{}
	}
	return cm, nil
}

func (mc *MintCache) updateCache(height uint64, mint Mint) {
	mc.mints[height] = mint
}

func (mc *MintCache) deleteFromCache(height uint64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.mints, height)
}

func (mc *MintCache) checkCache(height uint64) (mint Mint) {

	res, err := mc.Read(height)
	if err != nil {
		return res
	}

	return mint

}

func ConvertIntToCoin(params Params, amount int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(params.MintDenom, sdk.NewInt(int64(amount))))
}

func ConvertStringToAcc(address string) (sdk.AccAddress, error) {
	//sdk.GetConfig().SetBech32PrefixForAccount("unigrid", "unigrid")
	//s := strings.TrimPrefix(address, "unigrid")
	return sdk.AccAddressFromBech32(address)
}

func (mc *MintCache) callHedgehog(serverUrl string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second, // set to 10 seconds or as appropriate
	}

	response, err := client.Get(serverUrl)

	if err != nil {
		if err == io.EOF {
			fmt.Println("Received empty response from hedgehog server.")
		} else {
			fmt.Println("Error accessing hedgehog:", err.Error())
		}
		return
	}

	defer response.Body.Close()

	// Check if the response is empty
	if response.ContentLength == 0 {
		fmt.Println("Received empty response from hedgehog server.")
		return
	}

	var res HedgehogData
	body, err1 := io.ReadAll(response.Body)

	if err1 != nil {
		fmt.Println(err1.Error())
		//report response error in log
		return
	}

	e := json.Unmarshal(body, &res)
	//e := json.NewDecoder(response.Body).Decode(res)

	if e != nil {
		fmt.Println(e.Error())
		//report response error in log
		return
	}

	blockHeight := sdk.Context.BlockHeight(sdk.Context{})

	for key, amount := range res.Data.Mints {
		arr := strings.Split(key, "/")
		a := arr[0]
		height := arr[1]
		h, er := strconv.ParseInt(height, 10, 64)

		if er != nil {
			fmt.Println("Error parsing height:", er.Error())
			continue // Skip this iteration and move to the next one
		}

		if h >= blockHeight && strings.Contains(a, "unigrid") {
			uh := uint64(h)
			mc.mints[uh] = Mint{
				Address: a,
				height:  height,
				Amount:  amount,
			}
		}
	}
	for _, m := range mc.mints {
		fmt.Println(m)
	}
}

// NewMinter returns a new Minter object with the given subsidy halving interval.
func NewMinter(subsidyHalvingInterval sdk.Dec) Minter {
	return Minter{
		SubsidyHalvingInterval: subsidyHalvingInterval,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(subsidyHalvingInterval sdk.Dec) Minter {
	return NewMinter(
		subsidyHalvingInterval,
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses a subsidy halving interval of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(13, 2),
	)
}

// validate minter
func ValidateMinter(minter Minter) error {
	if minter.SubsidyHalvingInterval.IsNegative() {
		return fmt.Errorf("mint parameter subsidy halving interval should be positive, is %s",
			minter.SubsidyHalvingInterval.String())
	}
	return nil
}

// BlockProvision returns the provisions for a block based on the UGD algorithm
// provisions rate.
func (m Minter) BlockProvision(params Params, height uint64, ctx sdk.Context, prevCtx sdk.Context) sdk.Coins {
	fmt.Printf("[BlockProvision] Called: height=%d\n", height)
	fmt.Printf("[BlockProvision] Current Block Time: %d, Previous Block Time: %d\n", ctx.BlockTime().Unix(), prevCtx.BlockTime().Unix())
	// Calculate the number of blocks per minute dynamically
	//blocksPerMinute := calculateBlocksPerMinute(ctx, prevCtx)
	blocksPerMinute := 12
	fmt.Printf("[BlockProvision] blocksPerMinute=%d\n", blocksPerMinute)

	var nSubsidy float64 = 1
	fmt.Printf("[BlockProvision] Initial nSubsidy: %f\n", nSubsidy)

	adjustedHeight := height + 2685066
	fmt.Printf("[BlockProvision] Adjusted Height: %d\n", adjustedHeight)

	nBehalf := int64(adjustedHeight-1000000) / params.SubsidyHalvingInterval.Abs().TruncateInt64()
	fmt.Printf("[BlockProvision] nBehalf: %d\n", nBehalf)

	for i := 0; i < int(nBehalf); i++ {
		nSubsidy = nSubsidy * 99.0 / 100.0
		//fmt.Printf("[BlockProvision] nSubsidy after %dth halving: %f\n", i+1, nSubsidy)
	}

	if ctx.BlockTime().Unix() <= prevCtx.BlockTime().Unix() {
		nSubsidy = nSubsidy * (float64(ctx.BlockTime().Unix()-(ctx.BlockTime().Unix()-60)) / 60.0)
		fmt.Printf("[BlockProvision] nSubsidy adjusted for block time <= prev block time: %f\n", nSubsidy)
	} else {
		nSubsidy = nSubsidy * (float64(ctx.BlockTime().Unix()-prevCtx.BlockTime().Unix()) / 60.0)
		fmt.Printf("[BlockProvision] nSubsidy adjusted for block time > prev block time: %f\n", nSubsidy)
	}

	// Adjust nSubsidy based on the actual blocks per minute
	// This is the key adjustment to ensure the subsidy aligns with your intended reward rate
	nSubsidy = nSubsidy / float64(blocksPerMinute)

	if nSubsidy < 0 {
		nSubsidy = 0
	}

	// Convert to coin with the adjusted subsidy
	subsidyInSmallestUnit := int64(nSubsidy * math.Pow10(8))
	coin := sdk.NewCoin(params.MintDenom, sdk.NewInt(subsidyInSmallestUnit))
	fmt.Printf("[BlockProvision] Coin generated: %s\n", coin.String())

	return sdk.NewCoins(coin)
}

// this function is not working as intended
// TODO find a 100% consistent way to calculate the number of blocks per minute
// func calculateBlocksPerMinute(ctx sdk.Context, prevCtx sdk.Context) int {
// 	currentTime := ctx.BlockTime().Unix()
// 	previousTime := prevCtx.BlockTime().Unix()

// 	// Calculate the time difference in seconds between the blocks
// 	timeDiff := currentTime - previousTime
// 	fmt.Printf("[calculateBlocksPerMinute] Time difference between blocks: %d seconds\n", timeDiff)

// 	// If timeDiff is 0 or negative, which shouldn't normally happen, use 5 seconds as a default
// 	if timeDiff <= 0 {
// 		timeDiff = 5 // Default to 5 seconds if the time difference is too small
// 	}

// 	// Calculate the number of blocks that would be produced in a minute
// 	blocksPerMinute := 60 / timeDiff

// 	return int(blocksPerMinute)
// }
