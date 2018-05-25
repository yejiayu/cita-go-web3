package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/cryptape/go-web3/utils"

	"github.com/ethereum/go-ethereum/common"
)

const (
	getAbiMethod  = "eth_getAbi"
	ethCallMethod = "eth_call"
)

type CallMsg struct {
	From common.Address
	To   *common.Address
	Data []byte
}

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (b *backend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (string, error) {
	blockNumberHex := "latest"
	if blockNumber != nil {
		blockNumberHex = blockNumber.Text(16)
	}

	fmt.Println(blockNumberHex, contract.String())
	resp, err := b.provider.SendRequest(getAbiMethod, contract.String(), blockNumberHex)
	if err != nil {
		return "", err
	}
	return resp.GetString()
}

// ContractCall executes an Ethereum contract call with the specified data as the
// input.
func (b *backend) CallContract(ctx context.Context, result interface{}, call CallMsg, blockNumber *big.Int) error {
	data := utils.AddHexPrefix(hex.EncodeToString(call.Data))
	params := map[string]string{"from": call.From.Hex(), "data": data, "to": ""}
	if call.To != nil {
		params["to"] = call.To.Hex()
	}

	numHex := "latest"
	if blockNumber != nil {
		numHex = blockNumber.Text(16)
	}

	fmt.Println(params, numHex)
	resp, err := b.provider.SendRequest(ethCallMethod, data, numHex)
	if err != nil {
		return err
	}

	return resp.GetObject(result)
}
