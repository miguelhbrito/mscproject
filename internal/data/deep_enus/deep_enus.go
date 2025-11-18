package deepenus

import (
	"database/sql"
	"log"

	logz "github.com/rs/zerolog/log"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	dbConn "github.com/miguelhbrito/mscproject/platform/db_connect"
)

type DeepEnUSInt interface {
	Save(Question, *log.Logger) error
	List() (Questions, error)
	GetIndexLastQ() (LastQuestion, error)
	GetLinkDeepEnUS() (LinkDeepEnus, error)
	SaveLastQ(questionNumber int) error
}

type DeepEnusPostgres struct {
}

func (da DeepEnusPostgres) Save(q Question, log *log.Logger) error {
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

	var ans []Answer
	var cs []Comment

	qs := Question{
		Title:       q.Title,
		Question:    q.Question,
		Category:    q.Category,
		Tags:        q.Tags,
		Votes:       q.Votes,
		Author:      q.Author,
		Type:        q.Type,
		DataCreated: q.DataCreated,
		Points:      q.Points,
	}

	for _, value := range q.Answers {
		a := Answer{
			AnswerContent: value.AnswerContent,
			Votes:         value.Votes,
			Author:        value.Author,
			Type:          value.Type,
			DataCreated:   value.DataCreated,
			Points:        value.Points,
		}
		for _, comment := range value.Comments {
			c := Comment{
				Commentary:  comment.Commentary,
				Author:      comment.Author,
				Type:        comment.Type,
				DataCreated: comment.DataCreated,
			}
			cs = append(cs, c)
		}
		ans = append(ans, a)
	}

	var qsVerify Question
	// `SELECT title FROM deep_question_enus WHERE title = $1`
	sqlStatementVerify, _, _ := goqu.From("deep_question_enus").
		Select("id", "title").Where(goqu.C("title").Eq(qs.Title)).ToSQL()
	result := db.QueryRow(sqlStatementVerify)
	err = result.Scan(&qsVerify.Id, &qsVerify.Title)
	if err != nil {
		if err == sql.ErrNoRows {

			log.Println("Saving new question into db")
			questionId, err := da.saveQuestion(qs, db, log)
			if err != nil {
				log.Println("Error on saveQuestion:", err)
				return err
			}

			for _, answer := range ans {
				answerId, err := da.saveAnswer(questionId, answer, db, log)
				if err != nil {
					log.Println("Error on saveAnswer:", err)
					return err
				}
				for _, comment := range cs {
					if err := da.saveComment(answerId, comment, db, log); err != nil {
						log.Println("Error on saveComment:", err)
						return err
					}
				}
			}

			log.Println("New question EnUS, answer and comments was successfully saved into db!")
			return nil
		}
	}

	log.Println("DeepAnswers EnUS question already exist in DB!")
	log.Println("Check in new answers or comments!")

	var answersVerify []Answer
	// `SELECT * FROM deep_answer_enus WHERE question_id = $1`
	sqlAnsVerify, _, _ := goqu.From("deep_answer_enus").Where(goqu.C("question_id").Eq(qsVerify.Id)).ToSQL()
	rows, err := db.Query(sqlAnsVerify)
	if err != nil {
		log.Println("Error getting the answers from db, err:", err)
		return err
	}
	for rows.Next() {

		var awsVerify Answer
		err := rows.Scan(&awsVerify.Id, &awsVerify.QuestionId, &awsVerify.AnswerContent, &awsVerify.Votes, &awsVerify.Type, &awsVerify.Author, &awsVerify.DataCreated, &awsVerify.Points)
		if err != nil {
			log.Println("Error getting scanning from answer, err:", err)
			return err
		}
		answersVerify = append(answersVerify, awsVerify)
	}

	if len(ans) <= 0 {
		log.Println("There is no answers or comments in this question!")
		return nil
	}
	if len(ans) == len(qsVerify.Answers) {
		for idx, aswItem := range qsVerify.Answers {
			if len(aswItem.Comments) == len(ans[idx].Comments) {
				log.Println("No new answers and comments for this question!")
				return nil
			}
		}
	}

	asnwersMap := make(map[int]Answer)
	for i, a := range answersVerify {
		asnwersMap[i] = a
	}

	for idx, answerExisted := range ans {
		// If the key doesn't exists
		if answerValue, ok := asnwersMap[idx]; ok {

			if answerExisted.AnswerContent == answerValue.AnswerContent {

				log.Println("Old answer, check in new comments")
				var commentVerify []Comment
				// `SELECT * FROM deep_comment_enus WHERE answer_id = $1`
				sqlCommentVerify, _, _ := goqu.From("deep_comment_enus").Where(goqu.C("answer_id").Eq(answerValue.Id)).ToSQL()
				rows, err := db.Query(sqlCommentVerify)
				if err != nil {
					log.Println("Error getting the comments from db, err:", err)
					return err
				}
				for rows.Next() {

					var comVerify Comment
					err := rows.Scan(&comVerify.Id, &comVerify.AnswerId, &comVerify.Commentary, &comVerify.Type, &comVerify.Type, &comVerify.Author, &comVerify.DataCreated)
					if err != nil {
						log.Println("Error getting scanning from comments, err:", err)
						return err
					}
					commentVerify = append(commentVerify, comVerify)
				}

				commentsMap := make(map[int]Comment)
				for i, c := range commentVerify {
					commentsMap[i] = c
				}

				if err := da.saveCommentExistedQuestion(commentsMap, answerValue, answerExisted, db, log); err != nil {
					return err
				}
			}
		} else {
			log.Println("Inserting new answer in an existing question:", qsVerify.Id, answerExisted)

			ansIdExisted, err := da.saveAnswer(qsVerify.Id, answerExisted, db, log)
			if err != nil {
				log.Println("Error on saveAnswer of a previous question into db, err: ", err)
				return err
			}

			log.Println("New answer id from old question from answer saved on db ", ansIdExisted)
			for _, comment := range cs {
				if err := da.saveComment(ansIdExisted, comment, db, log); err != nil {
					log.Println("Error on saveComment of a previous question into db, err: ", err)
					return err
				}
			}
		}
	}

	log.Println("Successfully saved the question EnUS, answer and comments into db!")

	return nil
}

