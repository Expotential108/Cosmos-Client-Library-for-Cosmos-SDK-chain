package client

import (
	"io"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	provtypes "github.com/tendermint/tendermint/light/provider"
	prov "github.com/tendermint/tendermint/light/provider/http"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type ChainClient struct {
	Config         *ChainClientConfig
	Keybase        keyring.Keyring
	KeyringOptions []keyring.Option
	RPCClient      rpcclient.Client
	LightProvider  provtypes.Provider
	Input          io.Reader
	Output         io.Reader
	// TODO: GRPC Client type?

	Codec  Codec
	Logger log.Logger
}

func NewChainClient(ccc *ChainClientConfig, input io.Reader, output io.Writer, kro ...keyring.Option) (*ChainClient, error) {
	if err := ccc.Validate(); err != nil {
		return nil, err
	}

	// TODO: test key directory and return error if not created
	keybase, err := keyring.New(ccc.ChainID, ccc.KeyringBackend, ccc.KeyDirectory, nil)
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to deal with input or maybe just make all keyring backends test?

	timeout, _ := time.ParseDuration(ccc.Timeout)
	rpcClient, err := newRPCClient(ccc.RPCAddr, timeout)
	if err != nil {
		return nil, err
	}

	lightprovider, err := prov.New(ccc.ChainID, ccc.RPCAddr)
	if err != nil {
		return nil, err
	}

	return &ChainClient{
		Keybase:        keybase,
		RPCClient:      rpcClient,
		LightProvider:  lightprovider,
		KeyringOptions: kro,
		Config:         ccc,
		Codec:          MakeCodec(ccc.Modules),
		Logger:         log.NewTMLogger(log.NewSyncWriter(output)),
	}, nil
}

func (cc *ChainClient) GetKeyAddress() (sdk.AccAddress, error) {
	// TODO: implement
	return nil, nil
}

func newRPCClient(addr string, timeout time.Duration) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = timeout
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}
