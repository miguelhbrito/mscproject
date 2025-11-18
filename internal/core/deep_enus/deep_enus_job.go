package deepenus

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/miguelhbrito/mscproject/internal/commons"
	"github.com/robfig/cron/v3"
)

type DeepJob interface {
	Cron() string
	Name() string
	Run(das DAInt) error
}

type deepJob struct {
	log *log.Logger
}

func NewDeepJob(log *log.Logger) DeepJob {
	return deepJob{
		log: log,
	}
}

func (dj deepJob) Cron() string {
	return "0 0 * * 4"
}

func (dj deepJob) Name() string {
	return "Cron Job to execute a scrapper to DeepAnswers ver. English-US"
}

func (dj deepJob) Run(ds DAInt) error {

	cron := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(dj.log)))
	_, err := cron.AddFunc(dj.Cron(), func() {
		dj.runJob(ds)
	})

	if err != nil {
		dj.log.Println("fail to init job:", dj.Name())
		return err
	}

	cron.Start()
	dj.log.Printf("Init job: '%s' with cron: '%s'\n", dj.Name(), dj.Cron())

	return nil
}

func (dj deepJob) runJob(ds DAInt) {
	dj.log.Println("Running job:", dj.Name())

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("error on get pwd: %v", err)
	}

	// Logging
	t := time.Now()
	// open/create file
	fileName := fmt.Sprintf("%s/logs/%s-deep_enUS.log", pwd, t.Format(commons.YYYYMMDD))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open/create deep_enUS.log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %v", err)
		}
	}()
	log := log.New(file, "DEEP-ENUS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Channels done and errorCh
	done := make(chan commons.Done)
	errorCh := make(chan commons.Error)

	go ds.Scrapper(done, errorCh, log)
	select {
	case <-done:
		dj.log.Println("success:", <-done)
	case e := <-errorCh:
		dj.log.Println("failure:", e, <-errorCh)
		commons.SendEmail(dj.log)
		time.Sleep(6 * time.Hour)
		dj.runJob(ds)
	}
}
