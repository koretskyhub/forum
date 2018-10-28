package handlers

import (
	"forum/models"
	"log"
	"net/http"

	"github.com/mailru/easyjson"
	routing "github.com/qiangxue/fasthttp-routing"
)

func CreateForumHandler(ctx *routing.Context) (er error) {
	log.Println("createForumHandler")
	forum := models.Forum{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &forum); jsonEr != nil {
		return jsonEr
	}
	err := forum.Create()
	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	case models.Conflict:
		ctx.SetStatusCode(http.StatusConflict)
		if _, jsonEr := easyjson.MarshalToWriter(forum, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusCreated)
		if _, jsonEr := easyjson.MarshalToWriter(forum, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func GetForumHandler(ctx *routing.Context) (er error) {
	log.Println("getForumHandler")
	
	forum := models.Forum{}
	forum.Slug = ctx.Param("slug")

	err := forum.GetDetails()
	
	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(forum, ctx); jsonEr != nil {
			return jsonEr
		}
	}

	return er
}
