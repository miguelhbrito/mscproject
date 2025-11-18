package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miguelhbrito/mscproject/internal/commons"
	da "github.com/miguelhbrito/mscproject/internal/core/deep_ptbr"
	data "github.com/miguelhbrito/mscproject/internal/data/deep_ptbr"
)

type DeepptBRHandler struct {
	core da.DAInt
	log  *log.Logger
}

func NewDeepptBRHandler(env *Env) *DeepptBRHandler {

	db := data.DeepPtbrPostgres{}

	core := da.NewCore(
		db,
		env.log,
	)

	// FIXME
	// Deactivated due to server shut down
	/*job := da.NewDeepPtbrJob(env.log)
	err := job.Run(
		core,
	)
	if err != nil {
		log.Println("fail to init deepAnswers ptbr job")
	}*/

	return &DeepptBRHandler{
		core: core,
		log:  env.log,
	}
}

func (h DeepptBRHandler) post(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to starting DeepAnswers ptbr scraper")

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	t := time.Now()

	// Logging
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-deep_ptBR.log", pwd, t.Format(YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create deep_ptBR.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	h.log.Println("file deep_ptBR.log created on: ", fileName)
	log := log.New(file, "DEEP-PTBR : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

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

func (h DeepptBRHandler) get(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to list all DeepAnswers ptbr content")

	qts, err := h.core.List()
	if err != nil {
		h.log.Printf("Error to list DeepAnswers Ptbr questions from db: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"DeepAnswers Ptbr questions: ": qts,
		})
	}
}

func (h DeepptBRHandler) getJSON(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to create a JSON file with DeepAnswers ptbr content")

	err := h.core.CreateJSON()
	if err != nil {
		h.log.Printf("Error to create a JSON file with DeepAnswers ptbr content: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "success",
		})
	}
}
