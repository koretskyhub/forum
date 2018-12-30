package handlers

import (
	"forum/models"
	"net/http"
	"strconv"

	"github.com/mailru/easyjson"
	routing "github.com/qiangxue/fasthttp-routing"
)

func CreateThreadHandler(ctx *routing.Context) (er error) {
	// log.Println("createThreadHandler")
	thread := models.Thread{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &thread); jsonEr != nil {
		return jsonEr
	}

	thread.Forum = ctx.Param("slug")

	err := thread.Create()
	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	case models.Conflict:
		ctx.SetStatusCode(http.StatusConflict)
		if _, jsonEr := easyjson.MarshalToWriter(thread, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusCreated)
		if _, jsonEr := easyjson.MarshalToWriter(thread, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func GetForumThreadsHandler(ctx *routing.Context) (er error) {
	// log.Println("getForumThreadsHandler")
	limit, since, desc := ParseQueryParams(ctx)
	forumSlug := ctx.Param("slug")
	threads := models.Threads{}

	// log.Println(forumSlug, limit, since, desc)
	err := threads.GetThreadsByForum(forumSlug, limit, since, desc)

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(threads, ctx); jsonEr != nil {
			return jsonEr
		}
	}

	return er
}

func GetThreadHandler(ctx *routing.Context) (er error) {
	// log.Println("getThreadHandler")

	thread := models.Thread{}

	if id, er := strconv.Atoi(ctx.Param("slug_or_id")); er != nil {
		thread.Slug = ctx.Param("slug_or_id")
	} else {
		thread.Id = int64(id)
	}

	err := thread.Get()

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(thread, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func UpdateThreadHandler(ctx *routing.Context) (er error) {
	// log.Println("updateThreadHandler")
	thread := models.Thread{}
	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &thread); jsonEr != nil {
		return jsonEr
	}

	if id, er := strconv.Atoi(ctx.Param("slug_or_id")); er != nil {
		thread.Slug = ctx.Param("slug_or_id")
	} else {
		thread.Id = int64(id)
	}

	err := thread.Update()

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(thread, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}

func VoteThreadHandler(ctx *routing.Context) (er error) {
	// log.Println("voteThreadHandler")

	vote := models.Vote{}
	thread := models.Thread{}

	if jsonEr := easyjson.Unmarshal(ctx.PostBody(), &vote); jsonEr != nil {
		return jsonEr
	}

	if id, er := strconv.Atoi(ctx.Param("slug_or_id")); er != nil {
		thread.Slug = ctx.Param("slug_or_id")
	} else {
		thread.Id = int64(id)
	}

	err := vote.AddVoice(&thread)

	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusNotFound)
		if _, jsonEr := easyjson.MarshalToWriter(err, ctx); jsonEr != nil {
			return jsonEr
		}
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(thread, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}
