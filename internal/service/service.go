package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"unsafe"
)

import (
	"DGUT-yqfkgo/internal/config"
	"DGUT-yqfkgo/internal/log"
	"DGUT-yqfkgo/internal/push"
)

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

const (
	LOGIN_URL         = "https://cas.dgut.edu.cn/home/Oauth/getToken/appid/illnessProtectionHome/state/home.html"
	PRE_POST_DATA_URL = "https://yqfk.dgut.edu.cn/home/base_info/getBaseInfo"
	POST_DATA_URL     = "https://yqfk.dgut.edu.cn/home/base_info/addBaseInfo"

	AUTH_HEADER = "authorization"
)

type Service struct {
	c     *http.Client
	Conf  config.Config
	Token string
}

func NewService(conf *config.Config) *Service {
	jar, _ := cookiejar.New(nil)

	return &Service{
		Conf: *conf,
		c:    &http.Client{Jar: jar},
	}
}

func (s *Service) Login() error {
	XssToken, err := s.getXssToken()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("username", s.Conf.Username)
	params.Set("password", s.Conf.Password)
	params.Set("__token__", XssToken)

	req, err := http.NewRequest("POST", LOGIN_URL, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	code, err := jsonparser.GetInt(respBody, "code")
	if err != nil {
		return err
	}

	if code != 1 {
		msg, _ := jsonparser.GetString(respBody, "message")
		return errors.New(msg)
	}

	urlInfo, err := jsonparser.GetString(respBody, "info")
	if err != nil {
		return err
	}

	accessToken, err := s.getAccessToken(urlInfo)
	if err != nil {
		return err
	}

	s.Token = "Bearer " + accessToken
	log.Debug().Msgf("AccessToken: %s", accessToken)
	return nil
}

func (s *Service) getXssToken() (string, error) {
	req, _ := http.NewRequest("GET", LOGIN_URL, nil)
	req.Close = true

	resp, err := s.c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`var token = "(.*?)";`)
	res := re.FindAllStringSubmatch(*(*string)(unsafe.Pointer(&contents)), -1)
	token := res[0][1]
	if len(token) == 0 {
		return "", errors.New("Token Not Found")
	}

	return token, nil
}

func (s *Service) getAccessToken(urlInfo string) (string, error) {
	resp, err := s.c.Get(urlInfo)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	re := regexp.MustCompile(`access_token=(.*?)$`)
	accessToken := re.FindAllStringSubmatch(resp.Request.URL.String(), -1)[0][1]
	if len(accessToken) == 0 {
		return "", errors.New("Access Token Not Found")
	}

	return accessToken, nil
}

func (s *Service) ReadPrePost() ([]byte, error) {
	req, err := http.NewRequest("GET", PRE_POST_DATA_URL, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	//req.Header = http.Header{
	//	AUTH_HEADER: []string{s.Token},
	//}
	req.Header.Set(AUTH_HEADER, s.Token)

	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if code, err := jsonparser.GetInt(respBody, "code"); err != nil {
		return nil, err
	} else {
		if code != 200 {
			msg, _ := jsonparser.GetString(respBody, "message")
			return nil, errors.New(msg)
		}
	}

	info, _, _, err := jsonparser.Get(respBody, "info")
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Service) Post(postData []byte) error {
	req, _ := http.NewRequest("POST", POST_DATA_URL, bytes.NewReader(postData))
	req.Close = true
	req.Header.Set(AUTH_HEADER, s.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	code, err := jsonparser.GetInt(respBody, "code")
	if err != nil {
		return err
	}
	msg, err := jsonparser.GetString(respBody, "message")
	if err != nil {
		return err
	}

	if len(s.Conf.TgBotToken) != 0 {
		err = push.Append(msg)
		if err != nil {
			return err
		}
	}
	log.Info().Msg(msg)
	if code != 200 {
		if code == 400 {
			if msg == "今日已提交，请勿重复操作" {
				return nil
			}
		}
		return errors.New(msg)
	}
	err = push.Push()
	if err != nil {
		return err
	}

	return nil
}
