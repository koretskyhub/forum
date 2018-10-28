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
type Post struct {
	Author string `json:"author,omitempty"`

	Created time.Time `json:"created,omitempty"`

	Forum string `json:"forum,omitempty"`

	Id int `json:"id,omitempty"`

	IsEdited bool `json:"isEdited,omitempty"`

	Message string `json:"message,omitempty"`

	Parent int64 `json:"parent,omitempty"`

	Thread int64 `json:"thread,omitempty"`
}

//easyjson:json
type PostFull struct {
	Author *User `json:"author,omitempty"`

	Forum *Forum `json:"forum,omitempty"`

	Post *Post `json:"post"`

	Thread *Thread `json:"thread,omitempty"`
}

//easyjson:json
type Posts []*Post

func (posts *Posts) Create(t *Thread) (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", err)
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

	er = tx.QueryRow(`
			SELECT "thread"."id", forum.slug
			FROM "thread"
			JOIN forum on forum.id = "thread"."f_id"
			WHERE "thread".`+slugOrId+` = $1;`,
		value).Scan(&t.Id, &t.Forum)
	if er != nil {
		err = ModelError{Message: NotFound}
		return err
	}

	// tx.Commit()

	// postsBatch := database.DBConnPool.BeginBatch()

	createdTime := time.Now()

	for _, p := range *posts {

		p.Created = createdTime
		p.Thread = t.Id
		p.Forum = t.Forum

		var authorId int64

		er = tx.QueryRow(`
		select "id" from "user" where nickname = $1`, p.Author).
			Scan(&authorId)

		if authorId == 0 {
			err = ModelError{Message: NotFound}
		} else {
			er = tx.QueryRow(`
			INSERT INTO "post" (message, created, path, u_id, t_id)
			select 
			$1,
			$2,
			(SELECT path FROM "post" WHERE id = $3) || coalesce((select max(id) from "post")::integer, 0) + 1,
			$4,
			$5
			where exists(SELECT path FROM "post" WHERE id = $3 and t_id = $5) or $3 = 0
			returning id;`,
				p.Message, p.Created, p.Parent, authorId, p.Thread).Scan(&p.Id)

			if er != nil {
				err = ModelError{Message: Conflict}
			}
		}
	}

	// if er = postsBatch.Send(context.Background(), nil); er != nil {
	// 	log.Println("batch sending error")
	// }

	// tag, er := postsBatch.ExecResults()

	// if er != nil || tag.RowsAffected() < int64(len(*posts)) {
	// 	err = ModelError{Message: Conflict}
	// }

	// log.Println(tag.RowsAffected())

	// postsBatch.Close()
	tx.Commit()

	return err
}

