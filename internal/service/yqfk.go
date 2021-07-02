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
	task := func() {
		for i := 0; i < 5; i++ {
			initService(conf)
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
				_, t := gocron.NextRun()
				_ = push.Append("Next Time to Run: " + t.String())
				log.Info().Msgf("Next Time to Run: %s", t.String())
				err = push.Push()
				if err != nil {
					log.Warn().Msg(err.Error())
				}
				break
			}
		}
	}

	secondsEastOfUTC := int((8 * time.Hour).Seconds())
	beijing := time.FixedZone("Beijing Time", secondsEastOfUTC)
	gocron.ChangeLoc(beijing)
	err := gocron.Every(1).Day().At("5:00").From(gocron.NextTick()).Do(task)
	if err != nil {
		log.Panic().Msgf("Schedule Task Failed, %v", err)
	}
	<-gocron.Start()
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
	json.Set("acid_test_results", nil)
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
