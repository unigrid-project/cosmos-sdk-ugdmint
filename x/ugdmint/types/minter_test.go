package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	fmt "fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	//"github.com/cosmos/cosmos-sdk/x/bank/types"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/golang/mock/gomock"
)

const (
	jsonString string = "{\"timeStamp\":\"2023-06-16T19:03:33.104Z\"," +
		"\"previousTimeStamp\":\"2023-06-16T19:03:01.836Z\"," +
		"\"flags\":0,\"type\":\"MINT_STORAGE\"," +
		"\"data\":{" +
		"\"mints\":{" +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43hlvy4jatsg/80\":100," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43tlvy4jatsg/90\":1000," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43ulvy4jatsg/110\":1275," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43alvy4jatsg/150\":981256," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/165\":1236," +
		"\"pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/147621207\":1236," +
		"\"pk2sxhrywmrsqtnas3p7gu0t8x43rlvy4jatsg/1238965123\":1236" +
		"}}," +
		"\"previousData\":" +
		"{" +
		"\"mints\":{" +
		"\"yyy/1337\":1337,\"Shit/82\":5}" +
		"}," +
		"\"signature\":\"MIGIAkIBoUwt+6QGWerjHUrq/LFn0US3OL3pBgGoibLn6rONgZi8wM42XQR4zAFFycw8baXMilXvXvd8ik+RcXSyfBsiiSkCQgGQ2LDbzNfXObev1CqIfGm1OzXmoUblwoIWvIsEi+46ueYiKkUJL/0nz0AgeGaysZDvvbzrv/FhJiZxahIhyHrFKA==\"" +
		"}"

	/*"{" +
	"\"data\": {" +
	"\"mints\":{" +
	"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43hlvy4jatsg/80\":\"100\"," +
	"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43tlvy4jatsg/90\":\"1000\"," +
	"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43ulvy4jatsg/110\":\"1275\"," +
	"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43alvy4jatsg/150\":\"981256\"," +
	"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/165\":\"1236\"," +
	"\"pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/147621207\":\"1236\"," +
	"\"pk2sxhrywmrsqtnas3p7gu0t8x43rlvy4jatsg/1238965123\":\"1236\"" +
	"}" +
	"}" +
	"}"*/
)

//const ModuleName = "ugdmint"

var (
	mux    *http.ServeMux
	server *httptest.Server

	//_, _, addr   = testdata.KeyTestPubAddr()
	mintAcct = authtypes.NewModuleAddress(ModuleName)
	//TestProposal = getTestProposal()

)

func serverSetup() func() {
	mux = http.NewServeMux()

	// priv, err := rsa.GenerateKey(rand.Reader, *rsaBits)
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	/*
	   hosts := strings.Split(*host, ",")
	   for _, h := range hosts {
	   	if ip := net.ParseIP(h); ip != nil {
	   		template.IPAddresses = append(template.IPAddresses, ip)
	   	} else {
	   		template.DNSNames = append(template.DNSNames, h)
	   	}
	   }
	   if *isCA {
	   	template.IsCA = true
	   	template.KeyUsage |= x509.KeyUsageCertSign
	   }
	*/

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	out := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certKey := out.Bytes()
	fmt.Println(out.String())
	out.Reset()
	pem.Encode(out2, pemBlockForKey(priv))
	pubKey := out2.Bytes()
	fmt.Println(out2.String())

	cert, err := tls.X509KeyPair(certKey, pubKey)
	//tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Panic("bad server certs: ", err)
	}
	certs := []tls.Certificate{cert}

	server = httptest.NewUnstartedServer(mux)
	server.TLS = &tls.Config{Certificates: certs}
	//server.URL = "http://127.0.0.1:52884"
	//server. = "0.0.0.0:52884"
	server.StartTLS()

	return func() {
		server.Close()
	}
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func TestCanMintFromHedgehog(t *testing.T) {

	compareValue := []Mint{{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43hlvy4jatsg", 100, "80"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43tlvy4jatsg", 1000, "90"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43ulvy4jatsg", 1275, "110"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43alvy4jatsg", 981256, "150"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg", 1236, "165"},
	}

	//http.HandleFunc("/mint-storage", mintStorage)

	teardown := serverSetup()
	//defer teardown()

	mux.HandleFunc("/gridspork/mint-storage", func(w http.ResponseWriter, r *http.Request) {
		//r.RequestURI = "/gridspork/mint-storage"
		//r.Host = "https://127.0.0.1:52884"
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsonString))
	})

	//server.Config.Addr = "localhost:52884"
	//server.URL = "https://127.0.0.1:52884"
	//server.StartTLS()

	cache := NewCache()

	cache.callHedgehog(server.URL + "/gridspork/mint-storage")
	for _, cv := range compareValue {
		h, er := strconv.ParseInt(cv.Heigth, 10, 64)
		if er != nil {
			panic("it whent to hell")
		}
		if v, found := cache.mints[uint64(h)]; found {
			fmt.Println("Found mint in cache " + v.Address)
		} else {
			t.Error("compare value was not in mintcache")
		}
	}

	teardown()
}