func (posts *Posts) GetPostsByThread(t *Thread, sort string, limit int64, since string, desc bool) (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", er)
	}

	if (t.Id == 0) && (t.Slug != "") {
		er = tx.QueryRow(`select id from thread where thread.slug = $1`, t.Slug).Scan(&t.Id)
	} else {
		er = tx.QueryRow(`select id from thread where thread.id = $1`, t.Id).Scan(&t.Id)
	}

	if er != nil {
		log.Println(er)
		tx.Commit()
		return ModelError{Message: NotFound}
	}

	sinceInt, er := strconv.Atoi(since)
	var sinceInt64 int64
	if er != nil {
		sinceInt64 = 0
	}

	sinceInt64 = int64(sinceInt)

	queryDesc := "asc"

	if desc {
		queryDesc = "desc"
	}

	var sinceCond string
	var query string

	switch sort {
	case "flat", "":
		if (queryDesc == "asc") || (sinceInt64 == 0) {
			sinceCond = `and $1 < any (post.path)`
		} else {
			sinceCond = `and $1 > all (post.path)`
		}
		query = fmt.Sprintf(`
			select 
				"user"."nickname" as author,
				sorted_posts.created as created,
				forum.slug as forum,
				sorted_posts.id as id,
				sorted_posts.message as message,
				coalesce(sorted_posts.path[array_length(sorted_posts.path, 1) - 1], 0) as parent,
				thread.id as thread
			from
				(
					select *
					from post
					where post.t_id = $2
					%s
				) as sorted_posts
			join "user" on "user"."id" = sorted_posts.u_id
			join thread on thread.id = sorted_posts.t_id
			join forum on forum.id = thread.f_id
			order by sorted_posts.created `+queryDesc+`, sorted_posts.id `+queryDesc+`
			limit $3;`, sinceCond)
	case "tree":
		if queryDesc == "desc" {
			sinceCond = `<`
		} else {
			sinceCond = `>`
		}
		query = fmt.Sprintf(`
		select 
			"user"."nickname" as author,
			sorted_posts.created as created,
			forum.slug as forum,
			sorted_posts.id as id,
			sorted_posts.message as message,
			coalesce(sorted_posts.path[array_length(sorted_posts.path, 1) - 1], 0) as parent,
			thread.id as thread
		from
			(
				select *
				from post
				where post.t_id = $2
				and coalesce(path %s (select path from post where id = $1), true)
			) as sorted_posts
		join "user" on "user"."id" = sorted_posts.u_id
		join thread on thread.id = sorted_posts.t_id
		join forum on forum.id = thread.f_id
		order by sorted_posts.path[1] `+queryDesc+`, sorted_posts.path[2:]  `+queryDesc+` nulls first
		limit $3;`, sinceCond)

	case "parent_tree":
		if queryDesc == "desc" {
			sinceCond = `<`
		} else {
			sinceCond = `>`
		}
		query = fmt.Sprintf(`
		select 
			"user"."nickname" as author,
			sorted_posts.created as created,
			forum.slug as forum,
			sorted_posts.id as id,
			sorted_posts.message as message,
			coalesce(sorted_posts.path[array_length(sorted_posts.path, 1) - 1], 0) as parent,
			thread.id as thread
		from
			(
				select *
				from post
				where post.t_id = $2 and path[1] in
				(
					select path[1] from post
					where post.t_id = $2
					and coalesce(path[1] %s (select path[1] from post where id = $1), true)
					group by post.path[1]
					order by path[1] `+queryDesc+`
					limit $3
				)
			) as sorted_posts
		join "user" on "user"."id" = sorted_posts.u_id
		join thread on thread.id = sorted_posts.t_id
		join forum on forum.id = thread.f_id
		order by sorted_posts.path[1] `+queryDesc+`, sorted_posts.path[2:] asc nulls first, sorted_posts.created `+queryDesc+`;`, sinceCond)
	default:
		log.Println("undefined sort")
		return ModelError{}
	}

	// log.Println(query)

	row, er := tx.Query(query, sinceInt64, t.Id, limit)

	// log.Println(er)

	for row.Next() {
		p := Post{}
		er = row.Scan(&p.Author, &p.Created, &p.Forum, &p.Id, &p.Message, &p.Parent, &p.Thread)
		if er != nil {
			log.Println(er, "cannot scan row to post")
		}
		*posts = append(*posts, &p)
	}
	row.Close()

	tx.Commit()

	return err
}

func (postF *PostFull) Details(related *[]string) (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println(os.Stderr, "Unable to create transaction:", er)
	}

	er = tx.QueryRow(`
	select 
		"user"."nickname" as author,
		post.created as created,
		forum.slug as forum,
		post.id as id,
		post.is_edited as isEdited,
		post.message as message,
		coalesce(post.path[array_length(post.path, 1) - 1], 0) as parent,
		thread.id as thread
	from post
	join "user" on "user"."id" = post.u_id
	join thread on thread.id = post.t_id
	join forum on forum.id = thread.f_id
	where post.id = $1;`, postF.Post.Id).
		Scan(&postF.Post.Author, &postF.Post.Created, &postF.Post.Forum, &postF.Post.Id,
			&postF.Post.IsEdited, &postF.Post.Message, &postF.Post.Parent, &postF.Post.Thread)

	if er != nil {
		log.Println(er)
		err = ModelError{Message: NotFound}
		tx.Rollback()
	}

	tx.Commit()


	for _, rel := range *related{
	switch rel {
		case "user":
			user := User{Nickname: postF.Post.Author}
			user.GetProfile()
			postF.Author = &user
		case "thread":
			thread := Thread{Id: postF.Post.Thread}
			thread.Get()
			postF.Thread = &thread
		
		case "forum":
			forum := Forum{Slug: postF.Post.Forum}
			forum.GetDetails()
			postF.Forum = &forum
		}
	}

	return err
}

func (post *Post) Update() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	if er != nil {
		log.Println("Unable to create transaction:", er)
	}

	tag, er := tx.Exec(`
	UPDATE "post" SET
		is_edited = case when ($1 = message or $1 = '') then false
						else true
					end,
		message = case when ($1 = message or $1 = '') then message
					else $1
				end
	WHERE id = $2;`,post.Message, post.Id)

	log.Println(er)


	if tag.RowsAffected() == 0 {
		err = ModelError{Message: NotFound}
	} else {
		er = tx.QueryRow(`
		select 
			"user"."nickname" as author,
			post.created as created,
			forum.slug as forum,
			post.id as id,
			post.is_edited as isEdited,
			post.message as message,
			coalesce(post.path[array_length(post.path, 1) - 1], 0) as parent,
			thread.id as thread
		from post
		join "user" on "user"."id" = post.u_id
		join thread on thread.id = post.t_id
		join forum on forum.id = thread.f_id
		where post.id = $1;`, post.Id).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Id,
				&post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	}

	defer tx.Rollback()
	
	tx.Commit()
	
	return err
}