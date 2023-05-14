# cosmos-sdk-ugdmint
Unigrid mint module intended to replace standard Mint module for Unigrid Cosmos-SDK chains.  UGDMint implements the Unigrid algorithm for staking rewards instead of the standard algorithm for Cosmos-SDK chains based on inflation and bonded ratio.

## Installation
This module is designed to work with Cosmos-SDK v0.47.x.  

### Private module
Since this module is currently in a private Github repo, first make sure that you have access permissions.  It's easiest to set the `$GOPRIVATE` environment variable to tell `go mod install` to install it from a private repo.  Otherwise it will look at the online public registry of modules and complain that it cannot find it.
<pre>
> export GOPRIVATE="github.com/unigrid-project/cosmos-sdk-ugdmint"
</pre>

### Files
To install this module, we will basically replace the references of the Imports in several files in the `/simapp` folder.  We will modify the following files:
<pre>
- cosmos-sdk
    - simapp
        - app.go
        - app_config.go
        - app_test.go
        - app_v2.go
        - sim_test.go
        - test_helpers.go
        - upgrades.go
</pre>

### Imports
Basically anything that references the Cosmos-SDK Mint module we will now reference our Unigrid UGDMint module instead.  First change the following references in Imports where they are found:

**Before**
<pre>
import (
    ...
    mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
    "github.com/cosmos/cosmos-sdk/x/mint"
    mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
    minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
    ...
)
</pre>
**After**
<pre>
import (
    ...
    mintmodulev1 "<b><i>github.com/unigrid-project/cosmos-sdk-ugdmint/api/cosmos/ugdmint</i></b>/module/v1"
    <b><i>mintmodule</i></b> "github.com/<b><i>unigrid-project/cosmos-sdk-ugdmint/x/ugdmint</i></b>"
    mintkeeper "github.com/<b></i>unigrid-project/cosmos-sdk-ugdmint/x/ugdmint</i></b>/keeper"
    minttypes "github.com/<b><i>unigrid-project/cosmos-sdk-ugdmint/x/ugdmint</i></b>/types"
    ...
)
</pre>

### App.go
As I said before, some files will need a little extra changes as well.  Change the code in App.go as follows:
<pre>
    ModuleBasics = module.NewBasicManager(
        ...
        <b><i>mintmodule</i></b>.AppModuleBasic{},
        ...
    )
</pre>
and
<pre>
    app.ModuleManager = module.NewManager(
        ...
        <b><i>mintmodule</i></b>.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
        ...
    )
</pre>

### App_test.go
There are two places in the code where it constructs a module versionmap.  Change them both as follows:
<pre>
    module.VersionMap{
        ...
        "<b><i>ugdmint</i></b>" <b><i>mintmodule</i></b>.AppModule{}.ConsensusVersion(),
        ...
    }
</pre>

### App_v2.go
One change in the code here:
<pre>
    ModuleBasics = module.NewBasicManager(
        ...
        <b><i>mintmodule</i></b>.AooModuleBasic{},
        ...
    )
</pre>

## Build
After the above changes are made in `/simapp` for your Cosmos-SDK 0.47.x environment, go to the project base folder and run:
``` shell
$ make build
```
It should run a `go mod tidy` before the build that will pull this module down before building.

## Run
Follow the specific documentation for your cosmos chain to setup and run your chain.  If it is base cosmos-sdk, that documentation is found at:

- [Running a Node](https://docs.cosmos.network/v0.47/run-node/run-node)

## Learn more

- [Cosmos SDK docs](https://docs.cosmos.network/v0.47)
