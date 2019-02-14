package controllers

// TO DO: change codes to use leveldb.batch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	s "github.com/crypblorm/bitcoin/api-server/services"
	d "github.com/crypblorm/bitcoin/api-server/services/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type apiCreatePsbt struct {
	Name      string   `json:"name"`
	Multisig  string   `json:"multisig"`
	M         uint8    `json:"m"`
	Recipient string   `json:"recipient"`
	By        string   `json:"by"`
	Amount    uint64   `json:"amount"`
	Fee       uint64   `json:"fee"`
	Psbt      string   `json:"psbt"`
	PsbtBn    string   `json:"psbtbn"`
	Desc      string   `json:"desc"`
	Datetime  int64    `json:"datetime"`
	Signed    []string `json:"signed"`
}
type apiSignPsbt struct {
	Multisig string `json:"multisig"`
	Name     string `json:"name"`
	ID       int64  `json:"id"`
	By       string `json:"by"`
	Psbt     string `json:"psbt"`
	PsbtBn   string `json:"psbtbn"`
	Datetime int64  `json:"datetime"`
}
type apiCancel struct {
	Multisig string `json:"multisig"`
	Location uint8  `json:"location"`
	ID       int64  `json:"id"`
	By       string `json:"by"`
}
type apiGetComplete struct {
	Multisig string `json:"multisig"`
	ID       int64  `json:"id"`
}
type apiPushPsbt struct {
	Multisig string `json:"multisig"`
	ID       int64  `json:"id"`
	By       string `json:"by"`
}
type dataSigned struct {
	Datetime int64 `json:"datetime"`
	Complete bool  `json:"complete"`
	// Created  dataCreated  `json:"created"`
	Signers []dataSigner `json:"signers"`
	By      string       `json:"by"`
}
type dataSigner struct {
	Datetime int64  `json:"datetime"`
	Signer   string `json:"signer"`
	Psbt     string `json:"psbt"`
}
type pState struct {
	Created  uint8  `json:"created"`
	Signed   uint8  `json:"cigning"`
	Complete uint8  `json:"complete"`
	Pushed   uint32 `json:"pushed"`
	Canceled uint32 `json:"canceled"`
}
type dataHistory struct {
	Datetime int64  `json:"datetime"`
	Action   string `json:"action"`
	For      string `json:"for"`
	By       string `json:"by"`
}
type renderForm struct {
	ID         int64  `json:"id"`
	State      bool   `json:"state"`
	Incomplete bool   `json:"incomplete"`
	Code       int8   `json:"code"`
	Msg        string `json:"msg"`
}
type renderList struct {
	Success bool     `json:"success"`
	Data    []string `json:"data"`
}
type renderFormSimple struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}
type renderPsbtComplete struct {
	Success bool          `json:"success"`
	Created apiCreatePsbt `json:"created"`
	Signed  []apiSignPsbt `json:"signed"`
}

func decodePsbt(psbt string) *btcjson.DecodePsbtResult {
	client := s.HTTPClient()
	decoded, err := client.DecodePsbt(psbt)
	if err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}
	client.Shutdown()
	return decoded
}

func isValidRequest(auth string, psbtBn string) bool {
	psbtx := decodePsbt(psbtBn)
	fmt.Printf("apiCreatePsbt: %s\n", psbtx.Tx.Hash)
	return true
}

func getPsbtStats(db *leveldb.DB, key string, ps *pState) {
	val, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Fatal("# [ERROR] on getting state key: ", err)
	}
	if err := json.Unmarshal([]byte(val), &ps); err != nil {
		log.Fatal("# [ERROR] on val json.Unmarshal:", err)
	}
}

func getCreatedPsbt(db *leveldb.DB, key string, cp *apiCreatePsbt) {
	val, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Fatal("# [ERROR] on getting getCreatedPsbt data: ", err)
	}
	if err := json.Unmarshal([]byte(val), &cp); err != nil {
		log.Fatal("# [ERROR] on val json.Unmarshal in getCreatedPsbt:", err)
	}
}

func getSignedPsbt(db *leveldb.DB, key string, sp *apiSignPsbt) {
	val, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Fatal("# [ERROR] on getting state key in getSignedPsbt: ", err)
	}
	if err := json.Unmarshal([]byte(val), &sp); err != nil {
		log.Fatal("# [ERROR] on val json.Unmarshal in getSignedPsbt:", err)
	}
}

