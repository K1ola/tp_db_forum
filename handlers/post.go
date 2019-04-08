package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/K1ola/tp_db_forum/database"
	"github.com/K1ola/tp_db_forum/helpers"
	"github.com/K1ola/tp_db_forum/models"
	"github.com/valyala/fasthttp"
)

func PostsCreate(ctx *fasthttp.RequestCtx) {
	now := time.Now().Format(time.RFC3339)
	posts := []models.Post{}
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	matchThread, err := helpers.GetThreadBySlugOrID(ctx)
	if err != nil {
		return
	}

	for i := range posts {
		posts[i].Created = now
		posts[i].Forum = matchThread.Forum
		posts[i].Thread = matchThread.ID

		if posts[i].Parent != 0 {
			var parentThread int32 = 0
			database.QueryRow("SELECT thread FROM post WHERE id=$1", posts[i].Parent).Scan(&parentThread)
			if parentThread != posts[i].Thread {
				helpers.ResponseCtx(ctx, helpers.Msg("Parent post was created in another thread"), fasthttp.StatusConflict)
				return
			}
		}

		rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", posts[i].Author)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		defer rows.Close()
		if len(helpers.RowsToUsers(rows)) == 0 {
			helpers.ResponseCtx(ctx, helpers.Msg("Can't find post author by nickname: "+posts[i].Author), fasthttp.StatusNotFound)
			return
		}

		_, err = database.Exec("INSERT INTO post (author, created, forum, message, parent, thread) VALUES ($1, $2, $3, $4, $5, $6)",
			posts[i].Author, posts[i].Created, posts[i].Forum, posts[i].Message, posts[i].Parent, posts[i].Thread)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		database.QueryRow("SELECT id FROM post ORDER BY id DESC LIMIT 1").Scan(&posts[i].ID)
		var postCounter int64
		database.QueryRow("SELECT posts FROM forum WHERE LOWER(slug)=LOWER($1)", posts[i].Forum).Scan(&postCounter)
		_, err = database.Exec("UPDATE forum SET posts=$1 WHERE LOWER(slug)=LOWER($2)", postCounter+1, posts[i].Forum)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
	}

	j, err := json.Marshal(posts)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusCreated)
}

func PostGetOne(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)

	rows, err := database.Query("SELECT * FROM post WHERE id=$1", id)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := helpers.RowsToPosts(rows)

	if len(posts) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find post by id: "+id), fasthttp.StatusNotFound)
		return
	}

	type Responce struct {
		Author *models.User   `json:"author,omitempty"`
		Forum  *models.Forum  `json:"forum,omitempty"`
		Post   *models.Post   `json:"post"`
		Thread *models.Thread `json:"thread,omitempty"`
	}
	responce := Responce{}
	responce.Post = &posts[0]

	related := string(ctx.QueryArgs().Peek("related"))
	if strings.Index(related, "user") >= 0 {
		rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", responce.Post.Author)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		defer rows.Close()
		matchUsers := helpers.RowsToUsers(rows)
		if len(matchUsers) != 0 {
			responce.Author = &matchUsers[0]
		}
	}
	if strings.Index(related, "thread") >= 0 {
		rows, err := database.Query("SELECT * FROM thread WHERE id=$1", responce.Post.Thread)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		defer rows.Close()
		matchThreads := helpers.RowsToThreads(rows)
		if len(matchThreads) != 0 {
			responce.Thread = &matchThreads[0]
		}
	}
	if strings.Index(related, "forum") >= 0 {
		rows, err := database.Query("SELECT * FROM forum WHERE Lower(slug)=Lower($1)", responce.Post.Forum)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		defer rows.Close()
		matchForums := helpers.RowsToForums(rows)
		if len(matchForums) != 0 {
			responce.Forum = &matchForums[0]
		}
	}

	j, err := json.Marshal(responce)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func PostUpdate(ctx *fasthttp.RequestCtx) {
	post := models.Post{}
	id := ctx.UserValue("id").(string)
	err := json.Unmarshal(ctx.PostBody(), &post)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM post WHERE id=$1", id)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchPosts := helpers.RowsToPosts(rows)

	if len(matchPosts) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find post with id: "+id), fasthttp.StatusNotFound)
		return
	}

	if post.Message != "" && post.Message != matchPosts[0].Message {
		matchPosts[0].Message = post.Message
		matchPosts[0].IsEdited = true
		_, err = database.Exec("UPDATE post SET message=$1, isEdited=$2 WHERE id=$3",
			matchPosts[0].Message, matchPosts[0].IsEdited, id)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
	}

	j, _ := json.Marshal(matchPosts[0])
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}
