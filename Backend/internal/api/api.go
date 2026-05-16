package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MaweuPaul/eth-indexer/internal/hub"
	"github.com/MaweuPaul/eth-indexer/internal/store"
	"github.com/gorilla/websocket"
)

type API struct {
	store *store.Store
	hub   *hub.Hub
}

func New(store *store.Store, hub *hub.Hub) *API {
	return &API{
		store: store,
		hub:   hub,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (a *API) Start(port string) error {

	mux := http.NewServeMux()
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dashboard.html")
	})
	mux.HandleFunc("/events/", a.getEventsByContract)
	mux.HandleFunc("/health", a.health)
	mux.HandleFunc("/ws", a.wsHandler)

	log.Println("API running on port", port)

	return http.ListenAndServe(":"+port, mux)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (a *API) wsHandler(w http.ResponseWriter, r *http.Request) {

	// upgrade HTTP -> websocket
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("websocket upgrade failed:", err)

		http.Error(
			w,
			"failed to upgrade websocket",
			http.StatusInternalServerError,
		)

		return
	}

	// register client
	a.hub.Register(conn)

	log.Println("client connected")

	// unregister on disconnect
	defer func() {
		a.hub.Unregister(conn)
		log.Println("client disconnected")
	}()

	// keep connection alive
	for {

		// waits for messages
		_, _, err := conn.ReadMessage()

		if err != nil {
			break
		}
	}
}

func (a *API) getEventsByContract(
	w http.ResponseWriter,
	r *http.Request,
) {

	contract := r.URL.Path[len("/events/"):]

	if contract == "" {
		http.Error(
			w,
			"contract address required",
			http.StatusBadRequest,
		)

		return
	}

	events, err := a.store.GetEventsByContract(contract)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(events)
}
