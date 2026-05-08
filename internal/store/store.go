package store

import (
    "database/sql"
    "encoding/json"
    "fmt"
    _ "github.com/lib/pq"
)

type Store struct {
    db *sql.DB
}

type Event struct {
    ID          int
    Contract    string
    EventName   string
    BlockNumber uint64
    TxHash      string
    LogIndex    uint
    Data        map[string]interface{}
}

func New(databaseURL string) (*Store, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    return &Store{db: db}, nil
}

func (s *Store) SaveEvent(event *Event) error {
    data, err := json.Marshal(event.Data)
    if err != nil {
        return fmt.Errorf("failed to marshal event data: %w", err)
    }

    _, err = s.db.Exec(`
        INSERT INTO events (contract, event_name, block_number, tx_hash, log_index, data)
        VALUES ($1, $2, $3, $4, $5, $6)`,
        event.Contract,
        event.EventName,
        event.BlockNumber,
        event.TxHash,
        event.LogIndex,
        data,
    )
    return err
}

func (s *Store) GetEventsByContract(contract string) ([]*Event, error) {
    rows, err := s.db.Query(`
        SELECT id, contract, event_name, block_number, tx_hash, log_index, data
        FROM events WHERE contract = $1
        ORDER BY block_number DESC`, contract)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []*Event
    for rows.Next() {
        e := &Event{}
        var data []byte
        err := rows.Scan(&e.ID, &e.Contract, &e.EventName, &e.BlockNumber, &e.TxHash, &e.LogIndex, &data)
        if err != nil {
            return nil, err
        }
        json.Unmarshal(data, &e.Data)
        events = append(events, e)
    }
    return events, nil
}