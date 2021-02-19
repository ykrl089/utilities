package weixin

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ykrl089/utilities/str/uid"
)

type AccessToken struct {
	Token     string    `json:"access_token"`
	expiredAt time.Time `json:"-"`
}
type WxJsTicket struct {
	ErrCode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Ticket  string `json:"ticket"`
}

var token *AccessToken
var ticket *WxJsTicket

func getToken(appid string, sercet string) error {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appid, sercet)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, token); err != nil {
		return err
	}
	if token.Token == "" {
		return errors.New("获取Token失败")
	}
	token.expiredAt = time.Now().Add(time.Hour)
	return nil
}
func GetTicket(appid string, sercet string) error {
	if token == nil || token.Token == "" || token.expiredAt.After(time.Now()) {
		if err := getToken(appid, sercet); err != nil {
			return err
		}
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi", token.Token)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, ticket); err != nil {
		return err
	}
	if ticket.ErrCode != 0 && ticket.Ticket == "" {
		return errors.New("获取微信JS失败")
	}
	return nil
}

type WxConfig struct {
	Appid     string `json:"appId"` //微信sdk id 必须填写
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

func (config *WxConfig) Get(sercet string, url string) error {
	if i := strings.IndexByte(url, '#'); i >= 0 {
		url = url[:i]
	}
	config.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	config.NonceStr = uid.Rand(16)
	if config.NonceStr == "" {
		return errors.New("WxJSConfig：生成随机数失败")
	}
	if ticket == nil {
		if err := GetTicket(config.Appid, sercet); err != nil {
			return err
		}
	}
	n := len("jsapi_ticket=") + len(ticket.Ticket) +
		len("&noncestr=") + len(config.NonceStr) +
		len("&timestamp=") + len(config.Timestamp) +
		len("&url=") + len(url)
	buf := make([]byte, 0, n)

	buf = append(buf, "jsapi_ticket="...)
	buf = append(buf, ticket.Ticket...)
	buf = append(buf, "&noncestr="...)
	buf = append(buf, config.NonceStr...)
	buf = append(buf, "&timestamp="...)
	buf = append(buf, config.Timestamp...)
	buf = append(buf, "&url="...)
	buf = append(buf, url...)

	hashsum := sha1.Sum(buf)
	config.Signature = hex.EncodeToString(hashsum[:])
	return nil
}
