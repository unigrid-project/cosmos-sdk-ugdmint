package types

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	Heigth  string
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

	//mints *cache.Cache
}

type ErrorWhenGettingCache struct{}

const (
	//defaultExperation   = 1 * time.Minute
	cacheUpdateInterval = 15 * time.Second
)

var (
	c          = NewCache()
	currHeigth = uint64(1)
)

func (e *ErrorWhenGettingCache) Error() string {
	return "Faild to get address from cashe, cashe is probebly empty"
}

func (mc *MintCache) cleanupCache() {
	t := time.NewTicker(cacheUpdateInterval)
	defer t.Stop()
	for {
		select {
		case <-mc.stop:
			return
		case <-t.C:
			mc.mu.Lock()
			//update cache with new etries if any are found
			hedgehogUrl := viper.GetString("hedgehog.hedgehog_url")
			fmt.Println("hedgehogUrl in ugdmint:", hedgehogUrl)
			mc.callHedgehog(hedgehogUrl + "/gridspork/mint-storage")
			for h := range mc.mints {
				if h < currHeigth { //current heigth.
					mc.deleteFromCache(h)
				}
			}
			mc.mu.Unlock()
		}
	}
}

func GetCache() *MintCache {
	fmt.Println("Getting cache")
	fmt.Println(c)
	if c.mu == (sync.RWMutex{}) {
		return NewCache()
	}
	return c
}

func NewCache() *MintCache {
	mc := &MintCache{
		mints: make(map[uint64]Mint),
		stop:  make(chan struct{}),
	}

	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		mc.cleanupCache()
	}()

	return mc
}

func (mc *MintCache) Read(heigth uint64) (Mint, error) {

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	cm, ok := mc.mints[heigth]
	if !ok {
		return Mint{}, &ErrorWhenGettingCache{}
	}
	return cm, nil
}

func (mc *MintCache) updateCache(heigth uint64, mint Mint) {
	mc.mints[heigth] = mint
}

func (mc *MintCache) deleteFromCache(heigth uint64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.mints, heigth)
}

func (mc *MintCache) checkCache(heigth uint64) (mint Mint) {

	res, err := mc.Read(heigth)
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
	client := &http.Client{Transport: tr}
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

	blockHeigth := sdk.Context.BlockHeight(sdk.Context{})

	for key, amount := range res.Data.Mints {
		arr := strings.Split(key, "/")
		a := arr[0]
		heigth := arr[1]
		h, er := strconv.ParseInt(heigth, 10, 64)

		if er != nil {
			fmt.Println("Error parsing height:", er.Error())
			continue // Skip this iteration and move to the next one
		}

		if h >= blockHeigth && strings.Contains(a, "unigrid") {
			uh := uint64(h)
			mc.mints[uh] = Mint{
				Address: a,
				Heigth:  heigth,
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

	var nSubsidy float64 = 1

	currHeigth = height
	height = height + 2500000
	fmt.Println(params.SubsidyHalvingInterval.Abs().TruncateInt64())
	nBehalf := int64(height-1000000) / params.SubsidyHalvingInterval.Abs().TruncateInt64()
	fmt.Printf("nBehalf: %d \n", nBehalf)
	for i := 0; i < int(nBehalf); i++ {
		nSubsidy = nSubsidy * 99.0 / 100.0
	}

	fmt.Printf("nsubsidy: %f \n", nSubsidy)
	fmt.Println(ctx.BlockTime().Unix())
	fmt.Println(prevCtx.BlockTime().Unix())
	if ctx.BlockTime().Unix() <= prevCtx.BlockTime().Unix() {
		nSubsidy = nSubsidy * (float64(ctx.BlockTime().Unix()-(ctx.BlockTime().Unix()-60)) / 60.0)
	} else {
		nSubsidy = nSubsidy * (float64(ctx.BlockTime().Unix()-prevCtx.BlockTime().Unix()) / 60.0)
	}

	if nSubsidy < 0 {
		nSubsidy = 0
	}

	coin := sdk.NewCoin(params.MintDenom, sdk.NewIntFromUint64(uint64(nSubsidy*math.Pow10(8))))

	/*s := fmt.Sprintf("%f", nSubsidy)
	fmt.Printf("subsidy: %s \n", s)
	//unigrid1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8g6xaxu
	//Convertion from decimal to ugd and fermi. ugd is 10^8 and fermi is exponent 0.
	lDec, _ := sdk.NewDecFromStr(s)
	fmt.Println(lDec)
	deccoin := sdk.NewDecCoinFromDec("ugd", lDec)
	ugd, dcoin := deccoin.TruncateDecimal()
	fmt.Println(dcoin.Amount)
	d := dcoin.Amount.MulInt64(int64(math.Pow10(8)))
	dString := fmt.Sprintf("%dfermi", d)
	fmt.Println("d")
	fmt.Println(d)
	fermi, _ := sdk.ParseCoinNormalized(dString)

	fmt.Println(ugd.Amount)
	fmt.Println(fermi.Amount)

	coins := sdk.NewCoins(ugd, fermi) */

	return sdk.NewCoins(coin)
}
