package haenus

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/miguelhbrito/mscproject/internal/commons"
	"github.com/robfig/cron/v3"
)

type HAJob interface {
	Cron() string
	Name() string
	Run(ha HAInt) error
}

type haJob struct {
	log *log.Logger
}

func NewHAenUSJob(log *log.Logger) HAJob {
	return haJob{
		log: log,
	}
}

func (hj haJob) Cron() string {
	return "0 0 * * 2"
}

func (hj haJob) Name() string {
	return "Cron Job to execute a scrapper to Hidden Answers ver. En"
}

func (hj haJob) Run(ha HAInt) error {

	cron := cron.New()
	_, err := cron.AddFunc(hj.Cron(), func() {
		hj.runJob(ha)
	})

	if err != nil {
		hj.log.Println("fail to init job:", hj.Name())
		return err
	}

	cron.Start()
	hj.log.Printf("Init job: '%s' with cron: '%s'\n", hj.Name(), hj.Cron())

	return nil
}

func (hj haJob) runJob(ha HAInt) {
	hj.log.Println("Running job:", hj.Name())

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	// Logging
	t := time.Now()
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-ha_enUS.log", pwd, t.Format(commons.YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create ha_enUS.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	log := log.New(file, "HIDDEN_ANSWERS-ENUS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Channels done and errorCh
	done := make(chan commons.Done)
	errorCh := make(chan commons.Error)

	go ha.Scrapper(done, errorCh, log)
	select {
	case <-done:
		hj.log.Println("success:", <-done)
	case e := <-errorCh:
		hj.log.Println("failure:", e, <-errorCh)
		commons.SendEmail(hj.log)
		time.Sleep(6 * time.Hour)
		hj.runJob(ha)
	}
}
