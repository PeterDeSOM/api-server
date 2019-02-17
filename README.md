# Som32 Conveinent PSBT Service

Partially Signed Bitcoin Transaction API Service developed in Golang, and its data is stored in LevelDB. This is not a fully developed service as well.

## Key libraries

```go
import (
    "github.com/btcsuite/btcd/btcjson"
    "github.com/go-chi/chi"
    "github.com/gorilla/websocket"
    "github.com/syndtr/goleveldb/leveldb"
)
```

## Library souce files modified

To fully support Partially Signed Bitcoin Transaction functionalities, decoding encoded code is necessary, but the existing library didn't provide it. Hence the functions that provide decoding transaction code have been added to the existing library source file as below. Also, it is marked with initials 'PETERKIM'.

### Source code modified and updated 

```go
    // github.com/btcsuite/rpcserver.go line 172
    // SOM32 PETERKIM
    // Partially Signed Transaction
    "decodepsbt": handleDecodePsbt,

    // github.com/btcsuite/rpcserver.go line 287   
    // SOM32 PETERKIM
	// Partially Signed Transaction
    "decoderpsbt": {},
    
    // github.com/btcsuite/rpcserverhelp.go line 672   
	// SOM32 PETERKIM.
	// Partially Signed Transaction
	// DecodePsbtCmd help.
	"decodepsbt--synopsis":  "Returns a JSON object representing the provided serialized, base64-encoded psbt.",
	"decodepsbt-base64psbt": "Serialized, base64-encoded psbt",

    // github.com/btcsuite/rpcserverhelp.go line 744   
	// SOM32 PETERKIM.
	// Partially Signed Transaction
    "decodepsbt": {(*btcjson.DecodePsbtResult)(nil)},
```

### Files added 

```go
    // github.com/btcsuite/btcd/btcjson/chainsvrcmds_psbt.go
    // Copyright (c) 2019 The Crypblorm Som32
    // Use of this source code is governed by an ISC
    // license that can be found in the LICENSE file.

    // NOTE: This file is intended to house the RPC commands that are supported by
    // a chain server.

    package btcjson

    // DecodePsbtCmd defines the DecodePsbt JSON-RPC command.
    type DecodePsbtCmd struct {
        Base64Psbt string
    }

    // NewDecodePsbtCmd returns a new instance which can be used to issue
    // a DecodePsbt JSON-RPC command.
    func NewDecodePsbtCmd(psbt string) *DecodePsbtCmd {
        return &DecodePsbtCmd{
            Base64Psbt: psbt,
        }
    }
```

```go
    // github.com/btcsuite/btcd/btcjson/chainsvrresults_psbt.go
    // Copyright (c) 2019 The Crypblorm Som32
    // Use of this source code is governed by an ISC
    // license that can be found in the LICENSE file.

    package btcjson

    // DecodePsbtResult models the data from the decodepsbt command.
    type DecodePsbtResult struct {
        Tx      Psbtx       `json:"tx"`
        Unknown interface{} `json:"unknown"`
        Inputs  []Input     `json:"inputs"`
        Outputs []Output    `json:"outputs"`
        Fee     float64     `json:"fee"`
    }

    // Psbtx models the transaction to be pushed to network.
    type Psbtx struct {
        Txid     string `json:"txid"`
        Hash     string `json:"hash,omitempty"`
        Version  int32  `json:"version"`
        Size     int32  `json:"size,omitempty"`
        Vsize    int32  `json:"vsize,omitempty"`
        Weight   int32  `json:"weight,omitempty"`
        LockTime uint32 `json:"locktime"`
        Vin      []Vin  `json:"vin"`
        Vout     []Vout `json:"vout"`
    }

    // Input models the transaction to be pushed to network.
    type Input struct {
        WitnessUtxo   WitnessUtxo  `json:"witness_utxo,omitempty"`
        Signatures    *interface{} `json:"partial_signatures,omitempty"`
        Sighash       string       `json:"sighash,omitempty"`
        RedeemScript  ScriptBase   `json:"redeem_script,omitempty"`
        WitnessScript ScriptBase   `json:"witness_script,omitempty"`
    }

    // ScriptPubKey models the scriptPubKey data of a tx script.  It is
    // defined separately since it is used by multiple commands.
    type ScriptPubKey struct {
        Asm     string `json:"asm"`
        Hex     string `json:"hex"`
        Type    string `json:"type"`
        Address string `json:"address"`
    }

    // WitnessUtxo models the transaction to be pushed to network.
    type WitnessUtxo struct {
        Amount       float64      `json:"amount"`
        ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
    }

    // ScriptBase models the transaction to be pushed to network.
    type ScriptBase struct {
        Asm  string `json:"asm,omitempty"`
        Hex  string `json:"hex,omitempty"`
        Type string `json:"type,omitempty"`
    }

    // Output models the transaction to be pushed to network.
    type Output struct {
        RedeemScript  *ScriptBase `json:"redeem_script,omitempty"`
        WitnessScript *ScriptBase `json:"witness_script,omitempty"`
    }
```

and 

```go
    //github.com/btcsuite/btcd/btcclient/rawtransactions_psbt.go
    // Copyright (c) 2019 The Crypblorm Som32
    // Use of this source code is governed by an ISC
    // license that can be found in the LICENSE file.

    package rpcclient

    import (
        "encoding/json"

        "github.com/btcsuite/btcd/btcjson"
    )

    // FutureDecodePsbtResult is a future promise to deliver the result
    // of a DecodePsbtAsync RPC invocation (or an applicable error).
    type FutureDecodePsbtResult chan *response

    // Receive waits for the response promised by the future and returns information
    // about a transaction given its serialized bytes.
    func (r FutureDecodePsbtResult) Receive() (*btcjson.DecodePsbtResult, error) {
        res, err := receiveFuture(r)
        if err != nil {
            return nil, err
        }

        var i interface{}
        err = json.Unmarshal(res, &i)
        if err != nil {
            return nil, err
        }

        // Unmarshal result as a DecodePsbt result object.
        var decodedPsbtResult btcjson.DecodePsbtResult
        err = json.Unmarshal(res, &decodedPsbtResult)
        if err != nil {
            return nil, err
        }

        return &decodedPsbtResult, nil
    }

    // DecodePsbtAsync returns an instance of a type that can be used to
    // get the result of the RPC at some future time by invoking the Receive
    // function on the returned instance.
    //
    // See DecodePsbt for the blocking version and more details.
    func (c *Client) DecodePsbtAsync(psbt string) FutureDecodePsbtResult {
        // base64Psbt := base64.StdEncoding.EncodeToString(serializedPsbt)
        cmd := btcjson.NewDecodePsbtCmd(psbt)
        return c.sendCmd(cmd)
    }

    // DecodePsbt returns information about a psbt given itsserialized bytes.
    func (c *Client) DecodePsbt(psbt string) (*btcjson.DecodePsbtResult, error) {
        return c.DecodePsbtAsync(psbt).Receive()
    }
```
