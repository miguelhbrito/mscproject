package haenus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/miguelhbrito/mscproject/internal/commons"
	ha "github.com/miguelhbrito/mscproject/internal/data/ha_enus"
)

type HAInt interface {
	Scrapper(chan commons.Done, chan commons.Error, *log.Logger)
	List() (ha.Questions, error)
	CreateJSON() error
}

type core struct {
	haeng ha.HAEnglishInt
	log   *log.Logger
}

func NewCore(haeng ha.HAEnglishInt, log *log.Logger) HAInt {
	return core{
		haeng: haeng,
		log:   log,
	}
}

func (cS core) getLastestQuestion(linkOnion string, done chan commons.Done, errorCh chan commons.Error) string {
	url := fmt.Sprintf("%s/questions", linkOnion)

	// Instantiate default collectors
	clq := colly.NewCollector()

	rp, err := proxy.RoundRobinProxySwitcher("socks5://mytor:9050")
	if err != nil {
		cS.log.Fatal(err)
	}
	clq.SetProxyFunc(rp)
	clq.SetRequestTimeout(100 * time.Second)

	clq.OnResponse(func(r *colly.Response) {
		cS.log.Println("response received:", r.StatusCode, r.Request.URL)
	})

	clq.OnError(func(r *colly.Response, err error) {
		cS.log.Println("response error:", r.StatusCode, err, r.Request.URL)
		if err != nil {
			errorCh <- commons.Error{
				Msg:    "finish go routine execution, error getting last question",
				Status: 500,
			}

		}
	})

	var questionNumber string

	clq.OnHTML("div.qa-main ", func(h *colly.HTMLElement) {
		var questionLink string
		h.ForEach(".qa-q-list", func(index int, el *colly.HTMLElement) {
			if index == 0 {
				questionLink = el.ChildAttr(".qa-q-item-title > a", "href")
			}
		})
		splitQuestion := strings.Split(questionLink, "/")
		questionNumber = splitQuestion[2]
	})

	cS.log.Println("Visiting to get last question", url)

	_ = clq.Visit(url)

	return questionNumber
}

func (cS core) getLastIndex() (int, error) {
	lq, err := cS.haeng.GetIndexLastQ()
	if err != nil {
		cS.log.Println("error to get last question index, err:", err.Error())
		return 0, err
	}
	return lq.QuestionNumber, nil
}

func (cS core) getLinkHAEnglish() (string, error) {
	link, err := cS.haeng.GetLinkHAEnglish()
	if err != nil {
		cS.log.Println("error to get HA english link, err:", err.Error())
		return "", err
	}
	return link.Link, nil
}

