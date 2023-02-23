package signing

import (
	"context"
	"fmt"
	"log"

	"github.com/kwak-labs/liquidation-bot-v2/pkg/queryclient"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"google.golang.org/grpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	leverageTypes "github.com/umee-network/umee/v4/x/leverage/types"
)

type SigningClient struct {
	Grpc_url string

	SigningClient leverageTypes.MsgClient
}

func CreateSigningClient(grpc_url string) SigningClient {
	grpcConn, err := grpc.Dial(
		grpc_url,
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	
	signingClient := leverageTypes.NewMsgClient(grpcConn)

	if err != nil {
		panic(err)
	}

	return SigningClient{
		Grpc_url:      grpc_url,
		SigningClient: signingClient,
	}
}

func (s *SigningClient) Liquidate(target *query.HighestValue, Seed string) {

	// path := hd.BIP44Params{
	// 	Purpose:      44,
	// 	CoinType:     118,
	// 	Account:      0,
	// 	Change:       false,
	// 	AddressIndex: 0,
	// }
	seed := bip39.NewSeed(Seed, "")

	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, "m/44'/118'/0'/0/0'")

	if err != nil {
		log.Fatal(err)
	}

	var privKey secp256k1.PrivKey = secp256k1.PrivKey(priv)
	var pubKey crypto.PubKey = privKey.PubKey()

	var AccAdd = sdk.AccAddress(pubKey.Address().Bytes())

	// ctx := sdk.UnwrapSDKContext(context.Background())

	borrower, err := sdk.AccAddressFromBech32(target.Address)

	asset, err := sdk.ParseCoinNormalized(fmt.Sprintf("%s%s", target.Supplied.Amount, target.Supplied.Denom))

	if err != nil {
		fmt.Println(err)
	}

	msg := leverageTypes.NewMsgLiquidate(AccAdd, borrower, asset, target.Collateral.Denom)
	
	if err = msg.ValidateBasic(); err != nil {
		fmt.Println(err)
	}
	
	res, err := s.SigningClient.Liquidate(context.Background(), msg)

	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
