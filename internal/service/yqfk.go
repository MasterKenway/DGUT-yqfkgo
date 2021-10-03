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
	local, _ := time.LoadLocation("Asia/Shanghai")
	task := func() {
		for i := 0; i < 20; i++ {
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

				time.Sleep(time.Duration(10) * time.Second)
			} else {
				_, t := gocron.NextRun()
				now := time.Now().In(local).Format(time.RFC1123Z)
				_ = push.Append("Finished Time: " + now)
				log.Info().Msgf("Finished Time: " + now)
				log.Info().Msgf("Next Time to Run: %s", t.String())
				break
			}
		}

		if len(conf.TgBotToken) != 0 {
			for j := 0; j < 5; j++ {
				err := push.Push()
				if err != nil {
					log.Warn().Msg(err.Error())
				} else {
					break
				}
				time.Sleep(time.Duration(1) * time.Second)
			}
			push.Clear()
		}
	}

	secondsEastOfUTC := int((8 * time.Hour).Seconds())
	beijing := time.FixedZone("Beijing Time", secondsEastOfUTC)
	gocron.ChangeLoc(beijing)
	err := gocron.Every(1).Day().At(conf.RunAt).From(gocron.NextTick()).Do(task)
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
