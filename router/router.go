package router

import (
	"fmt"

	"github.com/K1ola/tp_db_forum/handlers"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc fasthttp.RequestHandler
}

type Routes []Route

func NewRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()
	for _, route := range routes {
		var handler fasthttp.RequestHandler
		handler = route.HandlerFunc
		router.Handle(route.Method,
			route.Pattern,
			handler)
	}

	return router
}

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hello World!")
}

var routes = Routes{
	Route{"GET", "/api/", Index},

	//conflict patterns (hhtprouter issue #175)
	//Route{"POST", "/api/forum/create",			handlers.ForumCreate,},

	Route{"POST", "/api/forum/:slug", handlers.ThreadCreate},
	Route{"POST", "/api/forum/:slug/create", handlers.ThreadCreate},
	Route{"GET", "/api/forum/:slug/details", handlers.ForumGetOne},
	Route{"GET", "/api/forum/:slug/threads", handlers.ForumGetThreads},
	Route{"GET", "/api/forum/:slug/users", handlers.ForumGetUsers},
	Route{"GET", "/api/post/:id/details", handlers.PostGetOne},
	Route{"POST", "/api/post/:id/details", handlers.PostUpdate},
	Route{"POST", "/api/thread/:slug_or_id/create", handlers.PostsCreate},
	Route{"GET", "/api/thread/:slug_or_id/details", handlers.ThreadGetOne},
	Route{"POST", "/api/thread/:slug_or_id/details", handlers.ThreadUpdate},
	Route{"GET", "/api/thread/:slug_or_id/posts", handlers.ThreadGetPosts},
	Route{"POST", "/api/thread/:slug_or_id/vote", handlers.ThreadVote},
	Route{"POST", "/api/user/:nickname/create", handlers.UserCreate},
	Route{"GET", "/api/user/:nickname/profile", handlers.UserGetOne},
	Route{"POST", "/api/user/:nickname/profile", handlers.UserUpdate},
	Route{"POST", "/api/service/clear", handlers.Clear},
	Route{"GET", "/api/service/status", handlers.Status},
}