func (da DeepEnusPostgres) saveQuestion(qs Question, db *sql.DB, log *log.Logger) (int, error) {
	var questionId int
	log.Println("Inserting new DeepAnswers EnUS question into DB")
	sqlStatementQ :=
		`INSERT INTO "deep_question_enus" (title, question, category, tags, votes, author, type, created_at, points) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err := db.QueryRow(sqlStatementQ, qs.Title, qs.Question, qs.Category, pq.Array(qs.Tags), qs.Votes, qs.Author, qs.Type, qs.DataCreated, qs.Points).Scan(&questionId)
	if err != nil {
		log.Println("Error to insert a new deepAnswers question enus question into db, err: ", err)
		return 0, err
	}
	log.Println("Question id from question saved on db ", questionId)

	return questionId, nil
}

func (da DeepEnusPostgres) saveAnswer(qsId int, answer Answer, db *sql.DB, log *log.Logger) (int, error) {
	var answerId int
	sqlStatementA :=
		`INSERT INTO deep_answer_enus (question_id, answer_content, votes, author, type, created_at, points) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.QueryRow(sqlStatementA, qsId, answer.AnswerContent, answer.Votes, answer.Author, answer.Type, answer.DataCreated, answer.Points).Scan(&answerId)
	if err != nil {
		log.Println("Error to insert a deepAnswers enus answer into db, err: ", err)
		return 0, err
	}
	log.Println("Answer id from answer saved on db ", answerId)

	return answerId, nil
}

