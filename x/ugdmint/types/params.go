package types

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance
func NewParams(
	mintDenom string, subsidyHalvingInterval, goalBonded sdk.Dec, blocksPerYear uint64,
) Params {
	return Params{
		MintDenom:              mintDenom,
		SubsidyHalvingInterval: subsidyHalvingInterval,
		GoalBonded:             goalBonded,
		BlocksPerYear:          blocksPerYear,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MintDenom:              "ugd",
		SubsidyHalvingInterval: sdk.NewDecWithPrec(100000000, 2),
		GoalBonded:             sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:          uint64(60 * 60 * 8766 / 5),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateSubsidyHalvingInterval(p.SubsidyHalvingInterval); err != nil {
		return err
	}
	if err := validateGoalBonded(p.GoalBonded); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateSubsidyHalvingInterval(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("subsidy halving interval cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("subsidy halving interval cannot be negative: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("goal bonded cannot be nil: %s", v)
	}
	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("goal bonded must be positive: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}
