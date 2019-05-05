package rpc_test

import (
	"context"
	"fmt"
	"github.com/DSiSc/astraia/client"
	"github.com/cespare/cp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

const (
	subscribeTimeout     = 5 * time.Second
	)

func tmpDatadirWithKeystore(t *testing.T) string {
	datadir := tmpdir(t)
	keystore := filepath.Join(datadir, "keystore")
	source := filepath.Join("..", "cmd", "testdata", "keystore")
	if err := cp.CopyAll(keystore, source); err != nil {
		t.Fatal(err)
	}
	return datadir
}

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "geth-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}


func TestExampleClientSubscription(t *testing.T) {
	// Connect the client.
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	result := ""
	err := client.Call(&result, "eth_blockNumber")
	fmt.Println(err)
	fmt.Println(result)
}

func TestClient_NewAccount(t *testing.T) {
	var result map[string]string
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	ctx, cancel := context.WithTimeout(context.Background(), subscribeTimeout)
	defer cancel()

	err := client.CallContext(ctx, &result, "personal_listAccounts", "")
	assert.Equal(t, nil, err)
	fmt.Println(result)
}

func TestClient_Accounts(t *testing.T) {
	var result map[string]string
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	ctx, cancel := context.WithTimeout(context.Background(), subscribeTimeout)
	defer cancel()

	err := client.CallContext(ctx, &result, "rpc_modules")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "personal_listAccounts", "")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "personal_unlockAccount", "0x1b192c4e353dc40871066023bf37fc632f1695d4", "123")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "personal_lockAccount", "0x1b192c4e353dc40871066023bf37fc632f1695d4")
	assert.Equal(t, nil, err)
	fmt.Println(result)


}

func TestClient_Newweb3(t *testing.T) {
	var result map[string]string
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	ctx, cancel := context.WithTimeout(context.Background(), subscribeTimeout)
	defer cancel()

	err := client.CallContext(ctx, &result, "eth_newWeb3", "127.0.0.1", "47768")
	assert.Equal(t, nil, err)
	fmt.Println(result)
}

func TestClient_Eth(t *testing.T){
	var result map[string]string
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	ctx, cancel := context.WithTimeout(context.Background(), subscribeTimeout)
	defer cancel()

	err := client.CallContext(ctx, &result, "eth_getBalance", "0x1b192c4e353dc40871066023bf37fc632f1695d4")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "eth_getTransactionCount", "0x1b192c4e353dc40871066023bf37fc632f1695d4")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	txStr := "{from: \"0x1b192c4e353dc40871066023bf37fc632f1695d4\", to: \"0x1b192c4e353dc40871066023bf37fc632f1695d4\", value: web3.toWei(1, \"wei\"), gas:\"123\", gasPrice:\"123\", nonce:\"1\"}"
	err = client.CallContext(ctx, &result, "personal_signTransaction", txStr, "123")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "eth_sendRawTransaction", "0xf877f872017b8094b0c066aa7f29c34f5ad32f900e2349c9dba9642e94b0c066aa7f29c34f5ad32f900e2349c9dba9642e01801ca08165c224ab38cf41325fbae788749f55c875cbb0379b90565f5dfc77c4b8593aa033574624fda3a88a5a976fca71e80e68e9def011e297af85cd743fd05439a6acc0c0c0")
	assert.Equal(t, nil, err)
	fmt.Println(result)

	err = client.CallContext(ctx, &result, "eth_getTransactionByHash", "0xef8dadde66af80a228e4899055f2ff202d3e6107904b0f05316ebfcc7a31a850")
	assert.Equal(t, nil, err)
	fmt.Println(result)
}