// GetPsbtComplete returns complete signed psbt
func GetPsbtComplete(w http.ResponseWriter, r *http.Request) {
	var _api apiGetComplete

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&_api)
	if err != nil {
		log.Fatal("[ERROR] on decoding api in the CancelPsbt:", err)
	}

	sid := strconv.FormatInt(_api.ID, 10)
	_db := d.OpenDB("/psbt/untrusty")
	defer _db.Close()

	var cp apiCreatePsbt
	getCreatedPsbt(_db, _api.Multisig+"CREATED"+sid, &cp)

	var list []apiSignPsbt
	_iter := _db.NewIterator(
		&util.Range{
			Start: []byte(_api.Multisig + "SIGNED" + "0"),
			Limit: []byte(_api.Multisig + "SIGNED" + "9"),
		},
		nil,
	)
	for _iter.Next() {
		var sp apiSignPsbt
		if err := json.Unmarshal([]byte(_iter.Value()), &sp); err != nil {
			log.Fatal("# [ERROR] on json.Unmarshal SIGNED key:", err)
		}
		if sp.ID == _api.ID {
			list = append(list, sp)
		}
	}
	_iter.Release()

	var res renderPsbtComplete
	res.Success = true
	res.Created = cp
	res.Signed = list

	render.JSON(w, r, res)
}

// PsbtList returns a psbt list.
func PsbtList(w http.ResponseWriter, r *http.Request) {

	_address := chi.URLParam(r, "address")
	_target := chi.URLParam(r, "target")
	_by := chi.URLParam(r, "by")
	_key := ""
	if _target == "COMPLETE" {
		_key = _address + "CREATED"
	} else {
		_key = _address + _target
	}

	_db := d.OpenDB("/psbt/untrusty")
	defer _db.Close()

	_iter := _db.NewIterator(
		&util.Range{
			Start: []byte(_key + "0"),
			Limit: []byte(_key + "9"),
		},
		nil,
	)

	var verified bool

	switch _target {
	case "CREATED":
		var list []apiCreatePsbt

		i := 0
		for _iter.Next() {
			var cp apiCreatePsbt
			if err := json.Unmarshal([]byte(_iter.Value()), &cp); err != nil {
				log.Fatal("# [ERROR] on json.Unmarshal CREATED key:", err)
			}

			if i == 0 {
				if verified = isValidRequest("1234", cp.PsbtBn); !verified {
					var res renderFormSimple
					res.Success = false
					res.Msg = "Permission denied"
					render.JSON(w, r, res)
					return
				}
			}

			if _by == "ALL" {
				list = append(list, cp)
			} else if cp.By == _by {
				list = append(list, cp)
			}

			i++
		}
		_iter.Release()
		render.JSON(w, r, list)
		break

	case "SIGNED":
		var list []apiSignPsbt

		for _iter.Next() {
			var sp apiSignPsbt
			if err := json.Unmarshal([]byte(_iter.Value()), &sp); err != nil {
				log.Fatal("# [ERROR] on json.Unmarshal SIGNED key:", err)
			}
			if _by == "ALL" {
				list = append(list, sp)
			} else if sp.By == _by {
				list = append(list, sp)
			}
		}
		_iter.Release()
		render.JSON(w, r, list)
		break

	case "COMPLETE":
		var list []apiCreatePsbt

		for _iter.Next() {
			var cp apiCreatePsbt
			if err := json.Unmarshal([]byte(_iter.Value()), &cp); err != nil {
				log.Fatal("# [ERROR] on json.Unmarshal CREATED key:", err)
			}
			if cp.M > 0 && int(cp.M) == len(cp.Signed) {
				list = append(list, cp)
			}
		}
		_iter.Release()
		render.JSON(w, r, list)
		break

	case "HISTORY":
		var list []dataHistory

		for _iter.Next() {
			var dh dataHistory
			if err := json.Unmarshal([]byte(_iter.Value()), &dh); err != nil {
				log.Fatal("# [ERROR] on json.Unmarshal HISTORY key:", err)
			}
			list = append(list, dh)
		}
		_iter.Release()
		render.JSON(w, r, list)
		break
	}
}

