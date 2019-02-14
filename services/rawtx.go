package services

import (
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

// GetRawTransaction returns information about a psbt given itsserialized bytes.
func GetRawTransaction(txid *chainhash.Hash) *btcutil.Tx {

	client := HTTPClient()
	tx, err := client.GetRawTransaction(txid)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}
	client.Shutdown()
	return tx
}

// DecodePsbt returns information about a psbt given itsserialized bytes.
// func DecodePsbt(serializedPsbt []byte) *btcutil.Tx {
func DecodePsbt(psbt string) *btcjson.DecodePsbtResult {

	client := HTTPClient()
	d, err := client.DecodePsbt(psbt)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}
	client.Shutdown()

	return d
}
