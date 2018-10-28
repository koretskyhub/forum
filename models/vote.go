package models

import (
	"fmt"
	"forum/database"
	"log"
	"os"
)

//easyjson:json
type Vote struct {
	Nickname string `json:"nickname"`

	Voice int64 `json:"voice"`
}

func (vote *Vote) AddVoice(t *Thread) (err ModelError) {
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
	SELECT "user".id, "thread".id
	from "thread", "user"
	WHERE "thread".%s = $1 and
	"user".nickname = $2;`, slugOrId)

	var uid int64

	er = tx.QueryRow(query, value, vote.Nickname).Scan(&uid, &t.Id)
	if er != nil {
		err = ModelError{Message: NotFound}
	} else {

		tag, er := tx.Exec(`
		INSERT INTO "vote" (voice, t_id, u_id)
		values (
			case when $1 = 1 then true
			when $1 = -1 then false end,
			$2,
			$3
		)
		on conflict (t_id, u_id) do update set voice = excluded.voice ;`,
			vote.Voice, t.Id, uid)

		if er != nil {
			log.Println(er, tag.RowsAffected())
		}

		tx.Commit()

		tx, er := database.DBConnPool.Begin()
		if er != nil {
			log.Println(os.Stderr, "Unable to create transaction:", er)
		}

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
			where thread.id = $1
			group by thread.id, "user"."id", "forum"."id";`, t.Id).
			Scan(&t.Author, &t.Created, &t.Forum, &t.Id, &t.Message, &t.Slug, &t.Title, &t.Votes)
	}

	tx.Commit()

	return err
}
