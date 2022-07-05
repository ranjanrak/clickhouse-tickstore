package tickstore

import (
	"fmt"
	"log"
	"time"

	kitemodels "github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
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
func (c *Client) onConnect() {
	fmt.Println("Connected")
	err := c.ticker.Subscribe(c.tokenList)
	if err != nil {
		fmt.Println("err: ", err)
	}
	// Set subscription mode for given list of tokens
	err = c.ticker.SetMode(kiteticker.ModeFull, c.tokenList)
	if err != nil {
		fmt.Println("err: ", err)
	}
}

// Triggered when tick is received
func (c *Client) onTick(tick kitemodels.Tick) {
	// Send {instrument token, timestamp, lastprice} struct to channel
	c.pipeline <- tickData{tick.InstrumentToken, tick.Timestamp.Time, tick.LastPrice}
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
func (c *Client) createBulkDump() {
	s := make([]tickData, 0)
	for i := range c.pipeline {
		// create array of ticks to do bulk insert
		s = append(s, i)
		if len(s) > c.dumpSize {
			// Send message array for the bulk dump
			err := c.InsertDB(s)
			if err != nil {
				log.Fatalf("Error inserting tick to DB: %v", err)
			}
			// Remove all the element from the array that is dumped to DB
			s = nil
		}
	}
}

// Insert tick data to clickhouse periodically
func (c *Client) InsertDB(tickArray []tickData) error {
	tx, err := c.dbClient.Begin()
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
	c.ticker.SetReconnectMaxRetries(5)
	// Assign callbacks
	c.ticker.OnError(onError)
	c.ticker.OnClose(onClose)
	c.ticker.OnConnect(c.onConnect)
	c.ticker.OnReconnect(onReconnect)
	c.ticker.OnNoReconnect(onNoReconnect)
	c.ticker.OnTick(c.onTick)

	// Go-routine that listens to pipeline channel forever
	// And performs periodic bulk insert based on user-input dumpSize
	go c.createBulkDump()

	// Start the connection
	c.ticker.Serve()
}