func (cS core) Scrapper(done chan commons.Done, errorCh chan commons.Error, log *log.Logger) {
	defer close(done)
	defer close(errorCh)

	cS.log.Println("Starting new scraper ha-enUS!")

	lq, err := cS.getLastIndex()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting last index saved",
			Status: 500,
		}
	}
	log.Println("Question number from last job: ", lq)

	url, err := cS.getLinkHAEnglish()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting the link to Hidden Answers English",
			Status: 500,
		}
	}
	log.Println("Got link Hidden Answers link: ", url)

	lastQuestionStr := cS.getLastestQuestion(url, done, errorCh)
	lastQuestionInt, _ := strconv.Atoi(lastQuestionStr)

	// Retroactive questions to check
	if lq > 0 {
		lq -= commons.DAENUS
	}

	if lq == lastQuestionInt {
		log.Println("Last question saved on db is equal to lastest one on homepage")
		errorCh <- commons.Error{
			Msg:    "There is no new questions",
			Status: 500,
		}
	}

	var questions []ha.Question

	// Instantiate default collector
	c := colly.NewCollector(
		colly.Async(true),
	)

	rp, err := proxy.RoundRobinProxySwitcher("socks5://mytor:9050")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.SetRequestTimeout(100 * time.Second)

	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10,
		Delay:       1 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received:", r.StatusCode, r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("response error:", r.StatusCode, err, r.Request.URL)
	})

	c.OnHTML("div.qa-main ", func(h *colly.HTMLElement) {
		var question ha.Question

		question.Title = h.ChildText("div.qa-main-heading")
		log.Println("Starting getting question from:", question.Title)
		question.UpVote = h.ChildText("span.qa-upvote-count-data ")
		question.DownVote = h.ChildText("span.qa-downvote-count-data ")
		question.Question = h.ChildText("div.qa-q-view-content.qa-post-content ")
		question.Category = h.ChildText(".qa-q-view-where-data > a ")

		h.ForEach(".qa-q-view-tag-item", func(_ int, el *colly.HTMLElement) {
			question.Tags = append(question.Tags, el.ChildText(".qa-tag-link"))
		})

		question.Type = h.ChildText("a.qa-q-view-what ")
		question.DataCreated = h.ChildAttr("span.qa-q-view-when-data > time", "datetime")
		question.Author = h.ChildText("span.qa-q-view-who-data > span > a > span")
		question.Points = h.ChildText("span.qa-q-view-who-points-data")

		h.ForEach(".qa-a-list-item ", func(_ int, el *colly.HTMLElement) {
			var answers ha.Answer

			answers.UpVote = el.ChildText("span.qa-upvote-count-data ")
			answers.DownVote = el.ChildText("span.qa-downvote-count-data ")
			answers.AnswerContent = el.ChildText(".qa-a-item-content.qa-post-content ")
			answers.Type = el.ChildText("a.qa-a-item-what ")
			answers.DataCreated = el.ChildAttr("span.qa-a-item-when-data > time", "datetime")
			answers.Author = el.ChildText("span.qa-a-item-who-data > span > a > span")
			answers.Points = el.ChildText("span.qa-a-item-who-points-data")

			el.ForEach(".qa-c-list-item ", func(_ int, comment *colly.HTMLElement) {
				var commentStruct ha.Comment

				commentStruct.Commentary = comment.ChildText(".qa-c-item-content.qa-post-content ")
				commentStruct.Type = comment.ChildText("a.qa-c-item-what ")
				commentStruct.DataCreated = comment.ChildAttr("span.qa-c-item-when-data > time", "datetime")
				commentStruct.Author = comment.ChildText("span.qa-c-item-who-data > span > a > span")

				answers.Comments = append(answers.Comments, commentStruct)
			})
			question.Answers = append(question.Answers, answers)
		})
		if !reflect.ValueOf(question).IsZero() {
			questions = append(questions, question)
		}
	})

	for i := lq; i < lastQuestionInt; i++ {
		log.Println("Visiting", fmt.Sprintf("%s/%d", url, i))
		_ = c.Visit(fmt.Sprintf("%s/%d", url, i))
	}

	// Wait until threads are finished
	c.Wait()

	for _, q := range questions {
		if q.DataCreated != "" {
			err = cS.haeng.Save(q, log)
			if err != nil {
				log.Println("Error to save an question into database")
			}
		}
	}

	if lastQuestionInt > 0 {
		err = cS.haeng.SaveLastQ(lastQuestionInt)
		if err != nil {
			log.Println("Error to save last question number into database")
		}
	} else {
		log.Println("Unknown error occurred, the last question was tried to save into db was 0")
		errorCh <- commons.Error{
			Msg:    "Unknown error occurred, the last question was tried to save into db was 0",
			Status: 500,
		}
	}

	cS.log.Println("Successfully scraped hidden-answers enUS")

	done <- commons.Done{
		Msg:    "Successfully got scraped HA english",
		Status: 200,
	}
}

func (cS core) List() (ha.Questions, error) {
	qts, err := cS.haeng.List()
	if err != nil {
		return nil, err
	}

	return qts, nil
}

func (cS core) CreateJSON() error {

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	questionsPtbr, err := cS.List()
	if err != nil {
		return err
	}

	file, _ := json.MarshalIndent(questionsPtbr, "", " ")

	t := time.Now()
	_ = ioutil.WriteFile(pwd+"/"+
		t.Format(commons.YYYYMMDDhhmmss)+
		"-HAenus.json", file, 0644)

	return nil
}
