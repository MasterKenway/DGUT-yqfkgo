package service

import (
	"time"
)

import (
	"DGUT-yqfkgo/internal/config"
	"DGUT-yqfkgo/internal/log"
	"DGUT-yqfkgo/internal/push"
)

import (
	"github.com/bitly/go-simplejson"
	"github.com/jasonlvhit/gocron"
)

var serv *Service

func initService(conf *config.Config) {
	serv = NewService(conf)
}

func Start(conf *config.Config) {
	initService(conf)
	task := func() {
		for {
			err := begin()
			if err != nil {
				if len(conf.TgBotToken) != 0 {
					err = push.Append(err.Error())
					if err != nil {
						log.Warn().Msg(err.Error())
					}
					err = push.Append("Run Task Again After 10 Seconds...")
					if err != nil {
						log.Warn().Msg(err.Error())
					}
				}
				log.Warn().Msg("Run Task Again After 10 Seconds...")
				err = push.Push()
				if err != nil {
					log.Warn().Msg(err.Error())
				}
				time.Sleep(time.Duration(10) * time.Second)
			} else {
				err = push.Push()
				if err != nil {
					log.Warn().Msg(err.Error())
				}
				break
			}
		}
	}

	task()
	s := gocron.NewScheduler()
	err := s.Every(1).Day().At(conf.RunAt).Do(task)
	if err != nil {
		log.Panic().Msgf("Schedule Task Failed, %v", err)
	}
	<-s.Start()
}

func begin() error {
	err := serv.Login()
	if err != nil {
		log.Warn().Msgf("Run Failed: Login Failed, %v", err)
		return err
	}
	log.Info().Msg("Logging Successfully")

	serializedFormData, err := serv.ReadPrePost()
	if err != nil {
		log.Warn().Msgf("Run Failed: Read Pre Post Data Failed, %v", err)
		return err
	}
	log.Info().Msg("Read Pre Post Data Successfully")

	json, err := simplejson.NewJson(serializedFormData)
	if err != nil {
		return err
	}

	json.Set("confirm", 1)
	json.Set("important_area", nil)
	//json.Set("current_region", nil)

	bytes, err := json.Encode()
	if err != nil {
		return err
	}

	err = serv.Post(bytes)
	if err != nil {
		log.Warn().Msgf("Post Data Failed, %v", err)
		return err
	}
	log.Info().Msg("Post Data Successfully")

	return nil
}
