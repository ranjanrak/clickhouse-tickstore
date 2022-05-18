package tickstore

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go"
)

// Client represents clickhouse DB client connection
type Client struct {
	dbClient    *sql.DB
	apiKey      string
	accessToken string
	tokenList   []uint32
	dumpSize    int
}

// ClientParam represents interface to connect clickhouse and kite ticker stream
type ClientParam struct {
	DBSource    string
	ApiKey      string
	AccessToken string
	TokenList   []uint32
	DumpSize    int
}

// Creates a new DB connection client
func New(userParam ClientParam) *Client {
	if userParam.DBSource == "" {
		userParam.DBSource = "tcp://127.0.0.1:9000?debug=true"
	}
	connect, err := sql.Open("clickhouse", userParam.DBSource)
	if err = connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
	}
	// Set default dump size to 5 times of the token list length
	if userParam.DumpSize == 0 {
		userParam.DumpSize = len(userParam.TokenList) * 5
	}
	// Create tickdata table for fresh instance
	// Replacingmergetree engine removes all duplicate entries with the same timestamp and price
	// As those won't be useful for candle creation
	_, err = connect.Exec(`
		CREATE TABLE IF NOT EXISTS tickdata (
			instrument_token   UInt32,
			timestamp          DateTime('Asia/Calcutta'),
			price              FLOAT()
		) engine=ReplacingMergeTree()
		ORDER BY (timestamp, instrument_token, price)
	`)
	if err != nil {
		log.Fatalf("Error creating tickdata table: %v", err)
	}

	return &Client{
		dbClient:    connect,
		apiKey:      userParam.ApiKey,
		accessToken: userParam.AccessToken,
		tokenList:   userParam.TokenList,
		dumpSize:    userParam.DumpSize,
	}
}
