package indexer

import (
	"context"
	"log"

	"github.com/MaweuPaul/eth-indexer/internal/decoder"
	"github.com/MaweuPaul/eth-indexer/internal/hub"
	"github.com/MaweuPaul/eth-indexer/internal/store"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Indexer struct {
	client    *ethclient.Client
	store     *store.Store
	decoder   *decoder.Decoder
	hub       *hub.Hub
	contracts []common.Address
}

func New(alchemyURL string, store *store.Store, h *hub.Hub, contracts []string) (*Indexer, error) {
	client, err := ethclient.Dial(alchemyURL)
	if err != nil {
		return nil, err
	}

	dec, err := decoder.New()
	if err != nil {
		return nil, err
	}

	var addresses []common.Address
	for _, c := range contracts {
		addresses = append(addresses, common.HexToAddress(c))
	}

	return &Indexer{client: client, store: store, decoder: dec, hub: h, contracts: addresses}, nil
}

func (i *Indexer) Start(ctx context.Context) error {
	query := ethereum.FilterQuery{
		Addresses: i.contracts,
	}

	logs := make(chan types.Log)
	sub, err := i.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return err
	}

	log.Println("Listening for on-chain events...")

	for {
		select {
		case err := <-sub.Err():
			return err
		case vLog := <-logs:
			i.handleLog(vLog)
		}
	}
}

func (i *Indexer) handleLog(vLog types.Log) {
	decoded, err := i.decoder.Decode(vLog)
	if err != nil {
		log.Println("Failed to decode log:", err)
		return
	}

	event := &store.Event{
		Contract:    vLog.Address.Hex(),
		EventName:   decoded.Name,
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash.Hex(),
		LogIndex:    vLog.Index,
		Data:        decoded.ToMap(),
	}

	if err := i.store.SaveEvent(event); err != nil {
		log.Println("Failed to save event:", err)
		return
	}

	// broadcast to connected websocket clients
	i.hub.Broadcast(&hub.Event{
		Type: "new_event",
		Data: event,
	})

	log.Printf("%s | block %d | from %s", decoded.Name, vLog.BlockNumber, decoded.Fields["from"])
}
