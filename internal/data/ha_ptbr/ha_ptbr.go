package haptbr

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"

	"log"

	dbConn "github.com/miguelhbrito/mscproject/platform/db_connect"
	logz "github.com/rs/zerolog/log"
)

type HAPtbrInt interface {
	Save(Question, *log.Logger) error
	List() (Questions, error)
	GetIndexLastQ() (LastQuestion, error)
	GetLinkHAPtbr() (LinkHAPtbr, error)
	SaveLastQ(questionNumber int) error
}

type HAPtbrPostgres struct{}

func (ha HAPtbrPostgres) Save(q Question, log *log.Logger) error {
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
		UpVote:      q.UpVote,
		DownVote:    q.DownVote,
		Author:      q.Author,
		Type:        q.Type,
		DataCreated: q.DataCreated,
		Points:      q.Points,
	}

	for _, value := range q.Answers {
		a := Answer{
			AnswerContent: value.AnswerContent,
			UpVote:        value.UpVote,
			DownVote:      value.DownVote,
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
	// `SELECT title FROM ha_question_ptbr WHERE title = $1`
	sqlStatementVerify, _, _ := goqu.From("ha_question_ptbr").
		Select("id", "title").Where(goqu.C("title").Eq(qs.Title)).ToSQL()
	result := db.QueryRow(sqlStatementVerify)
	err = result.Scan(&qsVerify.Id, &qsVerify.Title)
	if err != nil {
		if err == sql.ErrNoRows {

			log.Println("Saving new question into db")
			questionId, err := ha.saveQuestion(qs, db, log)
			if err != nil {
				log.Println("Error on saveQuestion:", err)
				return err
			}

			for _, answer := range ans {
				answerId, err := ha.saveAnswer(questionId, answer, db, log)
				if err != nil {
					log.Println("Error on saveAnswer:", err)
					return err
				}
				for _, comment := range cs {
					if err := ha.saveComment(answerId, comment, db, log); err != nil {
						log.Println("Error on saveComment:", err)
						return err
					}
				}
			}
			log.Println("New question HiddenAnswers-PtBR, answer and comments was successfully saved into db!")
			return nil
		}
	}

	log.Println("HiddenAnswers-PtBR question already exist in DB!")
	log.Println("Check in new answers or comments!")

	var answersVerify []Answer
	// `SELECT * FROM ha_answer_ptbr WHERE question_id = $1`
	sqlAnsVerify, _, _ := goqu.From("ha_answer_ptbr").Where(goqu.C("question_id").Eq(qsVerify.Id)).ToSQL()
	rows, err := db.Query(sqlAnsVerify)
	if err != nil {
		log.Println("Error getting the answers from db, err:", err)
		return err
	}
	for rows.Next() {

		var awsVerify Answer
		err := rows.Scan(&awsVerify.Id, &awsVerify.QuestionId, &awsVerify.AnswerContent, &awsVerify.UpVote, &awsVerify.DownVote, &awsVerify.Author, &awsVerify.Type, &awsVerify.DataCreated, &awsVerify.Points)
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
		// If the key exists
		if answerValue, ok := asnwersMap[idx]; ok {

			if answerExisted.AnswerContent == answerValue.AnswerContent {

				log.Println("Old answer, check in new comments")
				var commentVerify []Comment
				// `SELECT * FROM ha_comment_ptbr WHERE answer_id = $1`
				sqlCommentVerify, _, _ := goqu.From("ha_comment_ptbr").Where(goqu.C("answer_id").Eq(answerValue.Id)).ToSQL()
				rows, err := db.Query(sqlCommentVerify)
				if err != nil {
					log.Println("Error getting the comments from db, err:", err)
					return err
				}
				for rows.Next() {

					var comVerify Comment
					err := rows.Scan(&comVerify.Id, &comVerify.AnswerId, &comVerify.Commentary, &comVerify.Author, &comVerify.Type, &comVerify.DataCreated)
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

				if err := ha.saveCommentExistedQuestion(commentsMap, answerValue, answerExisted, db, log); err != nil {
					return err
				}
			}
		} else {
			log.Println("Inserting new answer in an existing question:", qsVerify.Id, answerExisted)

			ansIdExisted, err := ha.saveAnswer(qsVerify.Id, answerExisted, db, log)
			if err != nil {
				log.Println("Error on saveAnswer of a previous question into db, err: ", err)
				return err
			}

			log.Println("New answer id from old question from answer saved on db ", ansIdExisted)
			for _, comment := range cs {
				if err := ha.saveComment(ansIdExisted, comment, db, log); err != nil {
					log.Println("Error on saveComment of a previous question into db, err: ", err)
					return err
				}
			}
		}
	}

	log.Println("Successfully saved the question HiddenAnswers-PtBR, answer and comments into db!")
	return nil
}

func (ha HAPtbrPostgres) saveQuestion(qs Question, db *sql.DB, log *log.Logger) (int, error) {
	var questionId int
	log.Println("Inserting new HA Ptbr question into DB")
	sqlStatement :=
		`INSERT INTO "ha_question_ptbr" (title, question, category, tags, up_vote, down_vote, author, type, created_at, points) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	err := db.QueryRow(sqlStatement, qs.Title, qs.Question, qs.Category, pq.Array(qs.Tags), qs.UpVote, qs.DownVote, qs.Author, qs.Type, qs.DataCreated, qs.Points).Scan(&questionId)
	if err != nil {
		log.Println("Error to insert an new ha ptbr question into db, err: ", err)
		return 0, err
	}
	log.Println("Question id from question saved on db ", questionId)

	return questionId, nil
}

func (ha HAPtbrPostgres) saveAnswer(qsId int, answer Answer, db *sql.DB, log *log.Logger) (int, error) {
	var answerId int
	sqlStatement :=
		`INSERT INTO ha_answer_ptbr (question_id, answer_content, up_vote, down_vote, author, type, created_at, points) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := db.QueryRow(sqlStatement, qsId, answer.AnswerContent, answer.UpVote, answer.DownVote, answer.Author, answer.Type, answer.DataCreated, answer.Points).Scan(&answerId)
	if err != nil {
		log.Println("Error to insert an ha ptbr answer into db, err: ", err)
		return 0, err
	}
	log.Println("Answer id from answer saved on db ", answerId)

	return answerId, nil
}

func (ha HAPtbrPostgres) saveComment(ansId int, comment Comment, db *sql.DB, log *log.Logger) error {
	sqlStatement :=
		`INSERT INTO ha_comment_ptbr (answer_id, commentary, author, type, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStatement, ansId, comment.Commentary, comment.Author, comment.Type, comment.DataCreated)
	if err != nil {
		log.Println("Error to insert an ha ptbr commentary into db, err: ", err)
		return err
	}

	return nil
}

func (ha HAPtbrPostgres) saveCommentExistedQuestion(
	commentsMap map[int]Comment,
	answerValue Answer,
	answerExisted Answer,
	db *sql.DB,
	log *log.Logger) error {

	for idxC, comment := range answerExisted.Comments {
		// If the key doesn't exists
		if commentValue, ok := commentsMap[idxC]; !ok {
			if comment.Commentary != commentValue.Commentary {
				log.Println("Inserting new comment in an existing question/answer:", answerExisted.Id, commentValue)

				if err := ha.saveComment(answerValue.Id, comment, db, log); err != nil {
					log.Println("Error on saveComment of a previous question into db, err: ", err)
					return err
				}
			}
		}
	}
	return nil
}

func (ha HAPtbrPostgres) List() (Questions, error) {
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

	sqlStatement := `SELECT id, title, question, category, tags, up_vote, down_vote, author, type, created_at, points FROM ha_question_ptbr`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		logz.Error().Err(err).Msg("Error to get all ha_ptbr_question from db")
		return nil, err
	}

	for rows.Next() {
		var q Question
		err := rows.Scan(&q.Id, &q.Title, &q.Question, &q.Category, pq.Array(&q.Tags), &q.UpVote, &q.DownVote, &q.Author, &q.Type, &q.DataCreated, &q.Points)
		if err != nil {
			logz.Error().Err(err).Msg("Error to extract result from row")
		}

		sqlStatementAnswers := `SELECT id, question_id, answer_content, up_vote, down_vote, author, type, created_at, points FROM ha_answer_ptbr WHERE question_id = $1`
		answerRows, err := db.Query(sqlStatementAnswers, q.Id)
		if err != nil {
			logz.Error().Err(err).Msg("Error to get all ha_answer_question from db")
			return nil, err
		}

		for answerRows.Next() {
			var a Answer
			err := answerRows.Scan(&a.Id, &a.QuestionId, &a.AnswerContent, &a.UpVote, &a.DownVote, &a.Author, &a.Type, &a.DataCreated, &a.Points)
			if err != nil {
				logz.Error().Err(err).Msg("Error to extract result from answers rows")
				return nil, err
			}

			sqlStatementComment := `SELECT id, answer_id, commentary, author, type, created_at FROM ha_comment_ptbr WHERE answer_id = $1`
			commentRows, err := db.Query(sqlStatementComment, a.Id)
			if err != nil {
				logz.Error().Err(err).Msg("Error to get all ha_comment_question ptbr from db")
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

func (ha HAPtbrPostgres) GetIndexLastQ() (LastQuestion, error) {
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

	idHAPtbr := "ha_ptbr"

	var lq LastQuestion
	// SELECT id, question_number FROM ha_ptbr_last WHERE id = $1
	sqlStatement, _, _ := goqu.From("last_question").Select("id", "question_number").Where(goqu.C("id").Eq(idHAPtbr)).ToSQL()
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

func (ha HAPtbrPostgres) GetLinkHAPtbr() (LinkHAPtbr, error) {
	db, err := dbConn.ConnectDB()
	if err != nil {
		log.Println("Error to connect to db:", err)
		return LinkHAPtbr{}, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error to close database connection: %v", err)
		}
	}()

	idHAPtbr := "ha_ptbr"
	var linkHAPtbr LinkHAPtbr

	//SELECT id, link FROM link_scraper WHERE id = $1
	sqlStatement, _, _ := goqu.From("link_scraper").Select("id", "link").Where(goqu.C("id").Eq(idHAPtbr)).ToSQL()
	result := db.QueryRow(sqlStatement)
	err = result.Scan(&linkHAPtbr.Id, &linkHAPtbr.Link)
	if err != nil {
		if err == sql.ErrNoRows {
			logz.Error().Err(err).Msg("Error, no rows in result")
			return LinkHAPtbr{}, err
		}
		logz.Error().Err(err).Msg("Error to extract result from row")
		return LinkHAPtbr{}, err
	}

	return linkHAPtbr, nil
}

func (ha HAPtbrPostgres) SaveLastQ(questionNumber int) error {
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

	idHAPtbr := "ha_ptbr"

	sqlStatement := `UPDATE last_question SET question_number = $2 WHERE id = $1`
	_, err = db.Exec(sqlStatement, idHAPtbr, questionNumber)
	if err != nil {
		logz.Error().Err(err).Msgf("Error to update last question into db")
		return err
	}
	return nil
}
