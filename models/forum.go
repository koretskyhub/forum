package models

import (
	"forum/database"
	"log"
	"os"
)

//easyjson:json
type Forum struct {
	Posts   int64  `json:"posts,omitempty"`
	Slug    string `json:"slug"`
	Threads int64  `json:"threads,omitempty"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

func (f *Forum) Create() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	// log.Println("create forum")

	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}
	var uid int
	er = tx.QueryRow(`SELECT nickname, id FROM "user" WHERE nickname = $1;`, f.User).Scan(&f.User, &uid)

	if er != nil {
		log.Println(os.Stderr, err)
		err = ModelError{Message: NotFound}
	} else {
		tag, er := tx.Exec(`
		INSERT INTO "forum" (slug, title, u_id, u_nickname)
		VALUES (
			$1,
			$2,
			$3,
			$4)
		ON CONFLICT DO NOTHING;
		`,f.Slug, f.Title, uid, f.User)

		if tag.RowsAffected() == 0 {
			log.Println(os.Stderr, err)
			err = ModelError{Message: Conflict}
		}
		// er = tx.QueryRow(`
		// select
		// 	count("post"."id") as posts,
		// 	"forum"."slug" as slug,
		// 	count(distinct "thread"."id") as threads,
		// 	"forum"."title" as title,
		// 	"user"."nickname" as user
		// from "forum"
		// left join "thread" on "thread"."f_id" = "forum"."id"
		// left join "post" on "thread"."id" = "post"."t_id"
		// join "user" on "forum"."u_id" = "user"."id"
		// where "forum"."slug" = $1
		// group by "forum"."id", "user"."id";
		// `, f.Slug).
		// 	Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)

		er = tx.QueryRow(`
		select
			"posts" as posts,
			"slug" as slug,
			"threads" as threads,
			"title" as title,
			"u_nickname" as user
		from "forum"
		where "slug" = $1;
		`, f.Slug).
			Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)

		if er != nil {
			log.Println(er)
		}
	}

	tx.Commit()

	return err
}

//optimized
func (f *Forum) GetDetails() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	// log.Println("get forum")

	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}

	er = tx.QueryRow(`
	select
		"posts" as posts,
		"slug" as slug,
		"threads" as threads,
		"title" as title,
		"u_nickname" as user
	from "forum"
	where "slug" = $1;
	`, f.Slug).
		Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
	if er != nil {
		// log.Println("cannot get forum details", er)
		err = ModelError{Message: NotFound}
	}

	tx.Commit()

	return err
}
