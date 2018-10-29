package models

import (
	"forum/database"
	"log"
	"os"
)

//easyjson:json
type Status struct {
	Forum int64 `json:"forum"`

	Post int64 `json:"post"`

	Thread int64 `json:"thread"`

	User int64 `json:"user"`
}

func (s *Status) ResetData() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}
	_, er = tx.Exec(`
	delete from "vote";
	delete from "post";
	delete from "thread";
	delete from "forum";
	delete from "user";`)
	if er != nil {
		log.Println(er)
	}

	defer tx.Rollback()

	tx.Commit()

	database.DBConnPool.Reset()

	return err
}

func (s *Status) GetStatus() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}
	er = tx.QueryRow(`
	select 
		(select count(*) from forum) as forum,
		(select count(*) from post) as post,
		(select count(*) from thread) as thread,
		(select count(*) from "user") as user;`).Scan(&s.Forum, &s.Post, &s.Thread, &s.User)
	if er != nil {
		log.Println(er)
	}

	return err
}
