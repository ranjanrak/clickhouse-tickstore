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
    tickClient.StartTicker()
}

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
