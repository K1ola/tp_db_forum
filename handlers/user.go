package handlers

import (
	"encoding/json"

	"github.com/K1ola/tp_db_forum/database"
	"github.com/K1ola/tp_db_forum/helpers"
	"github.com/K1ola/tp_db_forum/models"
	"github.com/valyala/fasthttp"
)

func UserCreate(ctx *fasthttp.RequestCtx) {
	user := models.User{}
	user.Nickname = ctx.UserValue("nickname").(string)
	err := json.Unmarshal(ctx.PostBody(), &user)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1) OR LOWER(email)=LOWER($2)", user.Nickname, user.Email)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchUsers := helpers.RowsToUsers(rows)

	if len(matchUsers) != 0 {
		j, _ := json.Marshal(matchUsers)
		helpers.ResponseCtx(ctx, string(j), fasthttp.StatusConflict)
		return
	}

	_, err = database.Exec("INSERT INTO users (about, email, fullname, nickname) VALUES ($1, $2, $3, $4)", user.About, user.Email, user.Fullname, user.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(user)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusCreated)
	return
}

func UserGetOne(ctx *fasthttp.RequestCtx) {
	user := models.User{}
	user.Nickname = ctx.UserValue("nickname").(string)

	rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", user.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchUsers := helpers.RowsToUsers(rows)

	if len(matchUsers) == 1 {
		j, _ := json.Marshal(matchUsers[0])
		helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
	} else {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find user by nickname: "+user.Nickname), fasthttp.StatusNotFound)
	}
	return
}

func UserUpdate(ctx *fasthttp.RequestCtx) {
	user := models.User{}
	user.Nickname = ctx.UserValue("nickname").(string)
	err := json.Unmarshal(ctx.PostBody(), &user)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", user.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchUsers := helpers.RowsToUsers(rows)

	if len(matchUsers) == 0 {

		helpers.ResponseCtx(ctx, helpers.Msg("Can't find user by nickname: "+user.Nickname), fasthttp.StatusNotFound)

	} else {

		if user.Email != "" && user.Email != matchUsers[0].Email {
			rows, err = database.Query("SELECT * FROM users WHERE LOWER(email)=LOWER($1)", user.Email)
			otherUsers := helpers.RowsToUsers(rows)
			if len(otherUsers) != 0 {
				helpers.ResponseCtx(ctx, helpers.Msg("This email is already registered by user: "+otherUsers[0].Nickname), fasthttp.StatusConflict)
				return
			}
		}

		if user.Email == "" && user.Fullname == "" && user.About == "" {
			j, _ := json.Marshal(matchUsers[0])
			helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
			return
		}

		if user.About == "" {
			user.About = matchUsers[0].About
		}
		if user.Email == "" {
			user.Email = matchUsers[0].Email
		}
		if user.Fullname == "" {
			user.Fullname = matchUsers[0].Fullname
		}
		_, err = database.Exec("UPDATE users SET about=$1, email=$2, fullname=$3 WHERE nickname=$4", user.About, user.Email, user.Fullname, user.Nickname)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		j, _ := json.Marshal(user)
		helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
	}
	return
}
