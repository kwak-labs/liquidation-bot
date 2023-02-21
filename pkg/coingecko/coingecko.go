package cg

import (
	"log"
	"net/http"
	"time"

	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
	coingecko "github.com/superoo7/go-gecko/v3"
)

var (
	coins = []string{"cosmos", "axlusdc", "wrapped-bitcoin", "ethereum", "stride-staked-osmo", "binancecoin", "dai", "stride-staked-atom", "osmosis", "inter-stable-token", "umee"}
	coin  map[string]string
)

// Start loading all the prices into a map
func CachePrices(prices *map[string]float32) {
	c := cron.New()

	coinNames, err := os.ReadFile("./config/coingecko.toml")

	if err != nil {
		log.Fatal(err)
	}

	err = toml.Unmarshal(coinNames, &coin)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	CG := coingecko.NewClient(httpClient)
	makeRequest(CG, prices)()
	c.AddFunc("@every 5m", makeRequest(CG, prices))
	c.Start()
}

func makeRequest(CG *coingecko.Client, prices *map[string]float32) func() {
	return func() {
		sp, err := CG.SimplePrice(coins, []string{"usd"})
		if err != nil {
			log.Fatal(err)
		}

		for key, element := range *sp {
			(*prices)[coin[key]] = element["usd"]
		}
	}
}
