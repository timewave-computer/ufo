module github.com/timewave/ufo

go 1.22.0

toolchain go1.22.12

require (
	github.com/gorilla/websocket v1.5.3
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.71.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use fork of osmosis with the UFO modifications
replace github.com/osmosis-labs/osmosis/v28 => ./osmosis-fork/osmosis

// Override any other dependencies as needed to match Osmosis requirements
replace github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.47.5-osmo-v24

replace github.com/cosmos/gogoproto => github.com/cosmos/gogoproto v1.4.10

replace github.com/cometbft/cometbft => github.com/cometbft/cometbft v0.37.4
