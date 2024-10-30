package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/StevenSopilidis/kvs/core"
	"github.com/gorilla/mux"
)

// interface defining basic port to communicate with core bussiness logic
type Frontend interface {
	Start(kv *core.KeyValueStore) error
}

// returns frontend based on the type specified
func NewFrontend(frontend string) (Frontend, error) {
	switch frontend {
	case "rest":
		return &restFrontEnd{}, nil
	case "":
		return nil, fmt.Errorf("frontend type not specified")
	default:
		return nil, fmt.Errorf("not such type of frontend exists")
	}
}

// represents a REST API frontend port
type restFrontEnd struct {
	kv *core.KeyValueStore
}

type restFrontendGetResponse struct {
	Value string
}

func (f *restFrontEnd) keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := f.kv.Get(key)
	if err != nil {
		if errors.Is(err, &core.ErrNoSuckKey{}) {
			http.Error(w, err.Error(), http.StatusNotFound)
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := restFrontendGetResponse{
		Value: value,
	}

	buffer, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(buffer)
}

type restFrontendPutResponse struct {
	Data string
}

func (f *restFrontEnd) keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decodedBody restFrontendPutResponse
	json.Unmarshal(body, &decodedBody)

	if decodedBody.Data == "" {
		http.Error(w, "invalid data sent", http.StatusBadRequest)
		return
	}

	err = f.kv.Put(key, decodedBody.Data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (f *restFrontEnd) keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]

	err := f.kv.Delete(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (f *restFrontEnd) Start(kv *core.KeyValueStore) error {
	f.kv = kv

	r := mux.NewRouter()
	r.HandleFunc("/v1/{key}", f.keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", f.keyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", f.keyValueDeleteHandler).Methods("DELETE")

	addr := "127.0.0.1:8080"

	log.Printf("Server starting listening at at address: %s", addr)
	return http.ListenAndServeTLS(addr, "../cert.pem", "../key.pem", r)
}
