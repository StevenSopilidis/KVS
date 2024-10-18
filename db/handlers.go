package db

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/StevenSopilidis/kvs/persistance"
	"github.com/gorilla/mux"
)

type Server struct {
	store  *Store
	logger persistance.TransactionLogger
}

func NewServer(logger persistance.TransactionLogger) (*Server, error) {
	var err error

	store := NewStore()

	// read and replay all past events from log file
	events, errors := logger.ReadEvents()
	e, ok := persistance.Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.Type {
			case persistance.EventPut:
				store.Put(e.Key, e.Value)
			case persistance.EventDelete:
				store.Delete(e.Key)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	logger.Run()

	return &Server{
		store:  NewStore(),
		logger: logger,
	}, nil
}

type PutRequest struct {
	Data string
}

func (s *Server) KeyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decodedBody PutRequest
	json.Unmarshal(body, &decodedBody)

	if decodedBody.Data == "" {
		http.Error(w, "invalid data sent", http.StatusBadRequest)
		return
	}

	err = s.store.Put(key, decodedBody.Data)
	s.logger.WritePut(key, decodedBody.Data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type GetResponse struct {
	Value string
}

func (s *Server) KeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]

	value, err := s.store.Get(key)

	if errors.Is(err, &ErrNoSuckKey{}) {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	resBody := GetResponse{
		Value: value,
	}

	body, err := json.Marshal(resBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(body)
}

func (s *Server) KeyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]

	err := s.store.Delete(key)
	s.logger.WriteDelete(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
