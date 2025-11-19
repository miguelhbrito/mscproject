package deepptbr

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/miguelhbrito/mscproject/internal/commons"
	data "github.com/miguelhbrito/mscproject/internal/data/deep_ptbr"
)

type DAInt interface {
	Scrapper(chan commons.Done, chan commons.Error, *log.Logger)
	List() (data.Questions, error)
	CreateJSON() error
}

type core struct {
	data data.DeepPtbrInt
	log  *log.Logger
}

func NewCore(data data.DeepPtbrInt, log *log.Logger) DAInt {
	return core{
		data: data,
		log:  log,
	}
}

func (cS core) getLastIndex() (int, error) {
	lq, err := cS.data.GetIndexLastQ()
	if err != nil {
		cS.log.Println("error to get last question index from deep-ptbr, err:", err.Error())
		return 0, err
	}
	return lq.QuestionNumber, nil
}

func (cS core) getLinkDeepPtbr() (string, error) {
	link, err := cS.data.GetLinkDeepPtbr()
	if err != nil {
		cS.log.Println("error to get deep-ptbr link, err:", err.Error())
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
		questionNumber = commons.GetStringInBetweenTwoString(questionLink, "?qa=", "&qa")
		if questionNumber == "" {
			errorCh <- commons.Error{
				Msg:    "finish go routine execution, error getting questionNumber from last question link",
				Status: 500,
			}

		}
		cS.log.Println("Last question number from DeepAnswers Ptbr", questionNumber)
	})

	cS.log.Println("Visiting to get last question", url)

	_ = clq.Visit(url)

	return questionNumber

}

func (cS core) Scrapper(done chan commons.Done, errorCh chan commons.Error, log *log.Logger) {
	defer close(done)
	defer close(errorCh)

	cS.log.Println("Starting new scraper deep-ptBR!")

	lq, err := cS.getLastIndex()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting last index saved",
			Status: 500,
		}

	}
	log.Println("Question number from last job from deep-ptbr: ", lq)

	url, err := cS.getLinkDeepPtbr()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting the link to Hidden Answers Ptbr",
			Status: 500,
		}

	}
	log.Println("Got link DeepAnswers-PTBR link: ", url)

	lastQuestionStr := cS.getLastestQuestion(url, done, errorCh)
	lastQuestionInt, _ := strconv.Atoi(lastQuestionStr)

	// Retroactive questions to check
	if lq > 0 {
		lq -= commons.DEEPPTBR
	}

	if lq == lastQuestionInt {
		log.Println("Last question saved on db is equal to lastest one on homepage")
		errorCh <- commons.Error{
			Msg:    "There is no new questions",
			Status: 500,
		}

	}

	var questions []data.Question

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
		var question data.Question

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
			var answers data.Answer

			answers.Votes = el.ChildText("span.qa-netvote-count-data ")
			answers.AnswerContent = el.ChildText(".qa-a-item-content.qa-post-content ")
			answers.Type = el.ChildText("a.qa-a-item-what ")
			answers.DataCreated = el.ChildAttr("span.qa-a-item-when-data > time", "datetime")
			answers.Author = el.ChildText("span.qa-a-item-who-data > span > a > span")
			answers.Points = el.ChildText("span.qa-a-item-who-points-data")

			el.ForEach(".qa-c-list-item ", func(_ int, comment *colly.HTMLElement) {
				var commentStruct data.Comment

				commentStruct.Commentary = comment.ChildText(".qa-c-item-content.qa-post-content ")
				commentStruct.Type = comment.ChildText("a.qa-c-item-what ")
				commentStruct.DataCreated = comment.ChildAttr("span.qa-c-item-when-data > time", "datetime")
				commentStruct.Author = comment.ChildText("span.qa-c-item-who-data > span > a > span")

				answers.Comments = append(answers.Comments, commentStruct)
			})
			question.Answers = append(question.Answers, answers)
		})
		if !reflect.ValueOf(question).IsZero() && ((question.Category == "Tecnologia") ||
			(question.Category == "Segurança e privacidade") ||
			(question.Category == "Hacking") ||
			(question.Category == "Cracking/Warez") ||
			(question.Category == "Mercados") ||
			(question.Category == "Cripto-anarquia e rede sombria") ||
			(question.Category == "Golpes e golpistas") ||
			(question.Category == "Links")) {
			questions = append(questions, question)
		}
	})

	for i := lq; i < lastQuestionInt; i++ {
		log.Println("Visiting question from deepAnswers ptbr: ", fmt.Sprintf("%s%d", url, i))
		_ = c.Visit(fmt.Sprintf("%s%d", url, i))
	}

	// Wait until threads are finished
	c.Wait()

	for _, q := range questions {
		if q.DataCreated != "" {
			err = cS.data.Save(q, log)
			if err != nil {
				log.Println("Error to save an question into database")
			}
		}
	}

	if lastQuestionInt > 0 {
		err = cS.data.SaveLastQ(lastQuestionInt)
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

	cS.log.Println("Successfully scraped deepAnswers ptBR")

	done <- commons.Done{
		Msg:    "Successfully got scraped deepAnswers ptbr",
		Status: 200,
	}

}

func (cS core) List() (data.Questions, error) {
	qts, err := cS.data.List()
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
	_ = os.WriteFile(pwd+"/"+
		t.Format(commons.YYYYMMDDhhmmss)+
		"-DeepAnswersPtbr.json", file, 0644)

	return nil
}
