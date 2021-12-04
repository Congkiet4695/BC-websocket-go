package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type owner struct {
	new_owner_address string
	transaction_hash  string
	block_number      string
	timestamp         int64
}

func main() {

	//Create DB
	db, _ := sql.Open("sqlite3", ".testnet.db")

	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "Owner" (
			"new_owner_address"	TEXT,
			"timestamp"	INTEGER,
			"transaction_hash"	TEXT,
			"block_number"	TEXT,
			"id"	INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`)
	stmt.Exec()

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			fmt.Printf("%s Data String: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}

			new_owner_address := getnewOwnerAddress(string(msg))
			fmt.Printf("%s Address To: %s\n", conn.RemoteAddr(), new_owner_address)

			transaction_hash := getTransactionHash(string(msg))
			fmt.Printf("%s Transaction Hash: %s\n", conn.RemoteAddr(), transaction_hash)

			block_number := getblockNumber(string(msg))
			fmt.Printf("%s Block number: %s\n", conn.RemoteAddr(), block_number)

			now := time.Now() // current local time
			sec := now.Unix()

			o := owner{
				new_owner_address: new_owner_address,
				transaction_hash:  transaction_hash,
				block_number:      block_number,
				timestamp:         sec,
			}
			err = insert(db, o)
			if err != nil {
				log.Printf("Insert owner failed with error %s", err)
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}

func getnewOwnerAddress(address string) string {
	first := strings.Index(address, "newOwnerAdd") + 11
	last := len(address)
	return address[first:last]
}
func getTransactionHash(address string) string {
	first := strings.Index(address, "transactionHash") + 15
	last := strings.Index(address, "newOwnerAdd")
	return address[first:last]
}

func getblockNumber(address string) string {
	first := strings.Index(address, "blockNumber") + 11
	last := strings.Index(address, "transactionHash")
	return address[first:last]
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
