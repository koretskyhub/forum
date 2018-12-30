package models

import (
	"forum/database"
	"log"
)

//easyjson:json
type User struct {
	About    string `json:"about,omitempty"`
	Email    string `json:"email,omitempty"`
	Fullname string `json:"fullname,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

//easyjson:json
type Users []*User

func (u *User) Create() (users *Users, err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	log.Println("create user")

	if er != nil {
		log.Println("Unable to create transaction:", er)
	}

	if rows, er := tx.Query(`
		SELECT about, email, fullname, nickname
		FROM "user"
		WHERE email = $1 
		OR nickname = $2;`,
		u.Email,
		u.Nickname); er != nil {
		// log.Println(er)
	} else {
		if foundUser := rows.Next(); foundUser == true {
			users = &Users{}
			for foundUser {
				user := User{}
				if rows.Err() != nil {
					// log.Println(rows.Err())
				}
				rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
				*users = append(*users, &user)
				foundUser = rows.Next()
			}
			rows.Close()
			err = ModelError{Message: NotFound}
		} else {
			_, er = tx.Exec(`	
			INSERT INTO "user" (about, email, fullname, nickname)
			VALUES ($1, $2, $3, $4);`,
				u.About, u.Email, u.Fullname, u.Nickname)
			if er != nil {
				// log.Println(er)
			}
		}
	}

	tx.Commit()
	return users, err
}

//optimize
func (u *User) GetProfile() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	log.Println("Getting profile")

	if er != nil {
		log.Println("Unable to create transaction:", er)
	}

	er = tx.QueryRow(`
	SELECT about, email, fullname, nickname
	FROM "user" WHERE nickname = $1;`,
		u.Nickname).Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)

	if er != nil {
		log.Println("cannot get profile", er)
		err = ModelError{Message: NotFound}
	}

	tx.Commit()
	return err
}

func (u *User) UpdateProfile() (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	log.Println("update user")

	if er != nil {
		log.Println("Unable to create transaction:", er)
	}

	exist := false

	er = tx.QueryRow(`
		SELECT true FROM "user"
		WHERE nickname = $1;`,
		u.Nickname).Scan(&exist)

	if exist {
		er = tx.QueryRow(`
			UPDATE "user"
			SET about = COALESCE(NULLIF($1, ''), about),
			email = COALESCE(NULLIF($2, ''), email),
			fullname = COALESCE(NULLIF($3, ''), fullname)
			WHERE nickname = $4
			RETURNING "user".about,
					"user".email,
					"user".fullname,
					"user".nickname;`,
			u.About,
			u.Email,
			u.Fullname,
			u.Nickname,
		).Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)

		if er != nil {
			err = ModelError{Message: Conflict}
		}
	} else {
		err = ModelError{Message: NotFound}
	}

	tx.Commit()
	return err
}

//optimize
func (useres *Users) GetByForum(forum *Forum, limit int64, since string, desc bool) (err ModelError) {
	tx, er := database.DBConnPool.Begin()
	defer tx.Rollback()
	log.Println("get user by forum")

	if er != nil {
		log.Println("Unable to create transaction:", er)
	}

	var fid int64

	er = tx.QueryRow(`select id from forum where forum.slug = $1`, forum.Slug).
		Scan(&fid)
	if er != nil {
		// log.Println(er)
		err = ModelError{Message: NotFound}
		return err
	}

	if limit <= 0 {
		limit = 100
	}

	var descQuery string
	var sinceQuery string

	if desc {
		descQuery = "desc"
		if since != "" {
			sinceQuery = `"user"."nickname" < $2`
		} else {
			sinceQuery = `$2 = $2`
		}
	} else {
		if since != "" {
			sinceQuery = `"user"."nickname" > $2`
		} else {
			sinceQuery = `$2 = $2`
		}
	}

	row, er := tx.Query(`
		select 
			"user"."about" as about,
			"user"."email" as email,
			"user"."fullname" as fullname,
			"user"."nickname" as nickname
		from thread
		left join post on "post"."t_id" = thread.id 
		join "user" on ("user"."id" = post.u_id or "user"."id" = thread.u_id)
		where thread.f_id = $1
		and `+sinceQuery+`
		group by "user"."about", "user"."email", "user"."fullname", "user"."nickname"
		order by "user"."nickname" `+descQuery+`
		limit $3;`, fid, since, limit)

	if er != nil {
		// log.Println(er)
	} else {
		for row.Next() {
			u := User{}
			row.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
			*useres = append(*useres, &u)
		}
	}

	row.Close()
	tx.Commit()
	return err
}
