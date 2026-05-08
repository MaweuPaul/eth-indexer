package api

import (
	"encoding/json"
	"net/http"

	"github.com/MaweuPaul/eth-indexer/internal/store"
)

type API struct {
	store *store.Store
}

func New(store *store.Store) *API {
	return &API{store: store}
}

func (a *API) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events/", a.getEventsByContract)
	mux.HandleFunc("/health", a.health)
	return http.ListenAndServe(":"+port, mux)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *API) getEventsByContract(w http.ResponseWriter, r *http.Request) {
	contract := r.URL.Path[len("/events/"):]
	if contract == "" {
		http.Error(w, "contract address required", http.StatusBadRequest)
		return
	}

	events, err := a.store.GetEventsByContract(contract)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
