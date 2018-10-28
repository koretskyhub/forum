package handlers

import (
	"github.com/qiangxue/fasthttp-routing"
)

func ParseQueryParams(ctx *routing.Context) (limit int64, since string, desc bool) {
	limitInt := ctx.URI().QueryArgs().GetUintOrZero("limit")
	limit = int64(limitInt)
	since = string(ctx.URI().QueryArgs().Peek("since"))
	desc = ctx.URI().QueryArgs().GetBool("desc")
	if limit == 0 {
		limit = 100
	}

	return limit, since, desc
}

func CommonHandler(ctx *routing.Context) (er error) {
	ctx.Response.Header.Set("Content-type", "application/json")
	return er
}
