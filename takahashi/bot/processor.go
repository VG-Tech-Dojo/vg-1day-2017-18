package bot

import (
	"regexp"
	"strings"

	"fmt"

	"github.com/VG-Tech-Dojo/vg-1day-2017-18/takahashi/env"
	"github.com/VG-Tech-Dojo/vg-1day-2017-18/takahashi/model"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
)

type Output struct {
	Status       int    `json:"status"`
	Message     string `json:"message"`
	Results []struct {
		Perplexity float64 `json:"perplexity"`
		Reply string `json:"reply"`
	} `json:"results"`
}

const (
	keywordAPIURLFormat = "https://jlp.yahooapis.jp/KeyphraseService/V1/extract?appid=%s&sentence=%s&output=json"
)

type (
	// Processor はmessageを受け取り、投稿用messageを作るインターフェースです
	Processor interface {
		Process(message *model.Message) (*model.Message, error)
	}

	// HelloWorldProcessor は"hello, world!"メッセージを作るprocessorの構造体です
	HelloWorldProcessor struct{}

	// OmikujiProcessor は"大吉", "吉", "中吉", "小吉", "末吉", "凶"のいずれかをランダムで作るprocessorの構造体です
	OmikujiProcessor struct{}

	// KeywordProcessor はメッセージ本文からキーワードを抽出するprocessorの構造体です
	KeywordProcessor struct{}

	GachaProcessor struct{}

	TalkProcessor struct{}
)

// Process は"hello, world!"というbodyがセットされたメッセージのポインタを返します
func (p *HelloWorldProcessor) Process(msgIn *model.Message) (*model.Message, error) {
	return &model.Message{
		Body: msgIn.Body + ", world!",
	}, nil
}

func (p *GachaProcessor) Process(msgIN *model.Message) (*model.Message, error) {
	gacha := []string{
		"SSレア",
		"Sレア",
		"レア",
		"ノーマル",
	}
	result := gacha[randIntn(len(gacha))]
	return &model.Message{
		Body: result,
	}, nil
}

// Process は"大吉", "吉", "中吉", "小吉", "末吉", "凶"のいずれかがbodyにセットされたメッセージへのポインタを返します
func (p *OmikujiProcessor) Process(msgIn *model.Message) (*model.Message, error) {
	fortunes := []string{
		"大吉",
		"吉",
		"中吉",
		"小吉",
		"末吉",
		"凶",
	}
	result := fortunes[randIntn(len(fortunes))]
	return &model.Message{
		Body: result,
	}, nil
}

func (p *TalkProcessor) Process(msgIn *model.Message) (*model.Message, error) {
	r := regexp.MustCompile("\\Atalk (.*)\\z")
	matchedStrings := r.FindStringSubmatch(msgIn.Body)
	text := matchedStrings[1]

	client := &http.Client{}
	data := url.Values{"apikey":{"dPMQ92gZvtCYbvmk2kirZi9BzUCkAA5c"},"query":{text}}

	res, _ := client.Post(
		"https://api.a3rt.recruit-tech.co.jp/talk/v1/smalltalk",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	output := Output{}
	err = json.Unmarshal(body, &output)

	return &model.Message{
		Body: output.Results[0].Reply,
	}, nil
}

// Process はメッセージ本文からキーワードを抽出します
func (p *KeywordProcessor) Process(msgIn *model.Message) (*model.Message, error) {
	r := regexp.MustCompile("\\Akeyword (.*)\\z")
	matchedStrings := r.FindStringSubmatch(msgIn.Body)
	text := matchedStrings[1]

	url := fmt.Sprintf(keywordAPIURLFormat, env.KeywordAPIAppID, text)

	type keywordAPIResponse map[string]interface{}
	var json keywordAPIResponse
	get(url, &json)

	keywords := []string{}
	for k, v := range json {
		if k == "Error" {
			return nil, fmt.Errorf("%#v", v)
		}
		keywords = append(keywords, k)
	}

	return &model.Message{
		Body: "キーワード：" + strings.Join(keywords, ", "),
	}, nil
}