// CreatePsbt creates a psbt.
func CreatePsbt(w http.ResponseWriter, r *http.Request) {

	// decode api data

	var api apiCreatePsbt

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&api)
	if err != nil {
		log.Fatal("[ERROR] CreatePsbt decoder.Decode(&api):", err)
	}

	// open database

	db := d.OpenDB("/psbt/untrusty")
	defer db.Close()

	// declare batch

	batch := new(leveldb.Batch)

	// update created state

	var ps pState
	key := api.Multisig + "STATE"

	if !d.Exist(db, key) {
		pstate := `{ "created":0, "Signed":0, "complete":0, "pushed":0, "canceled":0 }`
		batch.Put([]byte(key), []byte(pstate))

		if err := json.Unmarshal([]byte(pstate), &ps); err != nil {
			log.Fatal("# [ERROR] on json.Unmarshal state key:", err)
		}
	} else {
		getPsbtStats(db, key, &ps)
	}

	ps.Created++
	pstate, err := json.Marshal(ps)
	batch.Put([]byte(key), pstate)

	// put new create psbt

	api.Datetime = time.Now().UnixNano() / 1e6
	sid := strconv.FormatInt(api.Datetime, 10)
	key = api.Multisig + "CREATED" + sid

	// fmt.Printf("key: %s\n", key)
	// fmt.Printf("apiCreatePsbt: %+v\n", api)

	bapi, err := json.Marshal(api)
	if err != nil {
		log.Fatal("# [ERROR] on json.Marshal to convert api to byte, bapi :", err)
	}
	batch.Put([]byte(key), bapi)

	// put history

	var hist dataHistory
	key = api.Multisig + "HISTORY" + sid
	hist.Datetime = api.Datetime
	hist.Action = "Create"
	hist.For = api.Name
	hist.By = api.By

	bhist, err := json.Marshal(hist)
	if err != nil {
		log.Fatal("# [ERROR] on json.Marshal pcreateds:", err)
	}
	batch.Put([]byte(key), bhist)

	// write batch all

	err = db.Write(batch, nil)
	if err != nil {
		log.Fatal("# [ERROR] on batch all in createPsbt:", err)
	}

	res := renderForm{
		ID:         api.Datetime,
		State:      true,
		Incomplete: true,
		Code:       100,
		Msg:        "",
	}
	render.JSON(w, r, res)
}

// SignPsbt signs psbt from Created or another Signed
func SignPsbt(w http.ResponseWriter, r *http.Request) {

	// open batch
	_batch := new(leveldb.Batch)

	var _api apiSignPsbt
	_api.Datetime = time.Now().UnixNano() / 1e6

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&_api)
	if err != nil {
		log.Fatal("[ERROR] CreatePsbt decoder.Decode(&api):", err)
	}

	// open database
	_db := d.OpenDB("/psbt/untrusty")
	defer _db.Close()

	// update state
	var _ps pState
	getPsbtStats(_db, _api.Multisig+"STATE", &_ps)
	_ps.Signed++
	bps, _ := json.Marshal(_ps)
	_batch.Put([]byte(_api.Multisig+"STATE"), bps)

	// update created and completed psbt
	sid := strconv.FormatInt(_api.ID, 10)

	var _cp apiCreatePsbt
	getCreatedPsbt(_db, _api.Multisig+"CREATED"+sid, &_cp)
	_cp.Signed = append(_cp.Signed, _api.By)
	bcp, _ := json.Marshal(_cp)
	_batch.Put([]byte(_api.Multisig+"CREATED"+sid), bcp)

	if int(_cp.M) == len(_cp.Signed) {
		bcp, _ := json.Marshal(_cp)
		_batch.Put([]byte(_api.Multisig+"COMPLETE"+sid), bcp)
	}

	// insert signed data
	newsid := strconv.FormatInt(_api.Datetime, 10)
	_api.Name = _cp.Name
	bapi, _ := json.Marshal(_api)
	_batch.Put([]byte(_api.Multisig+"SIGNED"+newsid), bapi)

	// put history
	var _hist dataHistory
	_hist.Datetime = _api.Datetime
	_hist.Action = "Sign"
	_hist.For = _cp.Name
	_hist.By = _api.By
	bhist, _ := json.Marshal(_hist)
	_batch.Put([]byte(_api.Multisig+"HISTORY"+newsid), bhist)

	err = _db.Write(_batch, nil)
	if err != nil {
		log.Fatal("# [ERROR] on batch all in signPsbt:", err)
	}

	res := renderForm{
		State:      true,
		Incomplete: true,
		Code:       100,
		Msg:        "",
	}
	render.JSON(w, r, res)
}

// PushPsbt finalizes and pushes psbt to network
func PushPsbt(w http.ResponseWriter, r *http.Request) {
	var _api apiPushPsbt

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&_api)
	if err != nil {
		log.Fatal("[ERROR] on decoding api in the CancelPsbt:", err)
	}

	sid := strconv.FormatInt(_api.ID, 10)
	_batch := new(leveldb.Batch)
	_db := d.OpenDB("/psbt/untrusty")
	defer _db.Close()

	var _ps pState
	_keyState := _api.Multisig + "STATE"
	getPsbtStats(_db, _keyState, &_ps)

	var cp apiCreatePsbt
	getCreatedPsbt(_db, _api.Multisig+"CREATED"+sid, &cp)
	_batch.Delete([]byte(_api.Multisig + "CREATED" + sid))
	_ps.Created--

	_iter := _db.NewIterator(
		&util.Range{
			Start: []byte(_api.Multisig + "SIGNED" + "0"),
			Limit: []byte(_api.Multisig + "SIGNED" + "9"),
		},
		nil,
	)
	for _iter.Next() {
		var sp apiSignPsbt
		if err := json.Unmarshal([]byte(_iter.Value()), &sp); err != nil {
			log.Fatal("# [ERROR] on json.Unmarshal SIGNED key:", err)
		}
		if sp.ID == _api.ID {
			var spid = strconv.FormatInt(sp.Datetime, 10)
			_batch.Delete([]byte(_api.Multisig + "SIGNED" + spid))
			_ps.Signed--
		}
	}
	_iter.Release()
	_ps.Pushed++

	// put history
	var _hist dataHistory
	_hist.Datetime = time.Now().UnixNano() / 1e6
	_hist.By = _api.By
	_hist.Action = "Push & Pay"
	_hist.For = cp.Name

	bhist, _ := json.Marshal(_hist)
	bps, _ := json.Marshal(_ps)

	_batch.Put([]byte(_api.Multisig+"HISTORY"+strconv.FormatInt(_hist.Datetime, 10)), bhist)
	_batch.Put([]byte(_keyState), bps)
	_db.Write(_batch, nil)

	var res renderFormSimple
	res.Success = true
	res.Msg = ""

	render.JSON(w, r, res)
}

