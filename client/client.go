// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DSiSc/crypto-suite/rlp"
	"github.com/DSiSc/lightClient/api"
	"github.com/DSiSc/lightClient/config"
	"github.com/DSiSc/p2p/common"
	wcommon "github.com/DSiSc/wallet/common"
	"math/big"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/DSiSc/craft/log"
	ctypes "github.com/DSiSc/craft/types"
	"github.com/DSiSc/wallet/accounts/keystore"
	wutils "github.com/DSiSc/wallet/utils"
	web3cmn "github.com/DSiSc/web3go/common"
	"github.com/DSiSc/web3go/web3"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	errClientReconnected         = errors.New("client reconnected")
	errDead                      = errors.New("connection lost")
)

const (
	// Timeouts
	tcpKeepAliveInterval = 30 * time.Second
	defaultDialTimeout   = 10 * time.Second // used if context has no deadline
	subscribeTimeout     = 5 * time.Second  // overall timeout eth_subscribe, rpc_modules calls
)

type Tx map[string]string

const (
	// Subscriptions are removed when the subscriber cannot keep up.
	//
	// This can be worked around by supplying a channel with sufficiently sized buffer,
	// but this can be inconvenient and hard to explain in the docs. Another issue with
	// buffered channels is that the buffer is static even though it might not be needed
	// most of the time.
	//
	// The approach taken here is to maintain a per-subscription linked list buffer
	// shrinks on demand. If the buffer reaches the size below, the subscription is
	// dropped.
	maxClientSubscriptionBuffer = 20000
)

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Args   []interface{}
	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result interface{}
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}

// Client represents a connection to an RPC server.
type Client struct {
	//idgen    func() ID // for subscriptions
	isHTTP   bool
	//services *serviceRegistry

	isLocal bool

	idCounter uint32

	// This function, if non-nil, is called when the connection is lost.
	reconnectFunc reconnectFunc

	// writeConn is used for writing to the connection on the caller's goroutine. It should
	// only be accessed outside of dispatch, with the write lock held. The write lock is
	// taken by sending on requestOp and released by sending on sendDone.
	writeConn jsonWriter

	//use to manager keystore wallets
	keystore *keystore.KeyStore

	//use to call apigateway
	web3 *web3.Web3

	// for dispatch
	close       chan struct{}
	closing     chan struct{}    // closed when client is quitting
	didClose    chan struct{}    // closed when client quits
	reconnected chan ServerCodec // where write/reconnect sends the new connection
	readOp      chan readOp      // read messages
	readErr     chan error       // errors from read
	reqInit     chan *requestOp  // register response IDs, takes write lock
	reqSent     chan error       // signals write completion, releases write lock
	reqTimeout  chan *requestOp  // removes response IDs when call timeout expires
}

type reconnectFunc func(ctx context.Context) (ServerCodec, error)

type clientContextKey struct{}

type readOp struct {
	msgs  []*jsonrpcMessage
	batch bool
}

type requestOp struct {
	ids  []json.RawMessage
	err  error
	resp chan *jsonrpcMessage // receives up to len(ids) responses
	//sub  *ClientSubscription  // only set for EthSubscribe requests
}

func (op *requestOp) wait(ctx context.Context, c *Client) (*jsonrpcMessage, error) {
	select {
	case <-ctx.Done():
		// Send the timeout to dispatch so it can remove the request IDs.
		select {
		case c.reqTimeout <- op:
		case <-c.closing:
		}
		return nil, ctx.Err()
	case resp := <-op.resp:
		return resp, op.err
	}
}

// Dial creates a new client for the given URL.
//
// The currently supported URL schemes are "http", "https", "ws" and "wss". If rawurl is a
// file name with no URL scheme, a local socket connection is established using UNIX
// domain sockets on supported platforms and named pipes on Windows. If you want to
// configure transport options, use DialHTTP, DialWebsocket or DialIPC instead.
//
// For websocket connections, the origin is set to the local host name.
//
// The client reconnects automatically if the connection is lost.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

// DialContext creates a new RPC client, just like Dial.
//
// The context is used to cancel or time out the initial connection establishment. It does
// not affect subsequent interactions with the client.
func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		return DialHTTP(rawurl)
	default:
		return nil, fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}

// Client retrieves the client from the context, if any. This can be used to perform
// 'reverse calls' in a handler method.
func ClientFromContext(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(clientContextKey{}).(*Client)
	return client, ok
}

func newClient(initctx context.Context, connect reconnectFunc) (*Client, error) {
	conn, err := connect(initctx)
	if err != nil {
		return nil, err
	}
	//c := initClient(conn, randomIDGenerator(), new(serviceRegistry))
	c := initClient(conn)
	c.reconnectFunc = connect
	return c, nil
}

