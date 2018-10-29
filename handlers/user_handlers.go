package handlers

import (
	"forum/models"
	"net/http"

	"github.com/mailru/easyjson"
	routing "github.com/qiangxue/fasthttp-routing"
)

func CreateUserHandler(ctx *routing.Context) (er error) {
	// log.Println("createUserHandler")
	user := models.User{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &user); jsonEr != nil {
		return jsonEr
	}
	user.Nickname = ctx.Param("nickname")
	users, err := user.Create()
	if err.Message == models.NotFound {
		ctx.SetStatusCode(http.StatusConflict)
		if _, jsonEr := easyjson.MarshalToWriter(*users, ctx); jsonEr != nil {
			return jsonEr
		}
	} else {
		ctx.SetStatusCode(http.StatusCreated)
		if _, jsonEr := easyjson.MarshalToWriter(user, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func GetProfileHandler(ctx *routing.Context) (er error) {
	// log.Println("getProfileHandler")
	user := models.User{}
	user.Nickname = ctx.Param("nickname")
	err := user.GetProfile()

	if err.Message == models.NotFound {
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	} else {
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(user, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func UpdateProfileHandler(ctx *routing.Context) (er error) {
	// log.Println("updateProfileHandler")
	user := models.User{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &user); jsonEr != nil {
		return jsonEr
	}
	user.Nickname = ctx.Param("nickname")

	err := user.UpdateProfile()

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	case models.Conflict:
		ctx.SetStatusCode(http.StatusConflict)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(user, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func GetForumUsersHandler(ctx *routing.Context) (er error) {
	// log.Println("getForumUsersHandler")

	limit, since, desc := ParseQueryParams(ctx)

	forum := models.Forum{}

	forum.Slug = ctx.Param("slug")

	users := models.Users{}

	err := users.GetByForum(&forum, limit, since, desc)

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(users, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}
