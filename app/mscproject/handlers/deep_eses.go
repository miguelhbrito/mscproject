package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miguelhbrito/mscproject/internal/commons"
	da "github.com/miguelhbrito/mscproject/internal/core/deep_eses"
	data "github.com/miguelhbrito/mscproject/internal/data/deep_eses"
)

type DeepesESHandler struct {
	core da.DAInt
	log  *log.Logger
}

func NewDeepesESHandler(env *Env) *DeepesESHandler {

	db := data.DeepEsesPostgres{}

	core := da.NewCore(
		db,
		env.log,
	)

	// FIXME
	// Deactivated due to server shut down
	/*job := da.NewDeepJob(env.log)
	err := job.Run(
		core,
	)
	if err != nil {
		log.Println("fail to init deepAnswers esES job")
	}*/

	return &DeepesESHandler{
		core: core,
		log:  env.log,
	}
}

func (h DeepesESHandler) post(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to starting DeepAnswers Spanish-ES scraper")

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	t := time.Now()

	// Logging
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-deep_esES.log", pwd, t.Format(YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create deep_esES.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	h.log.Println("file deep_esES.log created on: ", fileName)
	log := log.New(file, "DEEP-ESES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

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

func (h DeepesESHandler) get(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to list all DeepAnswers Spanish-ES content")

	qts, err := h.core.List()
	if err != nil {
		h.log.Printf("Error to list DeepAnswers Spanish-ES questions from db: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"DeepAnswers Spanish-ES questions: ": qts,
		})
	}
}

func (h DeepesESHandler) getJSON(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to create a JSON file with DeepAnswers Spanish-ES content")

	err := h.core.CreateJSON()
	if err != nil {
		h.log.Printf("Error to create a JSON file with DeepAnswers Spanish-ES content: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "success",
		})
	}
}
