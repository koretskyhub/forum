package api

import (
	"forum/database"
	"forum/handlers"
	"log"

	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func setRoutes(router *routing.Router) {
	//handler for all routes
	router.Use(handlers.CommonHandler)
	//Forum handlers
	router.Post("/api/forum/create", handlers.CreateForumHandler)
	router.Post("/api/forum/<slug>/create", handlers.CreateThreadHandler)
	router.Get("/api/forum/<slug>/details", handlers.GetForumHandler)
	router.Get("/api/forum/<slug>/threads", handlers.GetForumThreadsHandler)
	router.Get("/api/forum/<slug>/users", handlers.GetForumUsersHandler)

	//Post handlers
	router.Get("/api/post/<id:\\d+>/details", handlers.GetPostHandler)
	router.Post("/api/post/<id:\\d+>/details", handlers.UpdatePostHandler)

	//Service handlers
	router.Post("/api/service/clear", handlers.ClearDbHandler)
	router.Get("/api/service/status", handlers.GetDbInfoHandler)

	//Thread handlers
	router.Post("/api/thread/<slug_or_id>/create", handlers.CreatePostsHandler)
	router.Get("/api/thread/<slug_or_id>/details", handlers.GetThreadHandler)
	router.Post("/api/thread/<slug_or_id>/details", handlers.UpdateThreadHandler)
	router.Get("/api/thread/<slug_or_id>/posts", handlers.GetThreadPostsHandler)
	router.Post("/api/thread/<slug_or_id>/vote", handlers.VoteThreadHandler)

	//User handlers
	router.Post("/api/user/<nickname>/create", handlers.CreateUserHandler)
	router.Get("/api/user/<nickname>/profile", handlers.GetProfileHandler)
	router.Post("/api/user/<nickname>/profile", handlers.UpdateProfileHandler)
}

func StartServer() (err error) {
	router := routing.New()
	setRoutes(router)
	database.InitDb()
	log.Println("Server listening at localhost:5002")
	if err := fasthttp.ListenAndServe(":5002", router.HandleRequest); err != nil {
		log.Println(err)
	}
	return err

}