func TestBlockProvision(t *testing.T) {

	key := sdk.NewKVStoreKey(ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(types.Header{Time: time.Now()})

	prevCtx := testCtx.Ctx.WithBlockHeader(types.Header{Time: time.Now().Add(-time.Duration(10) * time.Second)})

	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	//bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)

	acctKeeper.EXPECT().GetModuleAddress(ModuleName).Return(mintAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAccount(gomock.Any(), ModuleName).Return(authtypes.NewEmptyModuleAccount(ModuleName)).AnyTimes()
	//trackMockBalances(bankKeeper)
	stakingKeeper.EXPECT().TokensFromConsensusPower(ctx, gomock.Any()).DoAndReturn(func(ctx sdk.Context, power int64) sdkmath.Int {
		return sdk.TokensFromConsensusPower(power, sdkmath.NewIntFromUint64(1000000))
	}).AnyTimes()
	stakingKeeper.EXPECT().BondDenom(ctx).Return("ugd").AnyTimes()
	stakingKeeper.EXPECT().IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().IterateDelegations(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).Return(sdkmath.NewInt(10000000)).AnyTimes()

	params := Params{
		MintDenom:              "ugd",
		SubsidyHalvingInterval: sdk.NewDecWithPrec(50000, 0),
		GoalBonded:             sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:          uint64(60 * 60 * 8766 / 5),
	}

	minter := NewMinter(params.SubsidyHalvingInterval)

	coins := Minter.BlockProvision(minter, params, 10, ctx, prevCtx)

	fmt.Println(coins.AmountOf("ugd"))
	fmt.Println(coins.AmountOf("fermi"))

	fmt.Println(coins.String())
	if !coins.AmountOf("ugd").IsZero() {
		fmt.Println("passed")
	} else {
		t.Error("faild amount 0")
	}

}

func TestFirstBlockProvision(t *testing.T) {

	key := sdk.NewKVStoreKey(ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(types.Header{Time: time.Now()})

	prevCtx := testCtx.Ctx.WithBlockHeader(types.Header{Time: time.Now()})

	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	//bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)

	acctKeeper.EXPECT().GetModuleAddress(ModuleName).Return(mintAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAccount(gomock.Any(), ModuleName).Return(authtypes.NewEmptyModuleAccount(ModuleName)).AnyTimes()
	//trackMockBalances(bankKeeper)
	stakingKeeper.EXPECT().TokensFromConsensusPower(ctx, gomock.Any()).DoAndReturn(func(ctx sdk.Context, power int64) sdkmath.Int {
		return sdk.TokensFromConsensusPower(power, sdkmath.NewIntFromUint64(1000000))
	}).AnyTimes()
	stakingKeeper.EXPECT().BondDenom(ctx).Return("ugd").AnyTimes()
	stakingKeeper.EXPECT().IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().IterateDelegations(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).Return(sdkmath.NewInt(10000000)).AnyTimes()

	params := Params{
		MintDenom:              "ugd",
		SubsidyHalvingInterval: sdk.NewDecWithPrec(50000, 0),
		GoalBonded:             sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:          uint64(60 * 60 * 8766 / 5),
	}

	minter := NewMinter(params.SubsidyHalvingInterval)

	coins := Minter.BlockProvision(minter, params, 10, ctx, prevCtx)

	fmt.Println(coins.AmountOf("ugd"))
	fmt.Println(coins.AmountOf("fermi"))

	fmt.Println(coins.String())
	if !coins.AmountOf("fermi").IsZero() {
		fmt.Println("passed")
	} else {
		t.Error("faild amount 0")
	}

}

func TestAddressConvertion(t *testing.T) {

	compareValue := []Mint{{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43hlvy4jatsg", 100, "80"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43tlvy4jatsg", 1000, "90"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43ulvy4jatsg", 1275, "110"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43alvy4jatsg", 981256, "150"},
		{"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg", 1236, "165"},
	}

	key := sdk.NewKVStoreKey(ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(types.Header{ChainID: "unigrid", Time: time.Now()})
	fmt.Println(ctx.BlockHeader().ChainID)
	//fmt.Println(ctx.BlockHeader())

	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	//bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)
	acctKeeper.EXPECT().GetModuleAddress(ModuleName).Return(mintAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAccount(gomock.Any(), ModuleName).Return(authtypes.NewEmptyModuleAccount(ModuleName)).AnyTimes()
	//trackMockBalances(bankKeeper)
	stakingKeeper.EXPECT().TokensFromConsensusPower(ctx, gomock.Any()).DoAndReturn(func(ctx sdk.Context, power int64) sdkmath.Int {
		return sdk.TokensFromConsensusPower(power, sdkmath.NewIntFromUint64(1000000))
	}).AnyTimes()
	stakingKeeper.EXPECT().BondDenom(ctx).Return("ugd").AnyTimes()
	stakingKeeper.EXPECT().IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().IterateDelegations(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).Return(sdkmath.NewInt(10000000)).AnyTimes()

	add, err := ConvertStringToAcc(compareValue[0].Address)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(add.String())
	if compareValue[0].Address == add.String() {
		fmt.Println("test passed")
	} else {
		t.Error("address convertion failed")
	}
	//unigrid1x9cxkvnn0p58y7thd4u8xut5deshxvmsxanh2vr58purgvmgd3m8jdr2v968xect00gs9
	//cosmos1x9cxkvnn0p58y7thd4u8xut5deshxvmsxanh2vr58purgvmgd3m8jdr2v968xecqaqucu
}
