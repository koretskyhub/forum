package handlers

import (
	"bytes"
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/mailru/easyjson"
	routing "github.com/qiangxue/fasthttp-routing"
)

func GetThreadPostsHandler(ctx *routing.Context) (er error) {
	// log.Println("getThreadPostsHandler")
	limit, since, desc := ParseQueryParams(ctx)
	sort := string(ctx.QueryArgs().Peek("sort"))

	thread := models.Thread{}

	if id, er := strconv.Atoi(ctx.Param("slug_or_id")); er != nil {
		thread.Slug = ctx.Param("slug_or_id")
	} else {
		thread.Id = int64(id)
	}

	posts := models.Posts{}

	err := posts.GetPostsByThread(&thread, sort, limit, since, desc)

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(posts, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func GetPostHandler(ctx *routing.Context) (er error) {
	// log.Println("getPostHandler")


	//string[array]
	// related := string(ctx.URI().QueryArgs().Peek("related"))
	// relatedBites := ctx.QueryArgs().PeekMulti("related")
	// if len(relatedBites) != 0 {

	// }

	related := make([]string, 4, 4)

	if len(ctx.QueryArgs().PeekMulti("related")) != 0 {
		for _, rel := range bytes.Split(ctx.QueryArgs().PeekMulti("related")[0], []byte(",")) {
			related = append(related, string(rel))
		}
	}


	postFull := models.PostFull{Post: &models.Post{}}

	postFull.Post.Id, er = strconv.Atoi(ctx.Param("id"))
	if er != nil {
		log.Println(er)
	}
	
	err := postFull.Details(&related)

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(postFull, ctx); jsonEr != nil {
			return jsonEr
		}
	}

	return er
}

func UpdatePostHandler(ctx *routing.Context) (er error) {
	// log.Println("updatePostHandler")

	post := models.Post{}

	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &post); jsonEr != nil {
		return jsonEr
	}

	post.Id, er = strconv.Atoi(ctx.Param("id"))
	if er != nil {
		log.Println(er)
	}

	err := post.Update()

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(post, ctx); jsonEr != nil {
			return jsonEr
		}
	}

	return er
}

func CreatePostsHandler(ctx *routing.Context) (er error) {
	// log.Println("createPostHandler")
	posts := models.Posts{}
	thread := models.Thread{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &posts); jsonEr != nil {
		return jsonEr
	}
	if id, er := strconv.Atoi(ctx.Param("slug_or_id")); er != nil {
		thread.Slug = ctx.Param("slug_or_id")
	} else {
		thread.Id = int64(id)
	}

	// err := models.ModelError{}

	// if len(posts) > 0 {
	err := posts.Create(&thread)
	// }

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
		ctx.SetStatusCode(http.StatusCreated)
		if _, jsonEr := easyjson.MarshalToWriter(posts, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}