// CancelPsbt cancels a psbt.
func CancelPsbt(w http.ResponseWriter, r *http.Request) {

	// decode api data

	var _api apiCancel

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&_api)
	if err != nil {
		log.Fatal("[ERROR] on decoding api in the CancelPsbt:", err)
	}

	_sid := strconv.FormatInt(_api.ID, 10)
	_keyLoc := _api.Multisig
	_batch := new(leveldb.Batch)
	_db := d.OpenDB("/psbt/untrusty")
	defer _db.Close()

	var _ps pState
	_keyState := _api.Multisig + "STATE"
	getPsbtStats(_db, _keyState, &_ps)

	// put history
	var _hist dataHistory
	_hist.Datetime = time.Now().UnixNano() / 1e6
	_hist.By = _api.By

	_found := false
	switch _api.Location {
	case 0:
		_keyLoc = _keyLoc + "CREATED" + _sid
		if _found = d.Exist(_db, _keyLoc); _found {
			var cp apiCreatePsbt
			getCreatedPsbt(_db, _keyLoc, &cp)

			_batch.Delete([]byte(_keyLoc))
			_ps.Created--

			_hist.Action = "Cancel Created"
			_hist.For = cp.Name
		}
		break
	case 1:
		_keyLoc = _keyLoc + "SIGNED" + _sid
		if _found = d.Exist(_db, _keyLoc); _found {
			var sp apiSignPsbt
			getSignedPsbt(_db, _keyLoc, &sp)
			refid := strconv.FormatInt(sp.ID, 10)

			var cp apiCreatePsbt
			getCreatedPsbt(_db, _api.Multisig+"CREATED"+refid, &cp)
			for i := range cp.Signed {
				if cp.Signed[i] == sp.By {
					cp.Signed = append(cp.Signed[:i], cp.Signed[i+1:]...)
					break
				}
			}
			bcp, _ := json.Marshal(cp)
			_batch.Put([]byte(_api.Multisig+"CREATED"+refid), bcp)

			_batch.Delete([]byte(_keyLoc))
			_ps.Signed--

			_hist.Action = "Cancel Signed"
			_hist.For = cp.Name
		}
		break
	case 2:
		_keyLoc = _keyLoc + "CREATED" + _sid

		if _found = d.Exist(_db, _keyLoc); _found {
			var cp apiCreatePsbt
			getCreatedPsbt(_db, _keyLoc, &cp)

			_hist.Action = "Cancel Complete"
			_hist.For = cp.Name

			_batch.Delete([]byte(_keyLoc))
			_ps.Created--

			_iter := _db.NewIterator(
				&util.Range{
					Start: []byte(_api.Multisig + "SIGNED0"),
					Limit: []byte(_api.Multisig + "SIGNED9"),
				},
				nil,
			)
			for _iter.Next() {
				var sp apiSignPsbt

				json.Unmarshal([]byte(_iter.Value()), &sp)
				if strconv.FormatInt(sp.ID, 10) == _sid {
					_batch.Delete(_iter.Key())
					_ps.Signed--
				}
			}
			_iter.Release()
		}
		break
	}

	var _res renderFormSimple
	if _found {
		msg := ""
		bhist, _ := json.Marshal(_hist)
		bps, _ := json.Marshal(_ps)

		_batch.Put([]byte(_api.Multisig+"HISTORY"+strconv.FormatInt(_hist.Datetime, 10)), bhist)
		_batch.Put([]byte(_keyState), bps)
		_db.Write(_batch, nil)

		_res = renderFormSimple{
			Success: true,
			Msg:     msg,
		}
	} else {
		_res = renderFormSimple{
			Success: false,
			Msg:     "Not found",
		}
	}
	render.JSON(w, r, _res)
}
