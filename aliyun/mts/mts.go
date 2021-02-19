package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"utilities/str/uid"
)

const (
	base64Table = "123QRSTUabcdVWXYZHijKLAWDCABDstEFGuvwxyzGHIJklmnopqr234560178912"
)

var coder = base64.NewEncoding(base64Table)

func base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

func get_gmt_iso8601(expire_end int64) string {
	var tokenExpire = time.Unix(expire_end, 0).Format("2006-01-02T15:04:05Z")
	return tokenExpire
}

func signEncode(str string, accessKeySecret string) string {
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(accessKeySecret))
	io.WriteString(h, str)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type MTS struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
}
type MtsInput struct {
	Bucket   string `json:"Bucket"`
	Location string `json:"Location"`
	Object   string `json:"Object"`
}
type MtsOutput struct {
	OutputObject string `json:"OutputObject"`
	TemplateId   string `json:"TemplateId"`
}
type JobOutput struct {
	Properties Properties `json:"Properties"`
}
type Properties struct {
	FileSize   string `json:"FileSize"`
	FileFormat string `json:"FileFormat"`
	Duration   string `json:"Duration"`
	Height     string `json:"Height"`
	Width      string `json:"Width"`
}
type Job struct {
	JobId        string    `json:"JobId"`
	Input        MtsInput  `json:"Input"`
	TemplateId   string    `json:"TemplateId"`
	State        string    `json:"State"`
	Code         string    `json:"Code"`
	Message      string    `json:"Message"`
	Percent      int       `json:"Percent"`
	PipelineId   string    `json:"PipelineId"`
	CreationTime string    `json:"CreationTime"`
	Output       JobOutput `json:"Output"`
}
type JobResult struct {
	Success bool   `json:"Success"`
	Code    string `json:"Code"`
	Message string `json:"Message"`
	Job     Job    `json:"Job"`
}
type JobResultList struct {
	JobResult []JobResult `json:"JobResult"`
}
type SubmitJobResult struct {
	RequestId     string        `json:"RequestId"`
	JobResultList JobResultList `json:"JobResultList"`
	Message       string        `json:"Message"`
}
type JobList struct {
	Job []Job `json:"Job"`
}
type QueryJobResult struct {
	RequestId string  `json:"RequestId"`
	JobList   JobList `json:"JobList"`
	Message   string  `json:"Message"`
}

// 提交转码作业
func (this *MTS) SubmitJobs(input MtsInput, outputs []MtsOutput) (jobId string, err error) {

	inputJson, _ := json.Marshal(input)
	outputsJson, _ := json.Marshal(outputs)
	jobParams := map[string]string{
		"Action":       "SubmitJobs",
		"Input":        string(inputJson),
		"Outputs":      string(outputsJson),
		"OutputBucket": "cst-media",
		"PipelineId":   "98bfac8902b04607bdd953e2b28a6dc8",
	}
	queryResult, httpErr := this.Req(jobParams, "GET")
	if httpErr != nil {
		return "", httpErr
	}
	var result SubmitJobResult
	json.Unmarshal(queryResult, &result)
	if len(queryResult) == 0 {
		return "", errors.New("返回错误")
	}
	if result.Message != "" {

		return "", errors.New(result.Message)
	}
	if result.JobResultList.JobResult[0].Success {
		return result.JobResultList.JobResult[0].Job.JobId, nil
	} else {
		return "", errors.New(result.JobResultList.JobResult[0].Message)
	}
}

//QueryJobs 转码查询
func (this *MTS) QueryJobs(jobIds string) (completed bool, pts Job, err error) {

	queryResult, httpErr := this.Req(map[string]string{
		"Action": "QueryJobList",
		"JobIds": jobIds,
	}, "GET")
	if httpErr != nil {
		return false, pts, httpErr
	}
	var result QueryJobResult
	json.Unmarshal(queryResult, &result)
	if len(queryResult) == 0 {
		return false, pts, errors.New("返回错误")
	}
	if result.Message != "" {
		return false, pts, errors.New(result.Message)
	}

	if len(result.JobList.Job) > 0 {
		if result.JobList.Job[0].State == "TranscodeSuccess" {
			return true, result.JobList.Job[0], nil
		} else if result.JobList.Job[0].State == "TranscodeFail" {
			return false, pts, errors.New(result.JobList.Job[0].Message)
		} else {
			return false, pts, nil
		}
	} else {
		return false, pts, errors.New("查询不到JobId")
	}

}
func (this *MTS) CancelJobs(jobIds string) error {

	queryResult, httpErr := this.Req(map[string]string{
		"Action": "CancelJob",
		"JobIds": jobIds,
	}, "GET")
	if httpErr != nil {
		return httpErr
	}
	var result QueryJobResult
	json.Unmarshal(queryResult, &result)
	if len(queryResult) == 0 {
		return errors.New("返回错误")
	}
	if result.Message != "" {
		return errors.New(result.Message)
	}
	return nil

}

// 媒体转码和查询URL生成
//
func (this *MTS) Req(params map[string]string, httpMethod string) (result []byte, err error) {
	AccessKeySecret := this.AccessKeySecret + "&"                         // 密钥
	Endpoint := this.Endpoint                                             //请求路径
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z") // 时间戳
	params["Format"] = "JSON"                                             //返回值的类型，支持JSON
	params["AccessKeyId"] = this.AccessKeyId                              //密钥ID
	params["Version"] = "2014-06-18"                                      //API版本号，为日期形式：YYYY-MM-DD
	params["SignatureMethod"] = "HMAC-SHA1"                               //签名方式
	params["SignatureVersion"] = "1.0"                                    //签名算法版本
	params["SignatureNonce"] = uid.UUID()                                 //唯一随机数
	paramsKeys := make([]string, len(params))
	i := 0
	for k, _ := range params {
		paramsKeys[i] = k
		i++
	}
	sort.Strings(paramsKeys)

	for k, v := range paramsKeys {
		paramsKeys[k] = v + "=" + urlEncode(params[v])
	}
	reqUrl := strings.Join(paramsKeys, "&")
	signToString := strings.ToUpper(httpMethod) + "&%2F&" + urlEncode(reqUrl)
	signature := signEncode(signToString, AccessKeySecret)
	reqUrl = Endpoint + "?Signature=" + urlEncode(signature) + "&" + reqUrl
	resp, err1 := http.Get(reqUrl)
	if err1 != nil {
		fmt.Println(err1.Error())
		return result, err1
	}
	defer resp.Body.Close()
	if res, err2 := ioutil.ReadAll(resp.Body); err2 == nil {
		return res, err2
	} else {
		return result, err2
	}
	return
}

// url 转码
func urlEncode(value string) string {
	value = url.QueryEscape(value)
	value = strings.Replace(value, "+", "%20", -1)
	value = strings.Replace(value, "~", "%7E", -1)
	return value
}
