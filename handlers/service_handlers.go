package handlers

import (
	"forum/models"
	"log"
	"net/http"

	"github.com/mailru/easyjson"
	routing "github.com/qiangxue/fasthttp-routing"
)

func ClearDbHandler(ctx *routing.Context) (er error) {
	log.Println("clearDbHandler")
	status := models.Status{}
	err := status.ResetData()
	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusInternalServerError)
	default:
		ctx.SetStatusCode(http.StatusOK)
	}
	return er
}

func GetDbInfoHandler(ctx *routing.Context) (er error) {
	log.Println("getDbInfoHandler")
	status := models.Status{}
	err := status.GetStatus()
	switch err.Message {
	case models.NotFound:
		ctx.SetStatusCode(http.StatusInternalServerError)
	default:
		ctx.SetStatusCode(http.StatusOK)
		if _, jsonEr := easyjson.MarshalToWriter(status, ctx); jsonEr != nil {
			return jsonEr
		}
	}
	return er
}
