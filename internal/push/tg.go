package push

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

import (
	"DGUT-yqfkgo/internal/config"
	"DGUT-yqfkgo/internal/log"
)

import (
	"github.com/bitly/go-simplejson"
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type tgPusher struct {
	TgBotToken string
	ChatId     string
	Text       string
}

func newTgPusher(conf *config.Config) *tgPusher {
	return &tgPusher{
		TgBotToken: conf.TgBotToken,
		ChatId:     conf.ChatId,
		Text:       "",
	}
}

type tgPusherWrapper struct {
	C  *http.Client
	Tp *tgPusher
}

var tpWrp *tgPusherWrapper

func NewTgPusherWrapper(conf *config.Config) *tgPusherWrapper {
	if tpWrp == nil {
		tpWrp = &tgPusherWrapper{
			C:  &http.Client{Jar: &cookiejar.Jar{}},
			Tp: newTgPusher(conf),
		}
	}
	return tpWrp
}

func Push() error {
	if tpWrp == nil {
		return errors.New("Telegram Push Msg Error, TgWrp Not Init")
	}

	json := simplejson.New()
	json.Set("chat_id", tpWrp.Tp.ChatId)
	json.Set("text", tpWrp.Tp.Text)
	data, err := json.Encode()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+tpWrp.Tp.TgBotToken+"/sendMessage", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	req.Close = true
	resp, err := tpWrp.C.Do(req)
	if err != nil {
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	ok, err := jsonparser.GetBoolean(body, "ok")
	if err != nil {
		return err
	}
	if ok {
		log.Info().Msg("Telegram Push Msg Successfully")
	}
	clear()
	return nil
}

func Append(msg string) error {
	if tpWrp == nil {
		return errors.New("Telegram Push Msg Error, TgWrp Not Init")
	}
	tpWrp.Tp.Text += msg + "\n\n"
	return nil
}

func clear() {
	tpWrp.Tp.Text = ""
}