func (da DeepEnusPostgres) saveComment(ansId int, comment Comment, db *sql.DB, log *log.Logger) error {
	sqlStatementC :=
		`INSERT INTO deep_comment_enus (answer_id, commentary, author, type, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStatementC, ansId, comment.Commentary, comment.Author, comment.Type, comment.DataCreated)
	if err != nil {
		log.Println("Error to insert a deepAnswers enus commentary into db, err: ", err)
		return err
	}

	return nil
}

func (da DeepEnusPostgres) saveCommentExistedQuestion(
	commentsMap map[int]Comment,
	answerValue Answer,
	answerExisted Answer,
	db *sql.DB,
	log *log.Logger) error {

	for idxC, comment := range answerExisted.Comments {
		// If the key doesn't exists
		if commentValue, ok := commentsMap[idxC]; !ok {
			if comment.Commentary != commentValue.Commentary {
				log.Println("Inserting new comment in an existing question/answer!")

				if err := da.saveComment(answerValue.Id, comment, db, log); err != nil {
					log.Println("Error on saveComment of a previous question into db, err: ", err)
					return err
				}
			}
		}
	}
	return nil
}

func (da DeepEnusPostgres) List() (Questions, error) {
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

	var qs []Question

	// `SELECT id, title, question, category, tags, votes, author, type, created_at, points FROM deep_question_enus`
	sqlStatement, _, _ := goqu.From("deep_question_enus").
		Select("id", "title", "question", "category", "tags", "votes", "author", "type", "created_at", "points").ToSQL()
	rows, err := db.Query(sqlStatement)
	if err != nil {
		logz.Error().Err(err).Msg("Error to get all deep_question_enus from db")
		return nil, err
	}

	for rows.Next() {
		var q Question
		err := rows.Scan(
			&q.Id,
			&q.Title,
			&q.Question,
			&q.Category,
			pq.Array(&q.Tags),
			&q.Votes, &q.Author,
			&q.Type, &q.DataCreated, &q.Points)
		if err != nil {
			logz.Error().Err(err).Msg("Error to extract result from row")
		}

		//`SELECT id, question_id, answer_content, votes, author, type, created_at, points FROM deep_answer_enus WHERE question_id = $1`
		sqlStatementAnswers, _, _ := goqu.From("deep_answer_enus").
			Select("id", "question_id", "answer_content", "votes", "author", "type", "created_at", "points").Where(goqu.C("question_id").Eq(q.Id)).ToSQL()
		answerRows, err := db.Query(sqlStatementAnswers)
		if err != nil {
			logz.Error().Err(err).Msg("Error to get all deep_answer_enus from db")
			return nil, err
		}

		for answerRows.Next() {
			var a Answer
			err := answerRows.Scan(&a.Id, &a.QuestionId, &a.AnswerContent, &a.Votes, &a.Author, &a.Type, &a.DataCreated, &a.Points)
			if err != nil {
				logz.Error().Err(err).Msg("Error to extract result from answers rows")
				return nil, err
			}

			// `SELECT id, answer_id, commentary, author, type, created_at FROM deep_comment_enus WHERE answer_id = $1`
			sqlStatementComment, _, _ := goqu.From("deep_comment_enus").
				Select("id", "answer_id", "commentary", "author", "type", "created_at").Where(goqu.C("answer_id").Eq(a.Id)).ToSQL()
			commentRows, err := db.Query(sqlStatementComment)
			if err != nil {
				logz.Error().Err(err).Msg("Error to get all deep_comment_enus EnUS from db")
				return nil, err
			}

			for commentRows.Next() {
				var c Comment
				err := commentRows.Scan(&c.Id, &c.AnswerId, &c.Commentary, &c.Author, &c.Type, &c.DataCreated)
				if err != nil {
					logz.Error().Err(err).Msg("Error to extract result from comment rows")
					return nil, err
				}
				a.Comments = append(a.Comments, c)
			}
			q.Answers = append(q.Answers, a)
		}
		qs = append(qs, q)
	}
	return qs, nil
}

func (da DeepEnusPostgres) GetIndexLastQ() (LastQuestion, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return LastQuestion{}, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idDeepEnUS := "deep_enus"

	var lq LastQuestion
	//SELECT id, question_number FROM last_question WHERE id = $1
	sqlStatement, _, _ := goqu.From("last_question").Select("id", "question_number").Where(goqu.C("id").Eq(idDeepEnUS)).ToSQL()
	result := db.QueryRow(sqlStatement)
	err = result.Scan(&lq.Id, &lq.QuestionNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			logz.Error().Err(err).Msg("Error, no rows in result")
			return LastQuestion{}, err
		}
		logz.Error().Err(err).Msg("Error to extract result from row")
		return LastQuestion{}, err
	}

	return lq, nil
}

func (da DeepEnusPostgres) GetLinkDeepEnUS() (LinkDeepEnus, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return LinkDeepEnus{}, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idDeepEnUS := "deep_enus"
	var linkDeepEnUS LinkDeepEnus

	//SELECT id, link FROM link_scraper WHERE id = $1
	sqlStatement, _, _ := goqu.From("link_scraper").Select("id", "link").Where(goqu.C("id").Eq(idDeepEnUS)).ToSQL()
	result := db.QueryRow(sqlStatement)
	err = result.Scan(&linkDeepEnUS.Id, &linkDeepEnUS.Link)
	if err != nil {
		if err == sql.ErrNoRows {
			logz.Error().Err(err).Msg("Error, no rows in result")
			return LinkDeepEnus{}, err
		}
		logz.Error().Err(err).Msg("Error to extract result from row")
		return LinkDeepEnus{}, err
	}

	return linkDeepEnUS, nil
}

func (da DeepEnusPostgres) SaveLastQ(questionNumber int) error {
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

	idDeepEnUS := "deep_enus"

	//UPDATE last_question SET question_number = $2 WHERE id = $1
	sqlStatement, _, _ := goqu.Update("last_question").Set(goqu.C("question_number").Eq(questionNumber)).Where(goqu.C("id").Eq(idDeepEnUS)).ToSQL()
	_, err = db.Exec(sqlStatement)
	if err != nil {
		logz.Error().Err(err).Msgf("Error to update last question into db")
		return err
	}
	return nil
}
