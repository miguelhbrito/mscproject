package deepenus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/miguelhbrito/mscproject/internal/commons"
	da "github.com/miguelhbrito/mscproject/internal/data/deep_enus"
)

type DAInt interface {
	Scrapper(chan commons.Done, chan commons.Error, *log.Logger)
	List() (da.Questions, error)
	CreateJSON() error
}

type core struct {
	deepStorage da.DeepEnUSInt
	log         *log.Logger
}

func NewCore(deepStorage da.DeepEnUSInt, log *log.Logger) DAInt {
	return core{
		deepStorage: deepStorage,
		log:         log,
	}
}

func (cS core) getLastIndex() (int, error) {
	lq, err := cS.deepStorage.GetIndexLastQ()
	if err != nil {
		cS.log.Println("error to get last question index from deep-enus, err:", err.Error())
		return 0, err
	}
	return lq.QuestionNumber, nil
}

func (cS core) getLinkDeepEnUS() (string, error) {
	link, err := cS.deepStorage.GetLinkDeepEnUS()
	if err != nil {
		cS.log.Println("error to get deep-enus link, err:", err.Error())
		return "", err
	}
	return link.Link, nil
}

func (cS core) getLastestQuestion(linkOnion string, done chan commons.Done, errorCh chan commons.Error) string {
	url := fmt.Sprintf("%squestions", linkOnion)

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
				Status: r.StatusCode,
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
		questionNumber = commons.GetStringInBetweenTwoString(questionLink, "?qa=", "&qa")
		if questionNumber == "" {
			errorCh <- commons.Error{
				Msg:    "finish go routine execution, error getting questionNumber from last question link",
				Status: 500,
			}

		}
		cS.log.Println("Last question number from DeepAnswers EnUS", questionNumber)
	})

	cS.log.Println("Visiting to get last question", url)

	_ = clq.Visit(url)

	return questionNumber

}

func (cS core) Scrapper(done chan commons.Done, errorCh chan commons.Error, log *log.Logger) {
	defer close(done)
	defer close(errorCh)

	cS.log.Println("Starting new scraper deep-enUS!")

	lq, err := cS.getLastIndex()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting last index saved",
			Status: 500,
		}
	}
	log.Println("Question number from last job from deep-enUS: ", lq)

	url, err := cS.getLinkDeepEnUS()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting the link to DeepAnswers EnUS",
			Status: 500,
		}

	}
	log.Println("Got link DeepAnswers-EnUS link: ", url)

	lastQuestionStr := cS.getLastestQuestion(url, done, errorCh)
	lastQuestionInt, _ := strconv.Atoi(lastQuestionStr)

	// Retroactive questions to check
	if lq > 0 {
		lq -= commons.DEEPENUS
	}

	if lq == lastQuestionInt {
		log.Println("Last question saved on db is equal to lastest one on homepage")
		errorCh <- commons.Error{
			Msg:    "There is no new questions",
			Status: 500,
		}

	}

	var questions []da.Question

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
		var question da.Question

		question.Title = h.ChildText("div.qa-main-heading")
		log.Println("Starting getting question from:", question.Title)
		question.Votes = h.ChildText("span.qa-netvote-count-data ")
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
			var answers da.Answer

			answers.Votes = el.ChildText("span.qa-netvote-count-data ")
			answers.AnswerContent = el.ChildText(".qa-a-item-content.qa-post-content ")
			answers.Type = el.ChildText("a.qa-a-item-what ")
			answers.DataCreated = el.ChildAttr("span.qa-a-item-when-data > time", "datetime")
			answers.Author = el.ChildText("span.qa-a-item-who-data > span > a > span")
			answers.Points = el.ChildText("span.qa-a-item-who-points-data")

			el.ForEach(".qa-c-list-item ", func(_ int, comment *colly.HTMLElement) {
				var commentStruct da.Comment

				commentStruct.Commentary = comment.ChildText(".qa-c-item-content.qa-post-content ")
				commentStruct.Type = comment.ChildText("a.qa-c-item-what ")
				commentStruct.DataCreated = comment.ChildAttr("span.qa-c-item-when-data > time", "datetime")
				commentStruct.Author = comment.ChildText("span.qa-c-item-who-data > span > a > span")

				answers.Comments = append(answers.Comments, commentStruct)
			})
			question.Answers = append(question.Answers, answers)
		})
		if !reflect.ValueOf(question).IsZero() && ((question.Category == "Technology") ||
			(question.Category == "Security and privacy") ||
			(question.Category == "Hacking") ||
			(question.Category == "Cracking/Warez") ||
			(question.Category == "Markets") ||
			(question.Category == "Cryptoanarchy and darknets") ||
			(question.Category == "Scams and scammers")) {
			questions = append(questions, question)
		}
	})

	for i := lq; i < lastQuestionInt; i++ {
		log.Println("Visiting question from deepAnswers enUS: ", fmt.Sprintf("%s%d", url, i))
		_ = c.Visit(fmt.Sprintf("%s%d", url, i))
	}

	// Wait until threads are finished
	c.Wait()

	for _, q := range questions {
		if q.DataCreated != "" {
			err = cS.deepStorage.Save(q, log)
			if err != nil {
				log.Println("Error to save an question into database")
			}
		}
	}

	if lastQuestionInt > 0 {
		err = cS.deepStorage.SaveLastQ(lastQuestionInt)
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

	cS.log.Println("Successfully scraped deepAnswers enUS")

	done <- commons.Done{
		Msg:    "Successfully got scraped deepAnswers enUS",
		Status: 200,
	}
}

func (cS core) List() (da.Questions, error) {
	qts, err := cS.deepStorage.List()
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
		"-DeepAnswersEnUS.json", file, 0644)

	return nil

}
