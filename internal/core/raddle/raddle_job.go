package raddle

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/miguelhbrito/mscproject/internal/commons"
	"github.com/robfig/cron/v3"
)

type RaJob interface {
	Cron() string
	Name() string
	Run(ra RaInt) error
}

type raJob struct {
	log *log.Logger
}

func NewRaddleJob(log *log.Logger) RaJob {
	return raJob{
		log: log,
	}
}

func (rj raJob) Cron() string {
	return "0 0 * * 6"
}

func (rj raJob) Name() string {
	return "Cron Job to execute a scrapper to Raddle"
}

func (rj raJob) Run(ra RaInt) error {

	cron := cron.New()
	_, err := cron.AddFunc(rj.Cron(), func() {
		rj.runJob(ra)
	})

	if err != nil {
		rj.log.Println("fail to init job:", rj.Name())
		return err
	}

	cron.Start()
	rj.log.Printf("Init job: '%s' with cron: '%s'\n", rj.Name(), rj.Cron())

	return nil
}

func (rj raJob) runJob(ra RaInt) {
	rj.log.Println("Running job:", rj.Name())

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	// Logging
	t := time.Now()
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-raddle.log", pwd, t.Format(commons.YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create raddle.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	log := log.New(file, "RADDLE : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Channels done and errorCh
	done := make(chan commons.Done)
	errorCh := make(chan commons.Error)

	go ra.Scrapper(done, errorCh, log)
	select {
	case <-done:
		rj.log.Println("success:", <-done)
	case e := <-errorCh:
		rj.log.Println("failure:", e, <-errorCh)
		commons.SendEmail(rj.log)
		time.Sleep(6 * time.Hour)
		rj.runJob(ra)
	}
}
