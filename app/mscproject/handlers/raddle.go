package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miguelhbrito/mscproject/internal/commons"
	ra "github.com/miguelhbrito/mscproject/internal/core/raddle"
	data "github.com/miguelhbrito/mscproject/internal/data/raddle"
)

type RaddleHandler struct {
	core ra.RaInt
	log  *log.Logger
}

func NewRaddleHandler(env *Env) *RaddleHandler {

	db := data.RaddlePostgres{}

	core := ra.NewCore(
		db,
		env.log,
	)

	// FIXME
	// Deactivated due to server shut down
	/*job := ra.NewRaddleJob(env.log)
	err := job.Run(
		core,
	)
	if err != nil {
		log.Println("fail to init raddle job")
	}*/

	return &RaddleHandler{
		core: core,
		log:  env.log,
	}
}

func (h RaddleHandler) post(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to starting raddle scraper")

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	t := time.Now()

	// Logging
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-raddle.log", pwd, t.Format(YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create raddle.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	h.log.Println("file raddle.log created on: ", fileName)
	log := log.New(file, "RADDLE : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Channel
	done := make(chan commons.Done)
	errorCh := make(chan commons.Error)

	go h.core.Scrapper(done, errorCh, log)
	select {
	case <-done:
		h.log.Println("success:", <-done)
	case e := <-errorCh:
		h.log.Println("failure:", e, <-errorCh)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "go routine started !",
	})
}

func (h RaddleHandler) get(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to list all raddle posts")

	qts, err := h.core.List()
	if err != nil {
		h.log.Printf("Error to list Raddle posts from db: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Hidden Answers Ptbr questions: ": qts,
		})
	}
}

func (h RaddleHandler) getJSON(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to create a JSON file with raddle post")

	err := h.core.CreateJSON()
	if err != nil {
		h.log.Printf("Error to create a JSON file with raddle posts: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "success",
		})
	}
}
