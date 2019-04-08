package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/K1ola/tp_db_forum/database"
	"github.com/K1ola/tp_db_forum/helpers"
	"github.com/K1ola/tp_db_forum/models"
	"github.com/valyala/fasthttp"
)

func ForumCreate(ctx *fasthttp.RequestCtx) {
	forum := models.Forum{}
	err := json.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", forum.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchUsers := helpers.RowsToUsers(rows)

	if len(matchUsers) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find user with nickname: "+forum.Nickname), fasthttp.StatusNotFound)
		return
	}

	rows, err = database.Query("SELECT * FROM forum WHERE LOWER(slug)=LOWER($1) OR LOWER(nickname)=LOWER($2)", forum.Slug, forum.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchForums := helpers.RowsToForums(rows)

	if len(matchForums) != 0 {

		j, _ := json.Marshal(matchForums[0])
		helpers.ResponseCtx(ctx, string(j), fasthttp.StatusConflict)
		return
	}

	_, err = database.Exec("INSERT INTO forum (slug, title, nickname) VALUES ($1, $2, (SELECT nickname FROM users WHERE LOWER(users.nickname)=LOWER($3)))", forum.Slug, forum.Title, matchUsers[0].Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	forum.Nickname = matchUsers[0].Nickname
	j, _ := json.Marshal(forum)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusCreated)
}

func ForumGetOne(ctx *fasthttp.RequestCtx) {
	forum := models.Forum{}
	forum.Slug = ctx.UserValue("slug").(string)

	rows, err := database.Query("SELECT * FROM forum WHERE LOWER(slug)=LOWER($1)", forum.Slug)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchForums := helpers.RowsToForums(rows)

	if len(matchForums) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find forum with slug: "+forum.Slug), fasthttp.StatusNotFound)
		return
	}

	j, _ := json.Marshal(matchForums[0])
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func ForumGetThreads(ctx *fasthttp.RequestCtx) {
	thread := models.Thread{}
	thread.Forum = ctx.UserValue("slug").(string)

	sql := "SELECT * FROM thread WHERE LOWER(forum)=LOWER($1)"
	rows, err := database.Query(sql, thread.Forum)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()
	totalThreads := helpers.RowsToThreads(rows)

	if len(totalThreads) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find forum by slug: "+thread.Forum), fasthttp.StatusNotFound)
		return
	}

	since := string(ctx.QueryArgs().Peek("since"))
	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")

	if since != "" {
		if desc {
			sql += " AND created<='" + since + "'"
		} else {
			sql += " AND created>='" + since + "'"
		}
	}
	if desc {
		sql += " ORDER BY created DESC"
	} else {
		sql += " ORDER BY created ASC"
	}
	if limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}

	rows, _ = database.Query(sql, thread.Forum)
	defer rows.Close()
	matchThreads := helpers.RowsToThreads(rows)

	j, _ := json.Marshal(matchThreads)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func ForumGetUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	since := string(ctx.QueryArgs().Peek("since"))
	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")

	rows, err := database.Query("SELECT * FROM forum WHERE LOWER(slug)=LOWER($1)", slug)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if len(helpers.RowsToForums(rows)) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find forum by slug: "+slug), fasthttp.StatusNotFound)
		return
	}

	sql := `SELECT DISTINCT users.* FROM users 
		LEFT JOIN post ON Lower(users.nickname)=Lower(post.author)
		LEFT JOIN thread ON Lower(users.nickname)=Lower(thread.author)
		WHERE (Lower(post.forum)=Lower($1) OR Lower(thread.forum)=Lower($1))`

	if desc {
		if since != "" {
			sql += " AND users.nickname < '" + since + "'"
		}
		sql += " ORDER BY users.nickname DESC"
	} else {
		if since != "" {
			sql += " AND users.nickname > '" + since + "'"
		}
		sql += " ORDER BY users.nickname ASC"
	}
	if limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}

	rows, err = database.Query(sql, slug)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := helpers.RowsToUsers(rows)
	j, _ := json.Marshal(users)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}
