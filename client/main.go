package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"

	"github.com/kwak-labs/liquidation-bot-v2/pkg/coingecko"
	"github.com/kwak-labs/liquidation-bot-v2/pkg/queryclient"
	"github.com/kwak-labs/liquidation-bot-v2/pkg/signingclient"
)

type Config struct {
	Seed_phrase   string
	Grpc_endpoint string
	Rest_endpoint string
	Min_usd_value float64
	Interval      int
	Memo          string

	Gas struct {
		Denom string
		Gas   string
		Fee   string
	}
}

var (
	prices     = make(map[string]float32)
	incentives = make(map[string]float32)
	configData Config
)

func main() {
	color.Blue("  Starting Liquidation Bot\n\n")

	color.Cyan("Fetching price data")

	cg.CachePrices(&prices)

	color.Green("Loading Config Data")

	configdata, err := os.ReadFile("./config/main.toml")
	err = toml.Unmarshal(configdata, &configData)

	if err != nil {
		panic(err)
	}

	color.Yellow("Loading Incentives\n\n")

	incentiveData, err := os.ReadFile("./config/incentives.toml")
	err = toml.Unmarshal(incentiveData, &incentives)

	if err != nil {
		panic(err)
	}

	color.Magenta("Starting Bot...")

	qClient := query.CreateQueryClient(configData.Grpc_endpoint, configData.Rest_endpoint)
	sClient := signing.CreateSigningClient(configData.Grpc_endpoint)

	target, err := qClient.FetchTarget()
	if err != nil {
		return
	}
	
	sClient.Liquidate(&target, configData.Seed_phrase)

	// Runs when the collateral is less then what the defined
	if target.Collateral.Usd < configData.Min_usd_value {
		return
	}
	gas, err := strconv.ParseFloat(configData.Gas.Fee, 32)

	if err != nil {
		fmt.Println(err)
		panic("Gas fee is not a string")
	}

	gasUsd := float32(gas) * prices["umee"]

	reward := target.Collateral.Usd * float64(incentives[target.Collateral.Denom])

	// Runs when gas fee makes it unprofitable
	if reward < float64(gasUsd) {
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
