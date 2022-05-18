package tickstore

import (
	"fmt"
	"time"
)

// CandleData represents OHLC candle data
type CandleData struct {
	InstrumentToken uint32
	TimeStamp       time.Time
	Open            float64
	High            float64
	Low             float64
	Close           float64
}

// Candles is an array of CandleData
type Candles []CandleData

// Creates OHLC candle from tickdata
func (c *Client) FetchCandle(instrumentToken int, startTime time.Time, endTime time.Time) (Candles, error) {
	startT := startTime.Format("2006-01-02 15:04:05")

	endT := endTime.Format("2006-01-02 15:04:05")

	// DB query to calculate OHLC between StartTime and EndTime for given instrument_token based on tickdata
	candleQueryStmt := fmt.Sprintf(`SELECT
			instrument_token,
			time_minute,
			groupArray(price)[1] AS open,
			max(price) AS high,
			min(price) AS low,
			groupArray(price)[-1] AS close
		FROM
		(
			SELECT
				instrument_token,
				toStartOfMinute(timestamp) AS time_minute,
				price
			FROM tickdata
			WHERE (instrument_token = %d) AND
			(timestamp >= '%s') AND
			(timestamp <= '%s')
		)
		GROUP BY (instrument_token, time_minute)
		ORDER BY time_minute ASC`, instrumentToken, startT, endT)

	rows, err := c.dbClient.Query(candleQueryStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candleArray Candles
	for rows.Next() {
		var (
			token       uint32
			time_minute time.Time
			open        float64
			high        float64
			low         float64
			close       float64
		)
		if err := rows.Scan(&token, &time_minute, &open, &high, &low, &close); err != nil {
			return nil, err
		}
		candle := CandleData{
			InstrumentToken: token,
			TimeStamp:       time_minute,
			Open:            open,
			High:            high,
			Low:             low,
			Close:           close,
		}
		candleArray = append(candleArray, candle)
	}

	return candleArray, nil

}
