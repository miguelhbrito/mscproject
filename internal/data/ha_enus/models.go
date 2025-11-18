package haenus

type Question struct {
	Id          int      `db:"id" json:"id"`
	Title       string   `db:"title" json:"title"`
	Question    string   `db:"question" json:"question"`
	Category    string   `db:"category" json:"category"`
	Tags        []string `db:"tags" json:"tags"`
	UpVote      string   `db:"up_vote" json:"upVotes"`
	DownVote    string   `db:"down_vote"  json:"downVotes"`
	Author      string   `db:"author" json:"author"`
	Type        string   `db:"type" json:"type"`
	DataCreated string   `db:"created_at" json:"dataCreated"`
	Points      string   `db:"points" json:"points"`
	Answers     []Answer `db:"answers" json:"answers"`
}

type Answer struct {
	Id            int       `db:"id" json:"id"`
	QuestionId    int       `db:"question_id" json:"questionId"`
	AnswerContent string    `db:"answer_content" json:"answerContent"`
	UpVote        string    `db:"up_vote" json:"upVotes"`
	DownVote      string    `db:"down_vote" json:"downVotes"`
	Type          string    `db:"type" json:"type"`
	Author        string    `db:"author" json:"author"`
	DataCreated   string    `db:"created_at" json:"dataCreated"`
	Points        string    `db:"points" json:"points"`
	Comments      []Comment `db:"comments" json:"comments"`
}

type Comment struct {
	Id          int    `db:"id" json:"id"`
	AnswerId    int    `db:"answer_id" json:"answerId"`
	Commentary  string `db:"commentary" json:"comment"`
	Type        string `db:"type" json:"type"`
	Author      string `db:"author" json:"author"`
	DataCreated string `db:"created_at" json:"dataCreated"`
}

type LastQuestion struct {
	Id             string `db:"id" json:"id"`
	QuestionNumber int    `db:"question_number" json:"questionNumber"`
}

type LinkHAEnglish struct {
	Id   string `db:"id" json:"id"`
	Link string `db:"link" json:"link"`
}

type Questions []Question
