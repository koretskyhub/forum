package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx"
)

var DBConnPool *pgx.ConnPool

func InitDb() (err error) {
	config := pgx.ConnConfig{
		Host:     "localhost",
		Database: "docker",
		User:     "docker",
		Password: "docker",
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 50,
	}

	DBConnPool, err = pgx.NewConnPool(poolConfig)

	if err != nil {
		log.Println(os.Stderr, "Unable to create conn pool:", err)
		os.Exit(1)
	}

	tx, err := DBConnPool.Begin()
	if err != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
		os.Exit(1)
	}

	bFile, err := ioutil.ReadFile(fmt.Sprintf("%s/src/forum/database/table_creater.sql", os.Getenv("GOPATH")))
	if err != nil {
		log.Println(os.Stderr, "Unable to read sql script:", err)
		os.Exit(1)
	}

	_, err = tx.Exec(string(bFile))
	if err != nil {
		log.Println(os.Stderr, "Unable to exec query:", err)
		os.Exit(1)
	}

	err = tx.Commit()

	if err != nil {
		log.Println(os.Stderr, "Unable to commit:", err)
		os.Exit(1)
	}

	DBConnPool.Reset()
	return nil
}
