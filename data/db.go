package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DBClient *sql.DB

func InitDB() {

	connStr := os.Getenv("DB_URI")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	DBClient = db
}

func InitTables() {
	queries, err := os.ReadFile("./data/tables.sql")
	if err != nil {
		panic(err)
	}

	querystr := string(queries)

	res, dberr := DBClient.Exec(querystr)

	if dberr != nil {
		fmt.Println("DB ERROR: ", dberr)
	}

	fmt.Println(res)
}