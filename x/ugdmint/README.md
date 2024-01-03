---
sidebar_position: 1
---

# `x/ugdmint`

## Contents

* [State](#state)
    * [Minter](#minter)
    * [Params](#params)
* [Begin-Block](#begin-block)
    * [BlockProvision](#blockprovision)
* [Parameters](#parameters)
* [Events](#events)
    * [BeginBlocker](#beginblocker)
* [Client](#client)
    * [CLI](#cli)
    * [gRPC](#grpc)
    * [REST](#rest)

## Concepts

### The Minting Mechanism

The Unigrid minting mechanism was designed to:

* gradually decrease minting amount as a chain ages

As a chain ages, meaning the block height increases, the minting reward will decrease until eventually minting rewards are exhausted.

Currently minting rewards will step down at hard coded height changes.  The first block in the chain receives the largest reward, then it steps down as shown in the following chart:

|      height       |  amount   |
|-------------------|-----------|
| 1                 | 1200000   |
| 2 - 4999          | 1         |
| 5000 - 1049999    | 8         |
| 1050000 - 2099999 | 6         |
| 2100000 - 3149999 | 4         |
| 3150000 - 4199999 | 2         |
| > 4199999         | 1         |

Subsidy-halving-interval is the one parameter that may adjust how long minting rewards continue, and at what rate the amount will decrease as the block height increases.  This decrease effect is applied to the rewards amount determined from the chart above.  Currently this decrease effect kicks in after a height of 1000000 blocks.  The interval is how often the minting amount will decrease after the threshold of 1000000 has been reached.  Each decrease interval will reduce the minting amount by an additional 1% until 100% is taken away.

## State

### Minter

The minter is a space for holding current subsidy-halving-interval.

* Minter: `0x00 -> ProtocolBuffer(minter)`

```protobuf reference
// Minter represents the minting state.
message Minter {
  // current subsidy halving interval
  string subsidy_halving_interval = 1 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable)   = false
  ];
}
```

### Params

The ugdmint module stores it's params in state with the prefix of `0x01`,
it can be updated with governance or the address with authority.

* Params: `ugdmint/params -> legacy_amino(params)`

```protobuf reference
// Params defines the parameters for the module.
message Params {
  option (gogoproto.goproto_stringer) = false;
  option (amino.name)                 = "cosmos-sdk/x/ugdmint/Params";

  // type of coin to mint
  string mint_denom = 1;
  // subsidy halving interval
  string subsidy_halving_interval = 2 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable)   = false
  ];
  // goal of percent bonded atoms
  string goal_bonded = 3 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable)   = false
  ];
  // expected blocks per year
  uint64 blocks_per_year = 4;
}
```

## Begin-Block

Minting parameters are recalculated and paid at the beginning of each block.

### BlockProvision

Calculate the provisions generated for each block based on current block height. The provisions are then minted by the `ugdmint` module's `ModuleMinterAccount` and then transferred to the `auth`'s `FeeCollector` `ModuleAccount`.

```go
BlockProvision(params Params, height uint64) sdk.Coin {

  nSubsidy := 1

  if (height == 1) {
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
  }

  if (height > 1000000) {
    nBehalf := sdk.NewDec(int64(height - 100000)).Quo(params.SubsidyHalvingInterval).TruncateInt().Int64()
    
    for i := 0; i < int(nBehalf); i++ {
      nSubsidy = nSubsidy * 99 / 100
    }
  }

  provisionAmt := sdk.NewInt(int64(nSubsidy))
  // provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerYear)))
  return sdk.NewCoin(params.MintDenom, provisionAmt)
}
```


## Parameters

The minting module contains the following parameters:

| Key                    | Type            | Example                |
|------------------------|-----------------|------------------------|
| MintDenom              | string          | "uatom"                |
| SubsidyHalvingInterval | string (dec)    | "1000000.000000000000" |
| GoalBonded             | string (dec)    | "0.670000000000000000" |
| BlocksPerYear          | string (uint64) | "6311520"              |


## Events

The minting module emits the following events:

### BeginBlocker

|  Type   | Attribute Key            | Attribute Value          |
|---------|--------------------------|--------------------------|
| ugdmint | bonded_ratio             | {bondedRatio}            |
| ugdmint | subsidy_halving_interval | {subsidyHalvingInterval} |
| ugdmint | amount                   | {amount}                 |


## Client

### CLI

A user can query and interact with the `ugdmint` module using the CLI.

#### Query

The `query` commands allow users to query `ugdmint` state.

```shell
simd query ugdmint --help
```

##### inflation

The `subsidy-halving-interval` command allow users to query the current minting subsidyHalvingInterval value

```shell
simd query ugdmint subsidy-halving-interval [flags]
```

Example:

```shell
simd query ugdmint subsidy-halving-interval
```

Example Output:

```shell
1000000.00
```

##### params

The `params` command allow users to query the current minting parameters

```shell
simd query ugdmint params [flags]
```

Example:

```yml
blocks_per_year: "4360000"
goal_bonded: "0.670000000000000000"
subsidy_halving_interval: "1000000.000000000000000000"
mint_denom: ugd
```

### gRPC

A user can query the `ugdmint` module using gRPC endpoints.

#### SubsidyHalvingInterval

The `SubsidyHalvingInterval` endpoint allow users to query the current minting subsidyHalvingInterval value

```shell
/cosmos.ugdmint.v1beta1.Query/SubsidyHalvingInterval
```

Example:

```shell
grpcurl -plaintext localhost:9090 cosmos.ugdmint.v1beta1.Query/SubsidyHalvingInterval
```

Example Output:

```json
{
  "subsidyHalvingInterval": "1000000.00"
}
```

#### Params

The `Params` endpoint allow users to query the current minting parameters

```shell
/cosmos.ugdmint.v1beta1.Query/Params
```

Example:

```shell
grpcurl -plaintext localhost:9090 cosmos.ugdmint.v1beta1.Query/Params
```

Example Output:

```json
{
  "params": {
    "mintDenom": "ugd",
    "subsidyHalvingInterval": "1000000.00000000000",
    "goalBonded": "0.670000000000000000",
    "blocksPerYear": "6311520"
  }
}
```

### REST

A user can query the `ugdmint` module using REST endpoints.

#### subsidyHalvingInterval

```shell
/cosmos/ugdmint/v1beta1/subsidyHalvingInterval
```

Example:

```shell
curl "localhost:1317/cosmos/ugdmint/v1beta1/subsidyHalvingInterval"
```

Example Output:

```json
{
  "subsidyHalvingInterval": "1000000.00"
}
```

#### params

```shell
/cosmos/ugdmint/v1beta1/params
```

Example:

```shell
curl "localhost:1317/cosmos/ugdmint/v1beta1/params"
```

Example Output:

```json
{
  "params": {
    "mintDenom": "ugd",
    "subsidyHalvingInterval": "1000000.00000000000",
    "goalBonded": "0.670000000000000000",
    "blocksPerYear": "6311520"
  }
}
```
