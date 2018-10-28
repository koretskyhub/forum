package models

import (
	"forum/database"
	"log"
	"os"
)

//easyjson:json
type Forum struct {
	Posts int64 `json:"posts,omitempty"`

	Slug string `json:"slug"`

	Threads int64 `json:"threads,omitempty"`

	Title string `json:"title"`

	User string `json:"user"`
}

func (f *Forum) Create() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}
	var uid int
	er = tx.QueryRow(`
	SELECT id FROM "user" WHERE nickname = $1;`,
		f.User).Scan(&uid)
	if er != nil {
		log.Println(os.Stderr, err)
		err = ModelError{Message: NotFound}
	} else {
		log.Println(uid)
		tag, er := tx.Exec(`
			INSERT INTO "forum" (slug, title, u_id)
			VALUES (
				$1,
				$2,
				$3)
			ON CONFLICT DO NOTHING;`,
			f.Slug, f.Title, uid)
		if tag.RowsAffected() == 0 {
			log.Println(os.Stderr, err)
			err = ModelError{Message: Conflict}
		}
		er = tx.QueryRow(`
		select
			count("post"."id") as posts,
			"forum"."slug" as slug,
			count(distinct "thread"."id") as threads,
			"forum"."title" as title,
			"user"."nickname" as user
		from "forum"
		left join "thread" on "thread"."f_id" = "forum"."u_id"
		left join "post" on "thread"."id" = "post"."t_id"
		join "user" on "forum"."u_id" = "user"."id"
		where "forum"."slug" = $1
		group by "forum"."id", "user"."id";`, f.Slug).
			Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
		if er != nil {
			log.Println("cannot get forum details", er)
		}
	}
	defer tx.Rollback()

	tx.Commit()

	return err
}

func (f *Forum) GetDetails() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}

	//removed filter by forum title in where
	er = tx.QueryRow(`
	select
		count("post"."id") as posts,
		"forum"."slug" as slug,
		count(distinct "thread"."id") as threads,
		"forum"."title" as title,
		"user"."nickname" as user
	from "forum"
	left join "thread" on "thread"."f_id" = "forum"."id"
	left join "post" on "thread"."id" = "post"."t_id"
	join "user" on "forum"."u_id" = "user"."id"
	where "forum"."slug" = $1
	group by "forum"."id", "user"."id";`, f.Slug).
		Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
	if er != nil {
		log.Println("cannot get forum details", er)
		err = ModelError{Message: NotFound}
	}

	defer tx.Rollback()

	tx.Commit()

	return err
}
