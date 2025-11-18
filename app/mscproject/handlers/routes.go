package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	YYYYMMDD = "2006-01-02"
)

type Env struct {
	log *log.Logger
}

func NewGin(log *log.Logger, db *sql.DB) http.Handler {
	router := gin.New()

	env := &Env{
		log: log,
	}

	//Starting postgres and core
	ha_enUS := NewHAEnUSHandler(env)
	ha_ptBR := NewHAptBRHandler(env)
	deep_ptBR := NewDeepptBRHandler(env)
	deep_enUS := NewDeepenUSHandler(env)
	deep_esES := NewDeepesESHandler(env)
	raddle := NewRaddleHandler(env)

	router.POST("/ha-enus", ha_enUS.post)
	router.GET("/ha-enus", ha_enUS.get)
	router.GET("/ha-enus/JSON", ha_enUS.getJSON)
	router.POST("/ha-ptbr", ha_ptBR.post)
	router.GET("/ha-ptbr", ha_ptBR.get)
	router.GET("/ha-ptbr/JSON", ha_ptBR.getJSON)
	router.POST("/deep-ptbr", deep_ptBR.post)
	router.GET("/deep-ptbr", deep_ptBR.get)
	router.GET("/deep-ptbr/JSON", deep_ptBR.getJSON)
	router.POST("/deep-enus", deep_enUS.post)
	router.GET("/deep-enus", deep_enUS.get)
	router.GET("/deep-enus/JSON", deep_enUS.getJSON)
	router.POST("/deep-eses", deep_esES.post)
	router.GET("/deep-eses", deep_esES.get)
	router.GET("/deep-eses/JSON", deep_esES.getJSON)
	router.POST("/raddle", raddle.post)
	router.GET("/raddle", raddle.get)
	router.GET("/raddle/JSON", raddle.getJSON)

	return router
}