//func initClient(conn ServerCodec, idgen func() ID, services *serviceRegistry) *Client {
func initClient(conn ServerCodec) *Client {
	_, isHTTP := conn.(*httpConn)

	keyStoreDir :=  keystore.KeyStoreScheme
	scryptN, scryptP, keydir, err := wutils.AccountConfig(keyStoreDir)
	if err != nil {
		fmt.Println("client init failed, err = ", err)
	}

	_keystore := keystore.NewKeyStore(keydir, scryptN, scryptP)

	hostname := config.GetApiGatewayHostName()
	port := config.GetApiGatewayPort()
	web, _ := wutils.NewWeb3(hostname, port, false)
	c := &Client{
		//idgen:       idgen,
		isHTTP:      isHTTP,
		isLocal:     true,
		keystore:    _keystore,
		web3:        web,
		//services:    services,
		writeConn:   conn,
		close:       make(chan struct{}),
		closing:     make(chan struct{}),
		didClose:    make(chan struct{}),
		reconnected: make(chan ServerCodec),
		readOp:      make(chan readOp),
		readErr:     make(chan error),
		reqInit:     make(chan *requestOp),
		reqSent:     make(chan error, 1),
		reqTimeout:  make(chan *requestOp),
	}
	if !isHTTP {
		//go c.dispatch(conn)
	}
	return c
}

func (c *Client) setWeb3(web *web3.Web3) error {
	if web == nil {
		return errors.New("set web3 can't be nil")
	}
	c.web3 = web
	return nil
}

// RegisterName creates a service for the given receiver type under the given name. When no
// methods on the given receiver match the criteria to be either a RPC method or a
// subscription an error is returned. Otherwise a new service is created and added to the
// service collection this client provides to the server.
//func (c *Client) RegisterName(name string, receiver interface{}) error {
//	return c.services.registerName(name, receiver)
//}

func (c *Client) nextID() json.RawMessage {
	id := atomic.AddUint32(&c.idCounter, 1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

// SupportedModules calls the rpc_modules method, retrieving the list of
// APIs that are available on the server.
func (c *Client) SupportedModules() (map[string]string, error) {
	var result map[string]string
	ctx, cancel := context.WithTimeout(context.Background(), subscribeTimeout)
	defer cancel()
	err := c.CallContext(ctx, &result, "rpc_modules")
	return result, err
}

// Close closes the client, aborting any in-flight requests.
func (c *Client) Close() {
	if c.isHTTP {
		return
	}
	select {
	case c.close <- struct{}{}:
		<-c.didClose
	case <-c.didClose:
	}
}

// Call performs a JSON-RPC call with the given arguments and unmarshals into
// result if no error occurred.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
func (c *Client) Call(result interface{}, method string, args ...interface{}) error {
	ctx := context.Background()
	return c.CallContext(ctx, result, method, args...)
}

// CallContext performs a JSON-RPC call with the given arguments. If the context is
// canceled before the call has successfully returned, CallContext returns immediately.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
func (c *Client) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	msg, err := c.newMessage(method, args...)
	if err != nil {
		return err
	}
	op := &requestOp{ids: []json.RawMessage{msg.ID}, resp: make(chan *jsonrpcMessage, 1)}

	if c.isHTTP {

		//no attach a node,call local function
		err = c.sendLocal(ctx, op, msg)
		//err = c.sendHTTP(ctx, op, msg)

	} else {
		err = c.send(ctx, op, msg)
	}
	if err != nil {
		return err
	}

	// dispatch has accepted the request and will close the channel when it quits.
	switch resp, err := op.wait(ctx, c); {
	case err != nil:
		return err
	case resp.Error != nil:
		return resp.Error
	case len(resp.Result) == 0:
		return ErrNoResult
	default:
		return json.Unmarshal(resp.Result, &result)
	}
}

// BatchCall sends all given requests as a single batch and waits for the server
// to return a response for all of them.
//
// In contrast to Call, BatchCall only returns I/O errors. Any error specific to
// a request is reported through the Error field of the corresponding BatchElem.
//
// Note that batch calls may not be executed atomically on the server side.
func (c *Client) BatchCall(b []BatchElem) error {
	ctx := context.Background()
	return c.BatchCallContext(ctx, b)
}

// BatchCall sends all given requests as a single batch and waits for the server
// to return a response for all of them. The wait duration is bounded by the
// context's deadline.
//
// In contrast to CallContext, BatchCallContext only returns errors that have occurred
// while sending the request. Any error specific to a request is reported through the
// Error field of the corresponding BatchElem.
//
// Note that batch calls may not be executed atomically on the server side.
func (c *Client) BatchCallContext(ctx context.Context, b []BatchElem) error {
	msgs := make([]*jsonrpcMessage, len(b))
	op := &requestOp{
		ids:  make([]json.RawMessage, len(b)),
		resp: make(chan *jsonrpcMessage, len(b)),
	}
	for i, elem := range b {
		msg, err := c.newMessage(elem.Method, elem.Args...)
		if err != nil {
			return err
		}
		msgs[i] = msg
		op.ids[i] = msg.ID
	}

	var err error
	if c.isHTTP {
		err = c.sendBatchHTTP(ctx, op, msgs)
	} else {
		err = c.send(ctx, op, msgs)
	}

	// Wait for all responses to come back.
	for n := 0; n < len(b) && err == nil; n++ {
		var resp *jsonrpcMessage
		resp, err = op.wait(ctx, c)
		if err != nil {
			break
		}
		// Find the element corresponding to this response.
		// The element is guaranteed to be present because dispatch
		// only sends valid IDs to our channel.
		var elem *BatchElem
		for i := range msgs {
			if bytes.Equal(msgs[i].ID, resp.ID) {
				elem = &b[i]
				break
			}
		}
		if resp.Error != nil {
			elem.Error = resp.Error
			continue
		}
		if len(resp.Result) == 0 {
			elem.Error = ErrNoResult
			continue
		}
		elem.Error = json.Unmarshal(resp.Result, elem.Result)
	}
	return err
}

// Notify sends a notification, i.e. a method call that doesn't expect a response.
func (c *Client) Notify(ctx context.Context, method string, args ...interface{}) error {
	op := new(requestOp)
	msg, err := c.newMessage(method, args...)
	if err != nil {
		return err
	}
	msg.ID = nil

	if c.isHTTP {
		return c.sendHTTP(ctx, op, msg)
	} else {
		return c.send(ctx, op, msg)
	}
}

func (c *Client) newMessage(method string, paramsIn ...interface{}) (*jsonrpcMessage, error) {
	msg := &jsonrpcMessage{Version: vsn, ID: c.nextID(), Method: method}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		if msg.Params, err = json.Marshal(paramsIn); err != nil {
			return nil, err
		}
	}
	return msg, nil
}

