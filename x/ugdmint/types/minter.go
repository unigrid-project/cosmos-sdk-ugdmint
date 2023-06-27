package types

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HedgehogData struct {
	Key    string
	amount string
}

type Mint struct {
	address string
	amount  int
	heigth  string
}

type mints struct {
	mints map[string]int
}

type hedgehogdata struct {
	timestamp         string `json:"timestamp"`
	previousTimeStamp string `json:"previousTimeStamp"`
	flags             int    `json:"flags"`
	hedgehogtype      string `json:"type"`
	data              mints  `json:"data"`
	previousData      mints  `json:"previousData"`
	signature         string `json:"signature"`
}

type mintCache struct {
	stop chan struct{}

	wg    sync.WaitGroup
	mu    sync.RWMutex
	mints map[int64]Mint

	//mints *cache.Cache
}

const (
	defaultExperation   = 1 * time.Minute
	cacheUpdateInterval = 2 * time.Minute
)

func (mc *mintCache) cleanupCache() {
	t := time.NewTicker(cacheUpdateInterval)
	defer t.Stop()

	blockHeigth := sdk.Context.BlockHeight(sdk.Context{})

	for {
		select {
		case <-mc.stop:
			return
		case <-t.C:
			mc.mu.Lock()
			//update cache with new etries if any are found
			mc.callHedgehog("mint-storage")
			for h := range mc.mints {
				if h < blockHeigth { //current heigth.
					mc.deleteFromCache(h)
				}
			}
			mc.mu.Unlock()
		}
	}
}

func NewCache() *mintCache {
	mc := &mintCache{
		mints: make(map[int64]Mint),
		stop:  make(chan struct{}),
	}

	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		mc.cleanupCache()
	}()

	return mc
}

func (mc *mintCache) read(heigth int64) (Mint, error) {

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	cm, ok := mc.mints[heigth]
	if !ok {
		return Mint{}, ErrIntOverflowGenesis
	}
	return cm, nil
}

func (mc *mintCache) updateCache(heigth int64, mint Mint) {
	mc.mints[heigth] = mint
}

func (mc *mintCache) deleteFromCache(heigth int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.mints, heigth)
}

func (mc *mintCache) checkCache(heigth int64) (mint Mint) {

	res, err := mc.read(heigth)
	if err != nil {
		return res
	}

	return mint

}

func (mc *mintCache) callHedgehog(serverUrl string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Get(serverUrl)

	if err != nil {
		panic("where is hedgehog " + err.Error())
	}
	defer response.Body.Close()
	var res hedgehogdata
	e := json.NewDecoder(response.Body).Decode(&res)

	if e != nil {
		fmt.Println(e.Error())
		//report response error in log
		return
	}

	blockHeigth := sdk.Context.BlockHeight(sdk.Context{})

	for key, amount := range res.data.mints {
		arr := strings.Split(key, "/")
		a := arr[0]
		heigth := arr[1]
		h, er := strconv.ParseInt(heigth, 10, 64)

		if er != nil {
			panic("error")
		}

		if h >= blockHeigth && strings.Contains(a, "unigrid1") {
			mc.mints[h] = Mint{
				address: a,
				heigth:  heigth,
				amount:  amount,
			}
		}
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
func (m Minter) BlockProvision(params Params, height uint64) sdk.Coin {

	nSubsidy := 1

	/*if (height == 1) {
		nSubsidy = 1200000
	} else if (height >= 5000 && height < 1050000) {
		nSubsidy = 8
	} else if (height >= 1050000 && height < 2100000) {
		nSubsidy = 6
	} else if (height >= 2100000 && height < 3150000) {
		nSubsidy = 4
	} else if (height >= 3150000 && height < 4200000) {
		nSubsidy = 2
	} else if (height >= 4200000 && height < 12600000) {
		nSubsidy = 1
	}*/

	if height > 1000000 {
		nBehalf := sdk.NewDec(int64(height - 100000)).Quo(params.SubsidyHalvingInterval).TruncateInt().Int64()

		for i := 0; i < int(nBehalf); i++ {
			nSubsidy = nSubsidy * 99 / 100
		}
	}

	provisionAmt := sdk.NewInt(int64(nSubsidy))
	// provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt)
}
