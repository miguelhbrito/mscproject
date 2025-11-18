package dbconnect

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

const (
	host     = "postgresdb_container"
	port     = 5432
	user     = "hel"
	password = "saymyname"
	dbname   = "asgard"
)

func ConnectDB() (db *sql.DB, err error) {
	log := log.New(os.Stdout, "mscproject DB: ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println("error to open db", err)
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		log.Println("error to ping db", err)
		panic(err)
	}

	log.Println("successfully db connected")
	return db, err
}
