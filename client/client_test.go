package rpc_test

import (

	"fmt"
	"github.com/DSiSc/lightClient/client"
	"testing"

)

func TestExampleClientSubscription(t *testing.T) {
	// Connect the client.
	client, _ := rpc.Dial("http://127.0.0.1:47768")
	result := ""
	err := client.Call(&result, "eth_blockNumber")
	fmt.Println(err)
	fmt.Println(result)
}