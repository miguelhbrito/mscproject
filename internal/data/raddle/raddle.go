package raddle

import (
	"database/sql"
	"log"

	"github.com/doug-martin/goqu/v9"

	dbConn "github.com/miguelhbrito/mscproject/platform/db_connect"
	logz "github.com/rs/zerolog/log"
)

type RaddleInt interface {
	Save(Post, *log.Logger) error
	List() (Posts, error)
	GetIndexLastQ() (LastPost, error)
	GetLinkRaddle() (LinkRaddle, error)
	SaveLastP(postNumber int) error
}

type RaddlePostgres struct{}

func (rp RaddlePostgres) Save(p Post, log *log.Logger) error {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	var cms []Commentary

	post := Post{
		Title:       p.Post,
		Link:        p.Link,
		Post:        p.Post,
		Forum:       p.Forum,
		Votes:       p.Votes,
		User:        p.User,
		DataCreated: p.DataCreated,
	}

	for _, value := range p.Commentaries {
		commentary := Commentary{
			Commentary:  value.Commentary,
			User:        value.User,
			Votes:       value.Votes,
			DataCreated: value.DataCreated,
		}
		cms = append(cms, commentary)
	}

	var postVerify Post
	// `SELECT title FROM raddle_post WHERE title = $1`
	sqlStatementVerify, _, _ := goqu.From("raddle_post").
		Select("id", "title").Where(goqu.C("title").Eq(post.Title)).ToSQL()
	result := db.QueryRow(sqlStatementVerify)
	err = result.Scan(&postVerify.Id, &postVerify.Title)
	if err != nil {
		if err == sql.ErrNoRows {

			postId, err := rp.savePost(post, db, log)
			if err != nil {
				log.Println("Error on savePost:", err)
				return err
			}

			for _, commentary := range cms {
				if err := rp.saveCommentary(postId, commentary, db, log); err != nil {
					log.Println("Error on saveCommentary:", err)
					return err
				}
			}
			log.Println("New post on Raddle, post and comments was successfully saved into db!")
			return nil
		}
	}

	log.Println("Raddle post already exist in DB!")
	log.Println("Check in new comments!")

	var commentsVerify []Commentary
	// `SELECT * FROM raddle_commentary WHERE post_id = $1`
	sqlCmsVerify, _, _ := goqu.From("raddle_commentary").Where(goqu.C("post_id").Eq(post.Id)).ToSQL()
	rows, err := db.Query(sqlCmsVerify)
	if err != nil {
		log.Println("Error getting the comments from db, err:", err)
		return err
	}
	for rows.Next() {

		var cmsVerify Commentary
		err := rows.Scan(&cmsVerify.Id, &cmsVerify.PostId, &cmsVerify.Commentary, &cmsVerify.User, &cmsVerify.Votes, &cmsVerify.DataCreated)
		if err != nil {
			log.Println("Error getting scanning from commentary, err:", err)
			return err
		}
		commentsVerify = append(commentsVerify, cmsVerify)
	}

	if len(cms) <= 0 {
		log.Println("There is no new comments in this question!")
		return nil
	}
	if len(cms) == len(postVerify.Commentaries) {
		log.Println("No new comments for this post!")
		return nil
	}

	commentaryMap := make(map[int]Commentary)
	for i, c := range commentsVerify {
		commentaryMap[i] = c
	}

	for idx, commentExisted := range cms {
		// If the key exists
		if commentValue, ok := commentaryMap[idx]; ok {

			// Double checking if have the same key/comment
			if commentValue.Commentary == commentExisted.Commentary {
				log.Println("No new commentary from old posts!")
				return nil
			}

			log.Println("Inserting new commentary in an existing post:", postVerify.Id, commentExisted)
			if err := rp.saveCommentary(postVerify.Id, commentExisted, db, log); err != nil {
				log.Println("Error on saveCommentary of a previous post into db, err: ", err)
				return err
			}

			log.Println("New commentary from old post from answer saved on db !")
		} else {

			log.Println("Inserting new commentary in an existing post:", postVerify.Id, commentExisted)
			if err := rp.saveCommentary(postVerify.Id, commentExisted, db, log); err != nil {
				log.Println("Error on saveCommentary of a previous post into db, err: ", err)
				return err
			}
		}
	}

	log.Println("Successfully saved Raddle post and comments into db!")
	return nil
}

