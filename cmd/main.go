package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/mattn/go-sqlite3"
)

type Websocket struct {
	*ethclient.Client
	query ethereum.FilterQuery
}

type owner struct {
	new_owner_address common.Address
	transaction_hash  string
	block_number      string
	timestamp         int64
}

func Init() {
	client, err := ethclient.Dial("wss://testnet-dex.binance.org/api/ws")
	if err != nil {
		log.Fatal(err)
	}
	websocket := &Websocket{Client: client, query: ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress("0x98b3f2219a2b7a047B6234c19926673ad4aac83A"),
		},
		Topics: [][]common.Hash{{
			common.HexToHash("0x342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a735"),
		}},
	}}

	websocket.Connect()
}

func handleEvent(vLog types.Log) {
	//sender := common.HexToAddress(vLog.Topics[1].Hex())
	receiver := common.HexToAddress(vLog.Topics[2].Hex())
	blockNumber := hexutil.EncodeUint64(vLog.BlockNumber)
	transactionHash := vLog.TxHash.String()
	//value := hexutil.Encode(vLog.Data)
	now := time.Now().Unix()
	//manual := false

	o := owner{
		new_owner_address: receiver,
		transaction_hash:  transactionHash,
		block_number:      blockNumber,
		timestamp:         now,
	}

	db, _ := sql.Open("sqlite3", ".testnet.db")

	err := insert(db, o)
	if err != nil {
		log.Printf("Insert owner failed with error %s", err)
		return
	}

}

func (websocket *Websocket) Connect() {
	channel := make(chan types.Log)

	sub, err := websocket.SubscribeFilterLogs(context.Background(), websocket.query, channel)
	log.Printf("err: %+v", err)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-channel:
			log.Println(vLog)
			handleEvent(vLog)
		}
	}
}

func main() {

	// stmt, _ := db.Prepare(`
	// 	CREATE TABLE IF NOT EXISTS "Owner" (
	// 		"new_owner_address"	TEXT,
	// 		"timestamp"	INTEGER,
	// 		"transaction_hash"	TEXT,
	// 		"block_number"	TEXT,
	// 		"id"	INTEGER NOT NULL,
	// 		PRIMARY KEY("id" AUTOINCREMENT)
	// 	);
	// `)
	// stmt.Exec()

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "index.html")
	// })

	// http.ListenAndServe(":8080", nil)

	Init()

}

func insert(db *sql.DB, o owner) error {
	query := "INSERT INTO Owner (new_owner_address, transaction_hash, block_number, timestamp) VALUES (?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, o.new_owner_address, o.transaction_hash, o.block_number, o.timestamp)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d owner created ", rows)
	return nil
}
