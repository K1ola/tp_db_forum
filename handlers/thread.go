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

func ThreadCreate(ctx *fasthttp.RequestCtx) {
	if strings.ToLower(ctx.UserValue("slug").(string)) == "create" {
		ForumCreate(ctx)
		return
	}

	thread := models.Thread{}
	thread.Forum = ctx.UserValue("slug").(string)
	err := json.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM forum WHERE LOWER(slug)=LOWER($1)", thread.Forum)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchForums := helpers.RowsToForums(rows)

	if len(matchForums) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find forum by slug: "+thread.Forum), fasthttp.StatusNotFound)
		return
	}

	rows, _ = database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", thread.Author)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matchUsers := helpers.RowsToUsers(rows)

	if len(matchUsers) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find thread author by nickname: "+thread.Author), fasthttp.StatusNotFound)
		return
	}

	if thread.Slug != "" {
		rows, err = database.Query("SELECT * FROM thread WHERE LOWER(slug)=LOWER($1)", thread.Slug)
		if err != nil {
			helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		defer rows.Close()
		matchThreads := helpers.RowsToThreads(rows)
		if len(matchThreads) != 0 {
			j, _ := json.Marshal(matchThreads[0])
			helpers.ResponseCtx(ctx, string(j), fasthttp.StatusConflict)
			return
		}
	}

	matchForums[0].Threads++
	thread.Forum = matchForums[0].Slug
	if thread.Created == "" {
		thread.Created = time.Now().Format(time.RFC3339Nano)
	}

	_, err = database.Exec("INSERT INTO thread (author, created, forum, message, slug, title) VALUES ($1, $2, $3, $4, $5, $6)",
		thread.Author, thread.Created, thread.Forum, thread.Message, thread.Slug, thread.Title)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	database.QueryRow("SELECT id FROM thread ORDER BY id DESC LIMIT 1").Scan(&thread.ID)
	_, err = database.Exec("UPDATE forum SET threads=$1 WHERE LOWER(slug)=LOWER($2)", matchForums[0].Threads, matchForums[0].Slug)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(thread)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusCreated)
}