// send registers op with the dispatch loop, then sends msg on the connection.
// if sending fails, op is deregistered.
func (c *Client) send(ctx context.Context, op *requestOp, msg interface{}) error {
	select {
	case c.reqInit <- op:
		err := c.write(ctx, msg)
		c.reqSent <- err
		return err
	case <-ctx.Done():
		// This can happen if the client is overloaded or unable to keep up with
		// subscription notifications.
		return ctx.Err()
	case <-c.closing:
		return ErrClientQuit
	}
}

func (c *Client) sendLocal(ctx context.Context, op *requestOp, msg *jsonrpcMessage) error {
	result := []string{}
	json.Unmarshal(msg.Params, &result)
	jsonReusult := []byte{'{', '"', 'p', 'e', 'r', 's', 'o', 'n', 'a', 'l', '"', ':', '"', '1', '.', '0', '"', '}'}

	switch msg.Method {
	case "personal_newAccount":
		wutils.NewAccount("", result[0])
		break
	case "personal_listAccounts":
		keystoreDir := result[0]
		wutils.ListAccounts(keystoreDir)
		break
	case "personal_unlockAccount":
		addr := result[0]
		password := result[1]
		err := wutils.Unlock(c.keystore, addr, password)
		if err != nil {
			msg := fmt.Sprintf("unlockAccount failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
		}
		break
	case "personal_lockAccount":
		addr := result[0]
		err := wutils.Lock(c.keystore, addr)
		if err != nil {
			msg := fmt.Sprintf("lockAccount failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
		}
		break

	case "eth_getTransactionCount":
		addr := result[0]
		quantity := result[1]

		count, err := api.GetTransactionCount(c.web3, addr, quantity)
		if err != nil {
			msg := fmt.Sprintf("eth_getTransactionCount failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}
		jsonReusult, _ = json.Marshal(count)

		break
		
	case "eth_sendTransaction":
		result := make([]Tx, 1)
		err := json.Unmarshal(msg.Params, &result)
		if err != nil {
			msg := fmt.Sprintf("eth_sendTransaction, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}

		from := result[0]["from"]
		to := result[0]["to"]
		gasPrice := result[0]["gasPrice"]
		gas := result[0]["gas"]
		value := result[0]["value"]
		data := result[0]["payload"]

		req := &web3cmn.TransactionRequest{
			From:     from,
			To:       to,
			Gas:      gas,
			GasPrice: gasPrice,
			Value:    value,
			Data:     data,
		}

		hash, err := wutils.SendTransactionWeb3(req)
		if err != nil {
			msg := fmt.Sprintf("eth_sendTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)

			break
		}

		jsonReusult, _ = json.Marshal(hash.String())

		break

	case "eth_sendRawTransaction":
		hash, err := wutils.SendRawTransactionWeb3(c.web3, result[0])
		if err != nil {
			msg := fmt.Sprintf("sendRawTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}
		//fmt.Println(hash.String())
		jsonReusult, _ = json.Marshal(hash.String())

		break
	case "eth_getTransactionByHash":
		tx, err := api.GetTransactionByHash(c.web3, result[0])
		if err != nil {
			msg := fmt.Sprintf("eth_getTransactionByHash failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}

		jsonReusult, _ = json.Marshal(tx.String())
		break

	case "eth_newWeb3":
		hostname := result[0]
		port := result[1]
		web, err := wutils.NewWeb3(hostname, port, false)
		if err != nil {
			msg := fmt.Sprintf("eth_newWeb3 failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}
		c.setWeb3(web)
		jsonReusult, _ = json.Marshal("new dial http:// " + hostname + ":" + port)
		break

		break
	//TODO: verify legal(important)
	case "personal_signTransaction":
		var rawMsg []json.RawMessage
		err := json.Unmarshal(msg.Params, &rawMsg)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}

		var result Tx
		err = json.Unmarshal(rawMsg[0], &result)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}

		var password string
		err = json.Unmarshal(rawMsg[1], &password)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}

		var transaction ctypes.Transaction
		transaction, err = TxToTransaction(result)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction failed, err = %v", err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}
		signed, err := wutils.SignTxByPassWord(&transaction, password)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction failed, tx = %s, err = %v", result, err)
			jsonReusult, _ = json.Marshal(msg)
			break
		}
		data, err := rlp.EncodeToBytes(signed)
		if err != nil {
			msg := fmt.Sprintf("personal_signTransaction rlp encode failed, tx = %s, err = %v", result, err)
			jsonReusult, _ = json.Marshal(msg)
		}

		jsonReusult, _ = json.Marshal(wcommon.ToHex(data))

		break

	default:
		//jsonReusult, _ = json.Marshal("This function is not currently implemented")
	}

	respmsg := jsonrpcMessage{
		Method:msg.Method,
		Params:msg.Params,
		Result:jsonReusult,
	}

	op.resp <- &respmsg
	return nil
}

func (c *Client) write(ctx context.Context, msg interface{}) error {
	// The previous write failed. Try to establish a new connection.
	if c.writeConn == nil {
		if err := c.reconnect(ctx); err != nil {
			return err
		}
	}

	err := c.writeConn.Write(ctx, msg)
	if err != nil {
		c.writeConn = nil
	}
	return err
}

func (c *Client) reconnect(ctx context.Context) error {
	if c.reconnectFunc == nil {
		return errDead
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, defaultDialTimeout)
		defer cancel()
	}
	newconn, err := c.reconnectFunc(ctx)
	if err != nil {
		log.Error("RPC client reconnect failed" + "err" + err.Error())
		return err
	}
	select {
	case c.reconnected <- newconn:
		c.writeConn = newconn
		return nil
	case <-c.didClose:
		newconn.Close()
		return ErrClientQuit
	}
}

func TxToTransaction(tx Tx) (ctypes.Transaction, error){
	if tx["gas"] == "" {
		return ctypes.Transaction{}, errors.New("gas not specified")
	}
	if tx["gasPrice"] == "" {
		return ctypes.Transaction{}, errors.New("gasPrice not specified")
	}
	if tx["nonce"] == "" {
		return ctypes.Transaction{}, errors.New("nonce not specified")
	}

	nonce, _ := strconv.ParseInt(tx["nonce"], 0, 64)
	from := common.HexToAddress(tx["from"])
	to := common.HexToAddress(tx["to"])
	gasPrice, _ := strconv.ParseInt(tx["gasPrice"], 0, 64)
	//gas, _ := strconv.ParseInt(tx["gas"], 0, 64)
	value, _ := strconv.ParseInt(tx["value"], 0, 64)
	data := common.Hex2Bytes(tx["payload"])
	gasLimit, _ := strconv.ParseInt(tx["gasLimit"], 0, 64)

	transaction := ctypes.Transaction{
		Data:ctypes.TxData{
			From: &from,
			Recipient: &to,
			AccountNonce: uint64(nonce),
			Amount: big.NewInt(value),
			GasLimit: uint64(gasLimit),
			Price: big.NewInt(gasPrice),
			Payload: data,
		},
	}
	return transaction, nil
}