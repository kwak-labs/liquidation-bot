package query

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"

	leverageTypes "github.com/umee-network/umee/v4/x/leverage/types"
)

type QueryClient struct {
	grpc_url string
	rest_url string

	queryClient leverageTypes.QueryClient
}

type Targets struct {
	Targets []string
}

type HighestValue struct {
	Address    string
	Collateral Collateral
	Supplied   Supplied
}

type Collateral struct {
	Denom  string
	Amount math.Int
	Usd    float64
}

type Supplied struct {
	Denom  string
	Amount math.Int
}

var (
	BadAddress []string
)

func CreateQueryClient(grpc_url string, rest_url string) QueryClient {
	grpcConn, err := grpc.Dial(
		grpc_url,
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	queryClient := leverageTypes.NewQueryClient(grpcConn)

	if err != nil {
		panic(err)
	}

	return QueryClient{
		grpc_url:    grpc_url,
		queryClient: queryClient,
		rest_url:    rest_url,
	}
}

func (q *QueryClient) accountBalances(address string) *leverageTypes.QueryAccountBalancesResponse {
	req := &leverageTypes.QueryAccountBalances{
		Address: address,
	}
	resp, err := q.queryClient.AccountBalances(context.Background(), req)

	if err != nil {
		panic(err)
	}

	return resp
}

func (q *QueryClient) accountSummary(address string) *leverageTypes.QueryAccountSummaryResponse {
	req := &leverageTypes.QueryAccountSummary{
		Address: address,
	}
	resp, err := q.queryClient.AccountSummary(context.Background(), req)

	if err != nil {
		panic(err)
	}

	return resp
}

// Fetch the target worth most
func (q *QueryClient) FetchTarget() (HighestValue, error) {
	resp, err := http.Get(q.rest_url)

	var targets Targets
	var highestValue HighestValue

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&targets)
	if len(targets.Targets) <= 0 {
		return highestValue, errors.New("No valid targets")
	}

	for _, address := range targets.Targets {
		var accountSummary = q.accountSummary(address)

		var balanceParsed, err = strconv.ParseFloat(accountSummary.CollateralValue.String(), 64)

		if err != nil {
			panic(err)
		}

		if balanceParsed > highestValue.Collateral.Usd {
			accountBalance := q.accountBalances(address)
			var cDenom string = accountBalance.Collateral.GetDenomByIndex(0)
			var cAmount math.Int = accountBalance.Collateral.AmountOf(cDenom)

			var sDenom string = accountBalance.Supplied.GetDenomByIndex(0)
			var sAmount math.Int = accountBalance.Supplied.AmountOf(sDenom)

			highestValue = HighestValue{
				address,
				Collateral{
					Denom:  cDenom,
					Amount: cAmount,
					Usd:    balanceParsed,
				},
				Supplied{
					Denom:  sDenom,
					Amount: sAmount,
				},
			}

			fmt.Println(highestValue)
		}
	}

	return highestValue, nil
}
