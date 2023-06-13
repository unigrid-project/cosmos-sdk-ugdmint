package types

import (
	"encoding/json"
	"fmt"
	"http"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type HedgehogData struct {
	entry map[string]string
}

type Mint struct {
	address string
	amount 	string
	heigth	string
}

type mintCache struct {
	stop chan struct{}

	wg	sync.WaitGroup
	mu 	sync.RWMutex
	mints map[int64]Mint

	//mints *cache.Cache
}

const (
	defaultExperation = 1 * time.Minute
	purgeTime         = 2 * time.Minute
)

var c = newCache()

//parse hedgehog data
func parse() {

}

func (c *mintCache) cleanupCache() {
	t := time.NewTicker(purgeTime)
	defer t.Stop()
}

func newCache() *mintCache {
	Cache := cache.New(defaultExperation, purgeTime)
	return &mintCache{
		mints: Cache,
	}
}

func (c *mintCache) read(heigth int64) (item []byte, ok bool) {
	mint, ok := c.mints.Get(heigth)

	if ok {
		res, err := json.Marshal(mint.(Mint))
		if err != nil {

		}
		return res, true
	}

	return nil, false
}

func (c *mintCache) updateCache(heigth int64, mint Mint) {
	c.mints.Set(heigth, mint, cache.DefaultExpiration)
}

func (c *mintCache) delete(heigth int64) {

}

func checkCache(heigth int64) Mint {

	res, ok := c.read(heigth)
	if ok {
		return res
	}
	mint Mint
	return mint

}

func callHedgehog(endpoint string) {
	response, err := http.Get("https://127.0.0.1:52884/" + endpoint)
}

func getMints(c *gin.Context) {
	c.IndentJSON(http.StatusOK, Mint)
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
