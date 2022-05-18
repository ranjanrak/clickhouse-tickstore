package tickstore

import (
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Setup mockclient
func setupMock(mockRow *sqlmock.Rows, query string) *Client {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectQuery(query).
		WillReturnRows(mockRow)

	cli := &Client{
		dbClient:    db,
		apiKey:      "your_api_key",
		accessToken: "your_access_token",
		tokenList:   []uint32{633601, 895745},
		dumpSize:    1000,
	}
	return cli
}

func TestFetchCandle(t *testing.T) {
	// Timestamp in time.Time object
	candleTime1 := time.Date(2022, 5, 18, 14, 04, 0, 0, time.Local)
	candleTime2 := time.Date(2022, 5, 18, 14, 05, 0, 0, time.Local)
	// Add mock row for test
	mockedRow := sqlmock.NewRows([]string{"instrument_token", "time_minute", "open", "high", "low", "close"}).
		AddRow(633601, candleTime1, 156.85, 158, 156, 157.75).
		AddRow(633601, candleTime2, 157.80, 158.75, 156.35, 156.90)

	// Add expected query
	query := `SELECT instrument_token, time_minute, groupArray(price)[1] AS open, max(price) AS high, min(price) AS low, groupArray(price)[-1] AS close FROM ( 
		SELECT instrument_token, toStartOfMinute(timestamp) AS time_minute, price FROM tickdata WHERE (instrument_token = 633601) AND 
		(timestamp >= '2022-05-18 14:04:00') AND (timestamp <= '2022-05-18 14:04:59') )
		GROUP BY (instrument_token, time_minute) ORDER BY time_minute ASC`

	dbMock := setupMock(mockedRow, query)

	timeStart := time.Date(2022, 5, 18, 14, 04, 0, 0, time.Local)
	timeEnd := time.Date(2022, 5, 18, 14, 04, 59, 0, time.Local)
	candles, err := dbMock.FetchCandle(633601, timeStart, timeEnd)
	if err != nil {
		log.Fatalf("Error fetching candle: %v", err)
	}

	// Expected output
	expectedCandles := Candles{
		CandleData{
			InstrumentToken: 633601,
			TimeStamp:       candleTime1,
			Open:            156.85,
			High:            158,
			Low:             156,
			Close:           157.75,
		},
		CandleData{
			InstrumentToken: 633601,
			TimeStamp:       candleTime2,
			Open:            157.8,
			High:            158.75,
			Low:             156.35,
			Close:           156.9,
		},
	}
	assert.Equal(t, expectedCandles, candles, "Actual candle data not matching with expectedCandles response")
}
