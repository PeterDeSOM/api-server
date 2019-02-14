// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package services

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/joho/godotenv"
)

// HTTPClient uses the rpcclient package to connect to a Bitcoin Core RPC server
// using HTTP POST mode.
func HTTPClient() *rpcclient.Client {

	e := godotenv.Load() //Load .env file
	if e != nil {
		fmt.Print(e)
	}

	_host := os.Getenv("Host")
	_user := os.Getenv("User")
	_pass := os.Getenv("Pass")
	_httpPostMode, _ := strconv.ParseBool(os.Getenv("HTTPPostMode"))
	_disableTLS, _ := strconv.ParseBool(os.Getenv("DisableTLS"))

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         _host,
		User:         _user,
		Pass:         _pass,
		HTTPPostMode: _httpPostMode, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   _disableTLS,   // Bitcoin core does not provide TLS by default
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

// WSClient uses the rpcclient package to connect to a Bitcoin Core RPC server
// using websockets.
func WSClient() *rpcclient.Client {
	e := godotenv.Load() //Load .env file
	if e != nil {
		fmt.Print(e)
	}

	// Only override the handlers for notifications you care about.
	// Also note most of the handlers will only be called if you register
	// for notifications.  See the documentation of the rpcclient
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := rpcclient.NotificationHandlers{
		OnAccountBalance: func(account string, balance btcutil.Amount, confirmed bool) {
			log.Printf("New balance for account %s: %v", account,
				balance)
		},
	}

	// Connect to local btcwallet RPC server using websockets.
	certHomeDir := btcutil.AppDataDir("btcwallet", false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18332",
		Endpoint:     "ws",
		User:         "yourrpcuser",
		Pass:         "yourrpcpass",
		Certificates: certs,
	}
	client, err := rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
