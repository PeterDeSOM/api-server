package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	s "github.com/crypblorm/bitcoin/api-server/services"
	"github.com/go-chi/render"
)

type apiDecodePsbt struct {
	Psbt string `json:"psbt"`
}

// DecodePsbt decodes psbt to network code
func DecodePsbt(w http.ResponseWriter, r *http.Request) {

	// decode api data

	var api apiDecodePsbt

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&api)
	if err != nil {
		log.Fatal("[ERROR] on decoding api in the CancelPsbt:", err)
	}

	d := s.DecodePsbt(api.Psbt)

	render.JSON(w, r, d)
}
