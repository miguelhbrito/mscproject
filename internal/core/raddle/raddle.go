package raddle

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
	ra "github.com/miguelhbrito/mscproject/internal/data/raddle"
)

type RaInt interface {
	Scrapper(chan commons.Done, chan commons.Error, *log.Logger)
	List() (ra.Posts, error)
	CreateJSON() error
}

type core struct {
	raddle ra.RaddleInt
	log    *log.Logger
}

func NewCore(raddle ra.RaddleInt, log *log.Logger) RaInt {
	return core{
		raddle: raddle,
		log:    log,
	}
}

func (cS core) getLatestPost(linkOnion string, done chan commons.Done, errorCh chan commons.Error) string {
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

	var postNumber string

	clq.OnHTML("main ", func(h *colly.HTMLElement) {
		var postLink string
		h.ForEach(".submission__nav ", func(index int, el *colly.HTMLElement) {
			if index == 0 {
				postLink = el.ChildAttr("a", "href")
			}
		})
		splitQuestion := strings.Split(postLink, "/")
		postNumber = splitQuestion[3]
	})

	cS.log.Println("Visiting to get last post", linkOnion)

	_ = clq.Visit(linkOnion)

	return postNumber
}

func (cS core) getLastIndex() (int, error) {
	lp, err := cS.raddle.GetIndexLastQ()
	if err != nil {
		cS.log.Println("error to get last post index, err:", err.Error())
		return 0, err
	}
	return lp.PostNumber, nil
}

func (cS core) getLinkRaddle() (string, error) {
	link, err := cS.raddle.GetLinkRaddle()
	if err != nil {
		cS.log.Println("error to get Raddle link, err:", err.Error())
		return "", err
	}
	return link.Link, nil
}

func (cS core) Scrapper(done chan commons.Done, errorCh chan commons.Error, log *log.Logger) {
	defer close(done)
	defer close(errorCh)

	cS.log.Println("Starting new scraper raddle!")

	lp, err := cS.getLastIndex()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting last index saved",
			Status: 500,
		}
	}
	log.Println("Post number from last job: ", lp)

	url, err := cS.getLinkRaddle()
	if err != nil {
		log.Println(err.Error())
		errorCh <- commons.Error{
			Err:    err,
			Msg:    "finish go routine execution, error getting the link to Hidden Answers Ptbr",
			Status: 500,
		}
	}
	log.Println("Got link Hidden Answers link: ", url)

	lastPostStr := cS.getLatestPost(url, done, errorCh)
	lastPostInt, _ := strconv.Atoi(lastPostStr)

	// Retroactive questions to check
	/*if lp > 0 {
		lp -= commons.RADDLE
	}*/

	if lp == lastPostInt {
		log.Println("Last post saved on db is equal to lastest one on homepage")
		errorCh <- commons.Error{
			Msg:    "There is no new posts",
			Status: 500,
		}
	}

	var posts []ra.Post

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

	c.OnResponse(func(r *colly.Response) {
		log.Println("response received:", r.StatusCode, r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("response error:", r.StatusCode, err, r.Request.URL)
	})

	c.OnHTML("main", func(h *colly.HTMLElement) {
		var post ra.Post

		post.Title = h.ChildText("a.submission__link")
		post.Link = h.ChildAttr("a.submission__link", "href")
		post.Post = h.ChildText("div.submission__body")
		post.User = h.ChildText("a.submission__submitter")
		post.Forum = h.ChildText("a.submission__forum")
		post.DataCreated = h.ChildAttr("p.submission__info > span > time", "datetime")
		post.Votes = h.ChildText("span.vote__net-score")

		h.ForEach(".comment__row", func(_ int, el *colly.HTMLElement) {
			var commentary ra.Commentary

			commentary.Commentary = el.ChildText(".comment__body > p")
			commentary.User = el.ChildText("span.fg-muted > a")
			commentary.DataCreated = el.ChildAttr("span.fg-muted > time", "datetime")
			commentary.Votes = h.ChildText("span.vote__net-score")

			post.Commentaries = append(post.Commentaries, commentary)
		})
		if !reflect.ValueOf(post).IsZero() && (post.Forum == "Tech" ||
			post.Forum == "hacktice" ||
			post.Forum == "hackbloc" ||
			post.Forum == "leaks" ||
			post.Forum == "netsec" ||
			post.Forum == "Privacy" ||
			post.Forum == "programming" ||
			post.Forum == "Cryptocurrency" ||
			post.Forum == "HackerNewsOuttakes") {
			posts = append(posts, post)
		}
	})

	for i := lp; i < lastPostInt; i++ {
		log.Println("Visiting", fmt.Sprintf("%s%d", url, i))
		_ = c.Visit(fmt.Sprintf("%s%d", url, i))
	}

	// Wait until threads are finished
	c.Wait()

	for _, p := range posts {
		if p.DataCreated != "" {
			err = cS.raddle.Save(p, log)
			if err != nil {
				log.Println("Error to save a post into database")
			}
		}
	}

	if lastPostInt > 0 {
		err = cS.raddle.SaveLastP(lastPostInt)
		if err != nil {
			log.Println("Error to save last post number into database")
		}
	} else {
		log.Println("Unknown error occurred, the last post was tried to save into db was 0")
		errorCh <- commons.Error{
			Msg:    "Unknown error occurred, the last post was tried to save into db was 0",
			Status: 500,
		}
	}

	cS.log.Println("Successfully scraped Raddle")

	done <- commons.Done{
		Msg:    "Raddle got successfully scraped",
		Status: 200,
	}

}

func (cS core) List() (ra.Posts, error) {
	ps, err := cS.raddle.List()
	if err != nil {
		return nil, err
	}

	return ps, nil
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
		"-Raddle.json", file, 0644)

	return nil
}
