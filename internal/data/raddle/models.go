package raddle

type Post struct {
	Id           int          `db:"id" json:"id"`
	Title        string       `db:"title" json:"title"`
	Link         string       `db:"link" json:"link"`
	Post         string       `db:"post" json:"post"`
	Forum        string       `db:"forum" json:"forum"`
	Votes        string       `db:"votes" json:"votes"`
	User         string       `db:"user" json:"user"`
	DataCreated  string       `db:"created_at" json:"dataCreated"`
	Commentaries []Commentary `db:"commentaries" json:"commentaries"`
}

type Commentary struct {
	Id          int    `db:"id" json:"id"`
	PostId      int    `db:"post_id" json:"postId"`
	Commentary  string `db:"commentary" json:"commentary"`
	User        string `db:"user" json:"user"`
	Votes       string `db:"votes" json:"votes"`
	DataCreated string `db:"created_at" json:"dataCreated"`
}

type LastPost struct {
	Id         string `db:"id" json:"id"`
	PostNumber int    `db:"question_number" json:"questionNumber"`
}

type LinkRaddle struct {
	Id   string `db:"id" json:"id"`
	Link string `db:"link" json:"link"`
}

type Posts []Post
