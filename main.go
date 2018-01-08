package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var databaseName, databaseURL string

func main() {
	port, databaseName, databaseURL, setUpDatabase := readFileConfig()
	initConnection(databaseName, databaseURL)
	if setUpDatabase == "create" {
		createDatabase()
	}
	http.HandleFunc("/save", saveData)
	fmt.Println("localhost:" + port + " runing...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func saveData(w http.ResponseWriter, r *http.Request) {
	db := getConnection()
	defer db.Close()
	if r.Method == http.MethodPost {
		body := bodyToJSON(r)
		t := time.Now()
		now := t.Format("2006-01-02 15:04:05")
		sqlQuery := `INSERT INTO BARCODE_COMPARE (
		createDate,
		part_master,
		qty_master,
		part_compare,
		batch_compare,
		qty_compare
	) VALUES (?,?,?,?,?,?)`
		_, err := db.Exec(
			sqlQuery,
			now,
			body["partMaster"],
			body["qtyMaster"],
			body["partCompare"],
			body["batchCompare"],
			body["qtyCompare"],
		)
		checkErr(err)
		mapData := make(map[string]string)
		mapData["status"] = "success"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapData)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func bodyToJSON(r *http.Request) map[string]string {
	body := map[string]string{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	r.Body.Close()
	return body
}

func initConnection(dbName, dbURL string) {
	databaseName = dbName
	databaseURL = dbURL
	fmt.Println("Database connect...")
}

func getConnection() *sql.DB {
	dbConnect, err := sql.Open(databaseName, databaseURL)
	if err != nil {
		log.Fatalf("can not connect database : %v", err)
	}
	return dbConnect
}

func createDatabase() {
	sqlCreateTable := `
	CREATE TABLE BARCODE_COMPARE (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		createDate datetime DEFAULT NULL,
		part_master varchar(255) DEFAULT NULL,
		qty_master INTEGER DEFAULT NULL,
		part_compare varchar(255) DEFAULT NULL,
		batch_compare varchar(255) DEFAULT NULL,
		qty_compare INTEGER DEFAULT NULL
	);
	`
	db := getConnection()
	defer db.Close()
	_, err := db.Exec(sqlCreateTable)
	checkErr(err)
	fmt.Println("Create Table...")
}

func readFileConfig() (portOut, databaseNameOut, databaseURLOut, setUpDatabaseOut string) {
	config, err := ioutil.ReadFile("./config.json")
	checkErr(err)
	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(config, &objmap)
	checkErr(err)
	var port, databaseName, databaseURL, setUpDatabase string
	err = json.Unmarshal(*objmap["port"], &port)
	checkErr(err)
	err = json.Unmarshal(*objmap["databaseName"], &databaseName)
	checkErr(err)
	err = json.Unmarshal(*objmap["databaseURL"], &databaseURL)
	checkErr(err)
	err = json.Unmarshal(*objmap["setUpDatabase"], &setUpDatabase)
	checkErr(err)
	return port, databaseName, databaseURL, setUpDatabase
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
