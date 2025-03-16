module github.com/timewave/ufo

go 1.21

// Use fork of osmosis with the UFO modifications
replace github.com/osmosis-labs/osmosis/v28 => ./osmosis-fork/osmosis

// Override any other dependencies as needed to match Osmosis requirements
replace github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.47.5-osmo-v24

replace github.com/cosmos/gogoproto => github.com/cosmos/gogoproto v1.4.10

replace github.com/cometbft/cometbft => github.com/cometbft/cometbft v0.37.4