func (rp RaddlePostgres) savePost(ps Post, db *sql.DB, log *log.Logger) (int, error) {
	var postId int
	log.Println("Inserting new Raddle post into DB")
	sqlStatement :=
		`INSERT INTO "raddle_post" (title, link, post, forum, votes, author, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.QueryRow(sqlStatement, ps.Title, ps.Link, ps.Post, ps.Forum, ps.Votes, ps.User, ps.DataCreated).Scan(&postId)
	if err != nil {
		log.Println("Error to insert a new raddle post into db, err: ", err)
		return 0, err
	}
	log.Println("Post id from post saved on db ", postId)

	return postId, nil
}

func (rp RaddlePostgres) saveCommentary(postId int, comment Commentary, db *sql.DB, log *log.Logger) error {
	sqlStatement :=
		`INSERT INTO raddle_commentary (post_id, commentary, author, votes, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStatement, postId, comment.Commentary, comment.User, comment.Votes, comment.DataCreated)
	if err != nil {
		log.Println("Error to insert a raddle commentary into db, err: ", err)
		return err
	}

	return nil
}

func (rp RaddlePostgres) List() (Posts, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return nil, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	var ps []Post

	sqlStatement := `SELECT id, title, link, post, forum, votes, author, created_at FROM raddle_post`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		logz.Error().Err(err).Msg("Error to get all raddle_post from db")
		return nil, err
	}

	for rows.Next() {
		var p Post
		err := rows.Scan(&p.Id, &p.Title, &p.Link, &p.Post, &p.Forum, &p.Votes, &p.User, &p.DataCreated)
		if err != nil {
			logz.Error().Err(err).Msg("Error to extract result from row")
		}

		sqlStatementComments := `SELECT id, post_id, commentary, author, votes, created_at FROM raddle_commentary WHERE post_id = $1`
		commentaryRows, err := db.Query(sqlStatementComments, p.Id)
		if err != nil {
			logz.Error().Err(err).Msg("Error to get all raddle_commentary from db")
			return nil, err
		}

		for commentaryRows.Next() {
			var c Commentary
			err := commentaryRows.Scan(&c.Id, &c.PostId, &c.Commentary, &c.User, &c.Votes, &c.DataCreated)
			if err != nil {
				logz.Error().Err(err).Msg("Error to extract result from commentaries rows")
				return nil, err
			}
			p.Commentaries = append(p.Commentaries, c)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func (rp RaddlePostgres) GetIndexLastQ() (LastPost, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return LastPost{}, nil
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idRaddle := "raddle"

	var lp LastPost
	// SELECT id, question_number FROM last_question WHERE id = $1
	sqlStatement, _, _ := goqu.From("last_question").Select("id", "question_number").Where(goqu.C("id").Eq(idRaddle)).ToSQL()
	result := db.QueryRow(sqlStatement)
	err = result.Scan(&lp.Id, &lp.PostNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			logz.Error().Err(err).Msg("Error, no rows in result")
			return LastPost{}, err
		}
		logz.Error().Err(err).Msg("Error to extract result from row")
		return LastPost{}, err
	}

	return lp, nil
}

func (rp RaddlePostgres) GetLinkRaddle() (LinkRaddle, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return LinkRaddle{}, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idRaddle := "raddle"

	var lk LinkRaddle
	//SELECT id, link FROM link_scraper WHERE id = $1
	sqlStatement, _, _ := goqu.From("link_scraper").Select("id", "link").Where(goqu.C("id").Eq(idRaddle)).ToSQL()
	result := db.QueryRow(sqlStatement)
	err = result.Scan(&lk.Id, &lk.Link)
	if err != nil {
		if err == sql.ErrNoRows {
			logz.Error().Err(err).Msg("Error, no rows in result")
			return LinkRaddle{}, err
		}
		logz.Error().Err(err).Msg("Error to extract result from row")
		return LinkRaddle{}, err
	}

	return lk, nil
}

func (rp RaddlePostgres) SaveLastP(postNumber int) error {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idRaddle := "raddle"

	sqlStatement := `UPDATE last_question SET question_number = $2 WHERE id = $1`
	_, err = db.Exec(sqlStatement, idRaddle, postNumber)
	if err != nil {
		logz.Error().Err(err).Msgf("Error to update last post into db")
		return err
	}
	return nil
}
