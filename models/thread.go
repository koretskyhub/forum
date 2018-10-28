package models

import (
	"fmt"
	"forum/database"
	"log"
	"os"
	"strconv"
	"time"
)

//easyjson:json
type Thread struct {
	Author string `json:"author"`

	Created time.Time `json:"created"`

	Forum string `json:"forum"`

	Id int64 `json:"id"`

	Message string `json:"message"`

	Slug string `json:"slug,omitempty"`

	Title string `json:"title"`

	Votes int64 `json:"votes,omitempty"`
}

//easyjson:json
type Threads []*Thread

func (t *Thread) Create() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}
	var uid, fid int
	er = tx.QueryRow(`
			SELECT "user".id, "forum".id, "forum".slug
			FROM "user", "forum"
			WHERE "user".nickname = $1 
			and "forum".slug = $2;`,
		t.Author, t.Forum).Scan(&uid, &fid, &t.Forum)
	if er != nil {
		log.Println(os.Stderr, err)
		err = ModelError{Message: NotFound}
	} else {
		er = tx.QueryRow(`
			INSERT INTO "thread" (slug, created, title, message, u_id, f_id)
			VALUES (NULLIF($1, ''),	$2,	$3,	$4, $5, $6)
			ON CONFLICT DO NOTHING
			RETURNING id;`,
			t.Slug, t.Created, t.Title, t.Message, uid, fid).Scan(&t.Id)
		log.Println(t.Id)
		if t.Id == 0 {
			log.Println(os.Stderr, err)
			err = ModelError{Message: Conflict}
			er = tx.QueryRow(`
			select 
				"user"."nickname" as author,
				thread.created as created,
				"forum"."slug" as forum,
				thread.id as id,
				thread.message as message,
				thread.slug as slug,
				thread.title as title,
				COALESCE(sum(CASE 
					WHEN vote.voice THEN 1 
					WHEN NOT vote.voice THEN -1
					END ),0) as votes
			from thread
			join "user" on "user"."id" = thread.u_id
			join "forum" on "forum"."id" = thread.f_id
			left join vote on vote.t_id = thread.id
			where thread.slug = $1
			group by thread.id, "user"."id", "forum"."id";`, t.Slug).
				Scan(&t.Author, &t.Created, &t.Forum, &t.Id, &t.Message, &t.Slug, &t.Title, &t.Votes)
		}
	}
	defer tx.Rollback()

	tx.Commit()

	return err
}

func (t *Thread) Get() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", er)
	}

	var slugOrId string
	var value interface{}

	if (t.Id != 0) && (t.Slug == "") {
		slugOrId = "id"
		value = t.Id
	} else {
		slugOrId = "slug"
		value = t.Slug
	}

	query := fmt.Sprintf(`
	select 
		"user"."nickname" as author,
		thread.created as created,
		"forum"."slug" as forum,
		thread.id as id,
		thread.message as message,
		thread.slug as slug,
		thread.title as title,
		COALESCE(sum(CASE 
						WHEN vote.voice THEN
							1 
						WHEN NOT vote.voice THEN
							-1
						END),0) as votes
	from thread
	join "user" on "user"."id" = thread.u_id
	join "forum" on "forum"."id" = thread.f_id
	left join vote on vote.t_id = thread.id
	where thread.%s = $1
	group by thread.id, "user"."id", "forum"."id";`, slugOrId)

	er = tx.QueryRow(query, value).
		Scan(&t.Author, &t.Created, &t.Forum, &t.Id, &t.Message, &t.Slug, &t.Title, &t.Votes)
	if er != nil {
		err = ModelError{Message: NotFound}
		log.Println(er)
	}

	defer tx.Rollback()

	tx.Commit()

	return err
}

func (t *Thread) Update() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", er)
	}

	var slugOrId string
	var value interface{}

	if (t.Id != 0) && (t.Slug == "") {
		slugOrId = "id"
		value = t.Id
	} else {
		slugOrId = "slug"
		value = t.Slug
	}

	query := fmt.Sprintf(`
	UPDATE "thread" SET 
		message = case when $1 <> '' then  $1 
		else message end,
		title = case when $2 <> '' then  $2 
		else title end 
	WHERE %s = $3
	RETURNING id;`, slugOrId)

	er = tx.QueryRow(query, t.Message, t.Title, value).Scan(&t.Id)

	if (er != nil) || t.Id == 0 {
		err = ModelError{Message: NotFound}
	} else {
		er = tx.QueryRow(`
		select 
			"user"."nickname" as author,
			thread.created as created,
			"forum"."slug" as forum,
			thread.id as id,
			thread.message as message,
			thread.slug as slug,
			thread.title as title,
			COALESCE(sum(CASE 
							WHEN vote.voice THEN
								1 
							WHEN NOT vote.voice THEN
								-1
								END),0) as votes
		from thread
		join "user" on "user"."id" = thread.u_id
		join "forum" on "forum"."id" = thread.f_id
		left join vote on vote.t_id = thread.id
		where thread.id = $1
		group by thread.id, "user"."id", "forum"."id";`, t.Id).
			Scan(&t.Author, &t.Created, &t.Forum, &t.Id, &t.Message, &t.Slug, &t.Title, &t.Votes)
		if er != nil {
			log.Println(er)
		}
	}

	defer tx.Rollback()

	tx.Commit()

	return err
}

func (threads *Threads) GetThreadsByForum(forumSlug string, limit int64, since string,
	desc bool) (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
	}

	var sinceQuery string
	var descQuery string

	if desc {
		descQuery = "desc"
		sinceQuery += " <= "
		if since == "" {
			sinceQuery += "'infinity'"
		}
	} else {
		descQuery = "asc"
		sinceQuery += " >= "
		if since == "" {
			sinceQuery += "'-infinity'"
		}
	}

	if since != "" {
		sinceQuery += "'" + since + "'"
	}

	var limitQuery string
	if limit < 0 {
		limitQuery = "null"
	} else {
		limitQuery = strconv.FormatInt(limit, 10)
	}

	row, er := tx.Query(`select 1 from forum where slug = $1`, forumSlug)
	forumExisting := row.Next()
	row.Close()
	if !forumExisting {
		err = ModelError{Message: NotFound}
		return err
	}

	row, er = tx.Query(`select 
							"user"."nickname" as author,
							thread.created as created,
							forum.slug as forum,
							thread.id as id,
							thread.message as message,
							thread.slug as slug,
							thread.title as title,
							COALESCE(sum(CASE WHEN vote.voice = true THEN 1
									WHEN vote.voice = false THEN -1
									ELSE 0
									END),0) as votes
						from thread
						join forum on thread.f_id = forum.id
						join "user" on thread.u_id = "user"."id"
						left join "vote" on thread.id = "vote"."t_id"
						where forum.slug = $1
						and thread.created`+sinceQuery+`
						group by "user"."nickname", thread.created, forum.slug, thread.id, thread.message, thread.slug, thread.title
						order by thread.created `+descQuery+` limit $2;`,
		forumSlug, limitQuery)
	if er != nil {
		log.Println(os.Stderr, err)
	} else {
		for row.Next() {
			t := Thread{}

			er = row.Scan(&t.Author, &t.Created, &t.Forum, &t.Id,
				&t.Message, &t.Slug, &t.Title, &t.Votes)

			if er != nil {
				log.Println(os.Stderr, err)
			}

			*threads = append(*threads, &t)
		}
	}

	row.Close()

	tx.Commit()

	return err
}
