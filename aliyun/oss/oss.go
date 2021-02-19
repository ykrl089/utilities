/*
* @Author: GuoDi
* @Date:   2016-04-11 22:13:41
* @Last Modified by:   guodi
* @Last Modified time: 2021-02-19 23:01:12
 */
package aliyun

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"mime/multipart"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Oss 阿里云OSS对象
type Oss struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	client          *oss.Client
	bucket          *oss.Bucket
}

// Init 初始化OSS
// func (o *Oss) Init() error {
// 	var err error
// 	if o.client, err = oss.New(o.AccessKeyID, o.AccessKeySecret, o.BucketName); err != nil {
// 		return err
// 	}

// 	return o.ChangeBucket()
// }

// ChangeBucket 切换Bucket
func (o *Oss) ChangeBucket() error {
	if o.BucketName == "" {
		return errors.New("Bucket 名称错误")
	}
	var err error
	if o.bucket, err = o.client.Bucket(o.BucketName); err != nil {
		return err

	}
	return nil
}

//PutFile 添加文件
func (o *Oss) PutFile(name string, file string) error {

	return o.bucket.PutObjectFromFile(name, file)
}

//PutObject 添加[]byte 类型对象
func (o *Oss) PutObject(name string, obj []byte) error {
	return o.bucket.PutObject(name, bytes.NewReader(obj))
}

//PutMimeFile 添加媒体文件
func (o *Oss) PutMimeFile(name string, fd multipart.File) error {

	return o.bucket.PutObject(name, fd)
}

// Exist 检查文件是否存在
func (o *Oss) Exist(name string) bool {
	if isExist, err := o.bucket.IsObjectExist(name); err == nil {
		return isExist
	} else {
		fmt.Println(err.Error())
		return false
	}

}

// Delete 删除文件
func (o *Oss) Delete(name string) error {
	return o.bucket.DeleteObject(name)
}

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

type PolicyToken struct {
	AccessKeyId string `json:"accessid"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
}

func (o *Oss) Token(upload_dir string) PolicyToken {
	host := o.Endpoint[:6] + o.BucketName + o.Endpoint[6:]
	//create post policy json
	var config ConfigStruct
	expired := time.Now().Add(time.Hour)
	config.Expiration = expired.Format("2006-01-02T15:04:05Z")
	var condition []string
	condition = append(condition, "starts-with")
	condition = append(condition, "$key")
	condition = append(condition, upload_dir)
	config.Conditions = append(config.Conditions, condition)

	//calucate signature
	result, _ := json.Marshal(config)
	debyte := base64.StdEncoding.EncodeToString(result)
	signedStr := signEncode(debyte, o.AccessKeySecret)

	var policyToken PolicyToken
	policyToken.AccessKeyId = o.AccessKeyID
	policyToken.Host = host
	policyToken.Expire = expired.Unix()
	policyToken.Signature = string(signedStr)
	policyToken.Directory = upload_dir
	policyToken.Policy = string(debyte)
	return policyToken
}
func signEncode(str string, accessKeySecret string) string {
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(accessKeySecret))
	io.WriteString(h, str)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
