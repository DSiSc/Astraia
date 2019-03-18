package api

import (
	"errors"
	"flag"
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/lightClient/config"
	"github.com/DSiSc/wallet/common"
	local "github.com/DSiSc/wallet/core/types"
	web3cmn "github.com/DSiSc/web3go/common"
	"github.com/DSiSc/web3go/provider"
	"github.com/DSiSc/web3go/rpc"
	"github.com/DSiSc/web3go/web3"
	"strconv"
)

//Send a signed transaction
func SendTransaction(tx *types.Transaction) (common.Hash, error) {
	//format 0x string
	from := fmt.Sprintf("0x%x", *(tx.Data.From))
	to := from
	gas := "0x" + strconv.FormatInt(int64(tx.Data.GasLimit),16)
	gasprice := "0x" + tx.Data.Price.String()
	value := "0x" + tx.Data.Amount.String()
	data := ""

	if tx.Data.Payload != nil {
		data = "0x" + string(tx.Data.Payload)
	} else {
		data = ""
	}

	configHostName := config.GetApiGatewayHostName()
	hostname := flag.String("hostname", configHostName, "The ethereum client RPC host")
	configPort := config.GetApiGatewayPort()
	port := flag.String("port", configPort, "The ethereum client RPC port")
	verbose := flag.Bool("verbose", true, "Print verbose messages")

	if *verbose {
		fmt.Printf("Connect to %s:%s\n", *hostname, *port)
	}

	provider := provider.NewHTTPProvider(*hostname+":"+*port, rpc.GetDefaultMethod())
	web3 := web3.NewWeb3(provider)

	req := &web3cmn.TransactionRequest{
		From:     from,
		To:       to,
		Gas:      gas,
		GasPrice: gasprice,
		Value:    value,
		Data:     data,
	}

	hash, err := web3.Eth.SendTransaction(req)
	return common.Hash(hash), err
}

func SendRawTransaction(tx *types.Transaction) (common.Hash, error) {

	configHostName := config.GetApiGatewayHostName()
	hostname := flag.String("hostname", configHostName, "The ethereum client RPC host")
	configPort := config.GetApiGatewayPort()
	port := flag.String("port", configPort, "The ethereum client RPC port")
	verbose := flag.Bool("verbose", false, "Print verbose messages")

	if *verbose {
		fmt.Printf("Connect to %s:%s\n", *hostname, *port)
	}

	provider := provider.NewHTTPProvider(*hostname+":"+*port, rpc.GetDefaultMethod())
	web3 := web3.NewWeb3(provider)

	txBytes, _ := local.EncodeToRLP(tx)
	hash, err := web3.Eth.SendRawTransaction(txBytes)

	return common.Hash(hash), err
}

func GetTransactionByHash(web *web3.Web3, txHash string) ( *web3cmn.Transaction, error) {
	if web == nil {
		return nil, errors.New("GetTransactionByHashWbe3 has call error web is nil")
	}

	bytes := web3cmn.HexToBytes(txHash)
	tx, err :=web.Eth.GetTransactionByHash(web3cmn.NewHash(bytes))
	if err != nil {
		return nil, err
	}
	return tx, err
}

func GetTransactionCount(web *web3.Web3, addr string, quantity string) (string, error) {
	if web == nil {
		return "", errors.New("GetTransactionCount has call error web is nil")
	}

	address := web3cmn.StringToAddress(addr)
	if quantity == "" {
		quantity = "pending"
	}
	count , err := web.Eth.GetTransactionCount(address, quantity)
	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("0x%x", count.Uint64())
	return result, err
}
