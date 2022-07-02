package tickstore

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	kitemodels "github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

// tickData is struct to store streaming tick data in clickhouse
type tickData struct {
	Token     uint32
	TimeStamp time.Time
	LastPrice float64
}

var (
	dbConnect   *sql.DB
	ticker      *kiteticker.Ticker
	wg          sync.WaitGroup
	isBulkReady sync.Mutex
	dumpSize    int
	tokens      []uint32
	pipeline    chan tickData
)

// Triggered when any error is raised
func onError(err error) {
	fmt.Println("Error: ", err)
}

// Triggered when websocket connection is closed
func onClose(code int, reason string) {
	fmt.Println("Close: ", code, reason)
}

// Triggered when connection is established and ready to send and accept data
func onConnect() {
	fmt.Println("Connected")
	err := ticker.Subscribe(tokens)
	if err != nil {
		fmt.Println("err: ", err)
	}
	// Set subscription mode for given list of tokens
	err = ticker.SetMode(kiteticker.ModeFull, tokens)
	if err != nil {
		fmt.Println("err: ", err)
	}
}

// Triggered when tick is received
func onTick(tick kitemodels.Tick) {
	// Send {instrument token, timestamp, lastprice} struct to channel
	pipeline <- tickData{tick.InstrumentToken, tick.Timestamp.Time, tick.LastPrice}
}

// Triggered when reconnection is attempted which is enabled by default
func onReconnect(attempt int, delay time.Duration) {
	fmt.Printf("Reconnect attempt %d in %fs\n", attempt, delay.Seconds())
}

// Triggered when maximum number of reconnect attempt is made and the program is terminated
func onNoReconnect(attempt int) {
	fmt.Printf("Maximum no of reconnect attempt reached: %d", attempt)
}

// Group all available channel messages and bulk insert to clickhouse
// Bulk insert is done at periodic interval depending on users input channel buffer size(dumpSize)
func createBulkDump() {
	s := make([]tickData, 0)
	for i := range pipeline {
		// create array of ticks to do bulk insert
		s = append(s, i)
		if len(s) > dumpSize {
			// Send message array for the bulk dump
			err := InsertDB(s)
			if err != nil {
				log.Fatalf("Error inserting tick to DB: %v", err)
			}
			// Remove all the element from the array that is dumped to DB
			s = nil
		}
	}
}

// Insert tick data to clickhouse periodically
func InsertDB(tickArray []tickData) error {
	tx, err := dbConnect.Begin()
	if err != nil {
		return err
	}

	sqlstmt := "INSERT INTO tickdata (instrument_token, timestamp, price) VALUES (?, ?, ?)"

	stmt, err := tx.Prepare(sqlstmt)
	if err != nil {
		return err
	}

	// Bulk write
	for _, tick := range tickArray {
		if _, err := stmt.Exec(
			tick.Token,
			tick.TimeStamp,
			tick.LastPrice,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Start ticker stream
func (c *Client) StartTicker() {

	dbConnect = c.dbClient

	dumpSize = c.dumpSize

	tokens = c.tokenList

	// Channel to store all upcoming streams of ticks
	pipeline = make(chan tickData, dumpSize)

	// Create new Kite ticker instance
	ticker = kiteticker.New(c.apiKey, c.accessToken)

	ticker.SetReconnectMaxRetries(5)

	// Assign callbacks
	ticker.OnError(onError)
	ticker.OnClose(onClose)
	ticker.OnConnect(onConnect)
	ticker.OnReconnect(onReconnect)
	ticker.OnNoReconnect(onNoReconnect)
	ticker.OnTick(onTick)

	// Go-routine that listens to pipeline channel forever
	// And performs periodic bulk insert based on user-input dumpSize
	go createBulkDump()

	// Start the connection
	ticker.Serve()
}
