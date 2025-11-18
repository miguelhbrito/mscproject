package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miguelhbrito/mscproject/internal/commons"
	ha "github.com/miguelhbrito/mscproject/internal/core/ha_enus"
	data "github.com/miguelhbrito/mscproject/internal/data/ha_enus"
)

type HAenUSHandler struct {
	core ha.HAInt
	log  *log.Logger
}

func NewHAEnUSHandler(env *Env) *HAenUSHandler {

	db := data.HAEnglishPostgres{}

	core := ha.NewCore(
		db,
		env.log,
	)

	job := ha.NewHAenUSJob(env.log)
	err := job.Run(
		core,
	)
	if err != nil {
		log.Println("fail to init ha_enUS job")
	}

	return &HAenUSHandler{
		core: core,
		log:  env.log,
	}
}

func (h HAenUSHandler) post(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to starting HA english scraper")

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	t := time.Now()

	// Logging
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-ha_enUS.log", pwd, t.Format(YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create ha_enUS.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()

	h.log.Println("file ha_enUS.log created on: ", fileName)
	log := log.New(file, "HIDDEN_ANSWERS-ENUS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

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

func (h HAenUSHandler) get(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to list all HA english content")

	qts, err := h.core.List()
	if err != nil {
		h.log.Printf("Error to list Hidden Answers questions from db: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Hidden Answers questions: ": qts,
		})
	}
}

func (h HAenUSHandler) getJSON(c *gin.Context) {
	h.log.Printf("%s %s -> %s", c.Request.Method, c.Request.URL, c.Request.RemoteAddr)
	h.log.Printf("receive request to create a JSON file with HA english content")

	err := h.core.CreateJSON()
	if err != nil {
		h.log.Printf("Error to create a JSON file with HA english content: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "success",
		})
	}
}