func ThreadGetOne(ctx *fasthttp.RequestCtx) {
	matchThread, err := helpers.GetThreadBySlugOrID(ctx)
	if err != nil {
		return
	}
	j, _ := json.Marshal(matchThread)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func cutPostsBySince(posts []models.Post, since int) []models.Post {
	cut := 0
	for i := range posts {
		if posts[i].ID == int64(since) {
			cut = i + 1
			break
		}
	}
	if cut > 0 {
		if cut == len(posts) {
			posts = []models.Post{}
		} else {
			posts = posts[cut:]
		}
	}
	return posts
}

func getPostsFlat(forum string, since int, desc bool, limit int) ([]models.Post, error) {
	sql := "SELECT * FROM post WHERE LOWER(forum)=LOWER($1)"
	if desc {
		sql += " ORDER BY id DESC"
	} else {
		sql += " ORDER BY id ASC"
	}

	rows, err := database.Query(sql, forum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := helpers.RowsToPosts(rows)
	if since > 0 {
		posts = cutPostsBySince(posts, since)
	}
	if limit > 0 && limit < len(posts) {
		posts = posts[:limit]
	}

	return posts, nil
}

func getPostsTree(forum string, since int, desc bool, limit int) ([]models.Post, error) {

	rootsql := "SELECT *, ARRAY[id] AS level FROM post WHERE LOWER(forum)=LOWER($1) AND parent=0"
	recusql := "SELECT post.*, tree.level||post.id FROM post JOIN tree ON tree.id=post.parent"
	sql := `WITH RECURSIVE tree AS (` + rootsql + ` UNION ALL ` + recusql + `) 
			SELECT author, created, forum, id, isEdited, message, parent, thread FROM tree`
	if desc {
		sql += " ORDER BY level DESC"
	} else {
		sql += " ORDER BY level ASC"
	}

	rows, err := database.Query(sql, forum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := helpers.RowsToPosts(rows)
	if since > 0 {
		posts = cutPostsBySince(posts, since)
	}
	if limit > 0 && limit < len(posts) {
		posts = posts[:limit]
	}

	return posts, nil
}

func getPostsParentTree(forum string, since int, desc bool, limit int) ([]models.Post, error) {

	sql := "SELECT * FROM post WHERE LOWER(forum)=LOWER($1) AND parent=0"
	if desc {
		sql += " ORDER BY id DESC"
	} else {
		sql += " ORDER BY id ASC"
	}

	rows, err := database.Query(sql, forum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rootposts := helpers.RowsToPosts(rows)
	posts := []models.Post{}

	nodesql := "SELECT *, ARRAY[id] AS level FROM post WHERE LOWER(forum)=LOWER($1) AND parent=$2"
	recusql := "SELECT post.*, tree.level||post.id FROM post JOIN tree ON tree.id=post.parent"
	sql = `WITH RECURSIVE tree AS (` + nodesql + ` UNION ALL ` + recusql + `) 
			SELECT author, created, forum, id, isEdited, message, parent, thread FROM tree ORDER BY level`

	for i := range rootposts {
		posts = append(posts, rootposts[i])
		rows, err = database.Query(sql, forum, rootposts[i].ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		posts = append(posts, helpers.RowsToPosts(rows)...)
	}

	if since > 0 {
		posts = cutPostsBySince(posts, since)
	}
	if limit > 0 {
		truelimit := 0
		for i := range posts {
			if posts[i].Parent == 0 {
				limit--
				if limit < 0 {
					truelimit = i
					break
				}
			}
		}
		if truelimit > 0 {
			posts = posts[:truelimit]
		}
	}

	return posts, nil
}

func ThreadGetPosts(ctx *fasthttp.RequestCtx) {
	since := ctx.QueryArgs().GetUintOrZero("since")
	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")
	sort := string(ctx.QueryArgs().Peek("sort"))

	matchThread, err := helpers.GetThreadBySlugOrID(ctx)
	if err != nil {
		return
	}

	matchPosts := []models.Post{}
	switch sort {
	case "tree":
		matchPosts, err = getPostsTree(matchThread.Forum, since, desc, limit)
	case "parent_tree":
		matchPosts, err = getPostsParentTree(matchThread.Forum, since, desc, limit)
	default: //take as "flat" by default
		matchPosts, err = getPostsFlat(matchThread.Forum, since, desc, limit)
	}
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(matchPosts)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func ThreadUpdate(ctx *fasthttp.RequestCtx) {
	thread := models.Thread{}
	err := json.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	matchThread, err := helpers.GetThreadBySlugOrID(ctx)
	if err != nil {
		return
	}
	if thread.Message != "" {
		matchThread.Message = thread.Message
	}
	if thread.Title != "" {
		matchThread.Title = thread.Title
	}

	_, err = database.Exec("UPDATE thread SET message=$1, title=$2 WHERE id=$3",
		matchThread.Message, matchThread.Title, matchThread.ID)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(matchThread)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}

func ThreadVote(ctx *fasthttp.RequestCtx) {
	vote := models.Vote{}
	err := json.Unmarshal(ctx.PostBody(), &vote)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	rows, err := database.Query("SELECT * FROM users WHERE LOWER(nickname)=LOWER($1)", vote.Nickname)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer rows.Close()
	if len(helpers.RowsToUsers(rows)) == 0 {
		helpers.ResponseCtx(ctx, helpers.Msg("Can't find user by nickname: "+vote.Nickname), fasthttp.StatusNotFound)
		return
	}

	matchThread, err := helpers.GetThreadBySlugOrID(ctx)
	if err != nil {
		return
	}

	if vote.Voice >= 0 {
		matchThread.Votes += vote.Voice
	} else {
		matchThread.Votes = 1
	}
	vote.Thread = matchThread.Slug

	_, err = database.Exec("UPDATE thread SET votes=$1 WHERE slug=$2", matchThread.Votes, matchThread.Slug)
	if err != nil {
		helpers.ResponseCtx(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	j, _ := json.Marshal(matchThread)
	helpers.ResponseCtx(ctx, string(j), fasthttp.StatusOK)
}
