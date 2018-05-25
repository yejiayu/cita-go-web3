package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cryptape/go-web3/types"
	"github.com/cryptape/go-web3/utils"
	"github.com/golang/protobuf/proto"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

const (
	sendTransactionMethod = "cita_sendTransaction"

	saveAbiAddress = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

// SendTransaction injects the transaction into the pending pool for execution.
func (b *backend) SendTransaction(ctx context.Context, tx *types.Transaction, hexPrivateKey string) (common.Hash, error) {
	sign, err := genSign(tx, hexPrivateKey)
	if err != nil {
		return common.Hash{}, err
	}

	unTx := &types.UnverifiedTransaction{
		Transaction: tx,
		Signature:   sign,
		Crypto:      types.Crypto_SECP,
	}

	data, err := proto.Marshal(unTx)
	if err != nil {
		return common.Hash{}, err
	}

	resp, err := b.provider.SendRequest(sendTransactionMethod, hex.EncodeToString(data))
	if err != nil {
		return common.Hash{}, err
	}

	var status types.TransactionStatus
	if err := resp.GetObject(&status); err != nil {
		return common.Hash{}, err
	}

	if status.Status != "OK" {
		return common.Hash{}, fmt.Errorf("send transaction failed, status is %s", status.Status)
	}
	return common.HexToHash(status.Hash), nil
}

// DeployContract deploys a contract onto the Ethereum blockchain and binds the
// deployment address with a Go wrapper.
func (b *backend) DeployContract(ctx context.Context, params *types.TransactParams, abi abi.ABI, code string) (*types.Transaction, *BoundContract, error) {
	codeB, err := hex.DecodeString(utils.CleanHexPrefix(code))
	if err != nil {
		return nil, nil, err
	}

	tx := &types.Transaction{
		To:              "",
		Data:            codeB,
		ValidUntilBlock: params.ValidUntilBlock.Uint64(),
		ChainId:         params.ChainID,
		Nonce:           params.Nonce,
		Quota:           params.Quota.Uint64(),
	}

	txHash, err := b.SendTransaction(ctx, tx, params.HexPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(txHash)
	return nil, nil, nil
}

func genSign(tx *types.Transaction, hexPrivateKey string) ([]byte, error) {
	if strings.HasPrefix(hexPrivateKey, "0x") {
		hexPrivateKey = strings.TrimPrefix(hexPrivateKey, "0x")
	}

	txB, err := proto.Marshal(tx)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	h := sha3.New256()
	h.Write(txB)
	hash := h.Sum(nil)
	return crypto.Sign(hash, privateKey)
}
