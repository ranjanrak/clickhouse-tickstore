# clickhouse-tickstore

Go package to store real time [streaming websocket data](https://kite.trade/docs/connect/v3/websocket/) in [clickhouse](https://clickhouse.tech/) using queuing and bulk insert based on go-routine and channels.

## Installation

```
go get -u github.com/ranjanrak/clickhouse-tickstore
```

## Usage

```go
package main

import (
	tickstore "github.com/ranjanrak/clickhouse-tickstore"
)
func main() {
    // Create new ticker instance
    tickClient := tickstore.New(tickstore.ClientParam{
        // Send DSN as per your clickhouse DB setup.
        // visit https://github.com/ClickHouse/clickhouse-go#dsn to know more
        DBSource:    "",
        ApiKey:      "your_api_key",
        AccessToken: "your_access_token",
        TokenList:   []uint32{633601, 895745, 1723649, 3050241, 975873, 969473, 3721473, 738561, 969473},
        DumpSize:    5000,
	})
    // Start the ticker instance
    // Nothing will run after this
    tickClient.StartTicker()

    // Fetch minute candle OHLC data
    timeStart := time.Date(2022, 5, 11, 9, 51, 0, 0, time.Local)
    timeEnd := time.Date(2022, 5, 11, 10, 02, 0, 0, time.Local)
    candles, err := tickClient.FetchCandle(633601, timeStart, timeEnd)
    if err != nil {
        log.Fatalf("Error fetching candle data: %v", err)
    }
    fmt.Printf("%+v\n", candles)
}

```

## Response

> FetchCandle(633601, timeStart, timeEnd)

```
[{InstrumentToken:633601 TimeStamp:2022-05-11 09:51:00 +0530 IST Open:156.65 High:156.75 Low:156.45 Close:156.65}
{InstrumentToken:633601 TimeStamp:2022-05-11 09:52:00 +0530 IST Open:156.75 High:156.95 Low:156.7 Close:156.75}
{InstrumentToken:633601 TimeStamp:2022-05-11 09:53:00 +0530 IST Open:156.75 High:156.75 Low:156.2 Close:156.3}
{InstrumentToken:633601 TimeStamp:2022-05-11 09:54:00 +0530 IST Open:156.3 High:156.3 Low:156 Close:156.1}
......]
```

## Example

```sql
SELECT *
FROM tickdata
FINAL
WHERE (instrument_token = 633601) AND
(timestamp >= toDateTime('2022-04-22 13:23:00', 'Asia/Calcutta')) AND
(timestamp <= toDateTime('2022-04-22 13:25:00', 'Asia/Calcutta'))
ORDER BY timestamp ASC
```

```sql

Query id: 8e356516-107c-4012-948b-df90e49e9906

┌─instrument_token─┬───────────timestamp─┬──price─┐
│           633601 │ 2022-04-22 13:23:00 │ 174.65 │
│           633601 │ 2022-04-22 13:23:01 │  174.7 │
│           633601 │ 2022-04-22 13:23:02 │ 174.65 │
│           633601 │ 2022-04-22 13:23:04 │  174.7 │
│           633601 │ 2022-04-22 13:23:05 │ 174.65 │
│           633601 │ 2022-04-22 13:23:06 │  174.7 │
│           633601 │ 2022-04-22 13:23:08 │  174.7 │
│           633601 │ 2022-04-22 13:23:09 │  174.7 │
│           633601 │ 2022-04-22 13:23:10 │  174.7 │
│           633601 │ 2022-04-22 13:23:13 │  174.7 │
│           633601 │ 2022-04-22 13:23:14 │ 174.65 │
│           633601 │ 2022-04-22 13:23:15 │ 174.65 │
│           633601 │ 2022-04-22 13:23:16 │  174.7 │
│           633601 │ 2022-04-22 13:23:17 │ 174.65 │
│           633601 │ 2022-04-22 13:23:19 │  174.7 │
│           633601 │ 2022-04-22 13:23:21 │ 174.65 │
│           633601 │ 2022-04-22 13:23:24 │ 174.65 │
│           633601 │ 2022-04-22 13:23:25 │  174.7 │
│           633601 │ 2022-04-22 13:23:26 │  174.7 │
│           633601 │ 2022-04-22 13:23:27 │ 174.65 │
│           633601 │ 2022-04-22 13:23:28 │ 174.65 │
│           633601 │ 2022-04-22 13:23:29 │  174.7 │
│           633601 │ 2022-04-22 13:23:31 │  174.7 │
│           633601 │ 2022-04-22 13:23:32 │  174.7 │
│           633601 │ 2022-04-22 13:23:33 │  174.7 │
│           633601 │ 2022-04-22 13:23:35 │  174.7 │
│           633601 │ 2022-04-22 13:23:36 │  174.7 │
|           ...... | ..................  | ...... |

84 rows in set. Elapsed: 0.006 sec. Processed 8.19 thousand rows, 98.30 KB (1.28 million rows/s., 15.37 MB/s.)
```

### Fetch OHLC candle data for any given interval using [CTE](https://clickhouse.com/docs/en/sql-reference/statements/select/with/) and [aggregate function](https://clickhouse.com/docs/en/sql-reference/aggregate-functions/reference/grouparray/)

```sql
WITH price_select AS (SELECT price
FROM tickdata
FINAL
WHERE (instrument_token = 975873) AND
(timestamp >= toDateTime('2022-05-02 14:47:00')) AND
(timestamp <= toDateTime('2022-05-02 14:47:59'))
ORDER BY timestamp ASC)
SELECT groupArray(price)[1] AS open,
max(price) AS high,
min(price) AS low,
groupArray(price)[-1] AS close FROM price_select;
```

```sql
Query id: 98d92c26-e054-4f0a-8448-064bc0d939a0
┌───open─┬───high─┬───low─┬─close─┐
│ 252.25 │ 252.35 │ 252.1 │ 252.2 │
└────────┴────────┴───────┴───────┘
```

### Create base minute candle OHLC

> Which will be used further to calculate other candle intervals(3Min, 5Min, 15Min, etc)

```sql
SELECT
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
    WHERE (instrument_token = 975873) AND
    (timestamp >= toDateTime('2022-05-02 14:47:00')) AND
    (timestamp <= toDateTime('2022-05-02 14:59:59'))
)
GROUP BY (instrument_token, time_minute)
ORDER BY time_minute ASC
```

```sql
Query id: 2ba74fd2-6047-42c9-9436-be8987a5d3a9

┌─instrument_token─┬─────────time_minute─┬───open─┬───high─┬────low─┬──close─┐
│           975873 │ 2022-05-02 14:47:00 │ 252.25 │ 252.35 │  252.1 │  252.2 │
│           975873 │ 2022-05-02 14:48:00 │  252.2 │  252.3 │  251.9 │ 252.25 │
│           975873 │ 2022-05-02 14:49:00 │  252.3 │  252.3 │ 252.05 │  252.1 │
│           975873 │ 2022-05-02 14:50:00 │  252.1 │ 252.45 │ 252.05 │ 252.35 │
│           975873 │ 2022-05-02 14:51:00 │  252.2 │ 252.45 │  252.2 │ 252.35 │
│           975873 │ 2022-05-02 14:52:00 │ 252.35 │ 252.35 │    252 │    252 │
│           975873 │ 2022-05-02 14:53:00 │    252 │ 253.15 │    252 │  252.8 │
│           975873 │ 2022-05-02 14:54:00 │  252.8 │  253.2 │  252.7 │  252.8 │
│           975873 │ 2022-05-02 14:55:00 │  252.8 │  253.4 │ 252.75 │  253.3 │
│           975873 │ 2022-05-02 14:56:00 │ 253.25 │  253.4 │    253 │  253.1 │
│           975873 │ 2022-05-02 14:57:00 │  253.1 │  253.1 │ 252.85 │ 252.85 │
│           975873 │ 2022-05-02 14:58:00 │  252.8 │ 253.05 │  252.5 │  252.6 │
│           975873 │ 2022-05-02 14:59:00 │  252.6 │  253.4 │  252.6 │  253.4 │
└──────────────────┴─────────────────────┴────────┴────────┴────────┴────────┘
```

### Create candle_data materialized views to store minute OHLC

```sql
CREATE MATERIALIZED VIEW candle_data
ENGINE = ReplacingMergeTree
ORDER BY (instrument_token, time_minute)
PRIMARY KEY (instrument_token, time_minute) POPULATE AS
SELECT
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
)
GROUP BY (instrument_token, time_minute)
```
