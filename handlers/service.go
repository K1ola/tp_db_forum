package handlers

import (
	"encoding/json"

	"github.com/K1ola/tp_db_forum/database"
	"github.com/K1ola/tp_db_forum/helpers"
	"github.com/valyala/fasthttp"
)

func Clear(ctx *fasthttp.RequestCtx) {
	err := database.LoadSchemaSQL()
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
	}
}

func Status(ctx *fasthttp.RequestCtx) {
	type Responce struct {
		Forum  int32 `json:"forum"`
		Post   int64 `json:"post"`
		Thread int32 `json:"thread"`
		User   int32 `json:"user"`
	}
	responce := Responce{}
	database.QueryRow("SELECT Count(*) FROM forum").Scan(&responce.Forum)
	database.QueryRow("SELECT Count(*) FROM post").Scan(&responce.Post)
	database.QueryRow("SELECT Count(*) FROM thread").Scan(&responce.Thread)
	database.QueryRow("SELECT Count(*) FROM users").Scan(&responce.User)
	j, _ := json.Marshal(responce)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}
