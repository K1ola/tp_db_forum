package helpers

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/K1ola/tp_db_forum/database"
	"github.com/K1ola/tp_db_forum/models"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

type answer struct {
	Message string `json:"message,omitempty"`
}

func Msg(message string) string {
	msg := answer{}
	msg.Message = message
	j, _ := json.Marshal(msg)
	return string(j)
}

func RowsToUsers(rows *pgx.Rows) []models.User {
	users := []models.User{}
	for rows.Next() {
		entry := models.User{}
		_ = rows.Scan(&entry.About, &entry.Email, &entry.Fullname, &entry.Nickname)
		users = append(users, entry)
	}
	return users
}

func RowsToForums(rows *pgx.Rows) []models.Forum {
	forums := []models.Forum{}
	for rows.Next() {
		entry := models.Forum{}
		_ = rows.Scan(&entry.Posts, &entry.Slug, &entry.Threads, &entry.Title, &entry.Nickname)
		forums = append(forums, entry)
	}
	return forums
}

func RowsToThreads(rows *pgx.Rows) []models.Thread {
	threads := []models.Thread{}
	for rows.Next() {
		entry := models.Thread{}
		var t time.Time
		_ = rows.Scan(&entry.Author, &t, &entry.Forum, &entry.ID, &entry.Message, &entry.Slug, &entry.Title, &entry.Votes)
		entry.Created = t.Format(time.RFC3339Nano)
		threads = append(threads, entry)
	}
	return threads
}

func RowsToPosts(rows *pgx.Rows) []models.Post {
	posts := []models.Post{}
	for rows.Next() {
		entry := models.Post{}
		var t time.Time
		_ = rows.Scan(&entry.Author, &t, &entry.Forum, &entry.ID, &entry.IsEdited, &entry.Message, &entry.Parent, &entry.Thread)
		entry.Created = t.Format(time.RFC3339Nano)
		posts = append(posts, entry)
	}
	return posts
}

func ResponseCtx(ctx *fasthttp.RequestCtx, content string, status int) {
	ctx.Write([]byte(content))
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(status)
	return
}

func GetThreadBySlugOrID(ctx *fasthttp.RequestCtx) (models.Thread, error) {
	slug_or_id := ctx.UserValue("slug_or_id").(string)
	rows, err := database.Query("SELECT * FROM thread WHERE LOWER(slug)=LOWER($1)", slug_or_id)
	if err != nil {
		ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return models.Thread{}, err
	}
	defer rows.Close()

	threads := RowsToThreads(rows)

	if len(threads) == 0 {
		id, err := strconv.Atoi(slug_or_id)
		if err == nil {
			rows, _ = database.Query("SELECT * FROM thread WHERE id=$1", id)
			defer rows.Close()
			threads = RowsToThreads(rows)
		}
	}
	if len(threads) == 0 {
		msg := "Can't find thread by slug or id: " + slug_or_id
		ResponseCtx(ctx, Msg(msg), fasthttp.StatusNotFound)
		return models.Thread{}, errors.New(msg)
	}

	return threads[0], nil
}
