//Package isi is alibaba Intelligent Speech Interaction go sdk
package isi

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

//Client isi client
type Client struct {
	ak string
	sk string

	skbs []byte
}

//New retrurn new isi client
func New(ak, sk string) *Client {
	return &Client{ak: ak, sk: sk, skbs: []byte(sk)}
}

func (c Client) gmt() string {
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		return ""
	}
	return time.Now().In(loc).Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

func (c Client) md5base64(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	h := md5.New()
	h.Write(bs)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c Client) sha1base64(txt string) string {
	h := hmac.New(sha1.New, c.skbs)
	h.Write([]byte(txt))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c Client) sendRequest(method, url string, body []byte, contentType string) []byte {

	var (
		gmt     = c.gmt()
		md5body = c.md5base64(body)
	)

	if strings.Contains(url, "https://nlsapi.aliyun.com/recognize") {
		md5body = c.md5base64([]byte(md5body))
	}

	if contentType == "" {
		contentType = "application/json"
	}

	sign := c.sha1base64(method + "\napplication/json\n" + md5body + "\n" + contentType + "\n" + gmt)
	client := &http.Client{}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
	}
	request.Header.Set("Authorization", "Dataplus "+c.ak+":"+sign)
	request.Header.Set("Content-type", contentType)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Date", gmt)
	response, _ := client.Do(request)
	respBody, _ := ioutil.ReadAll(response.Body)
	// fmt.Println(string(respBody))
	return respBody
}

//Get return get url http response body
func (c Client) Get(url string) []byte {
	return c.sendRequest("GET", url, nil, "")
}

// GetVocabs get vocabs
// {"request_id":"669865051b154e588ba29dcba0e8b2ae","global_weight":1,"words":["香蕉","牛奶","哈密瓜"],"word_weights":null}
func (c Client) GetVocabs(id string) Vocab {
	vocab := Vocab{}
	json.Unmarshal(c.sendRequest("GET", "https://nlsapi.aliyun.com/asr/custom/vocabs/"+id, nil, ""), &vocab)
	return vocab
}

// CreateVocabs return created vocabs id
// {"request_id":"d3b0f95d700d4bbd81db8090b34e670d","vocabulary_id":"85ea46c8a7c1456685b0df479d929477"}
func (c Client) CreateVocabs(v Vocab) string {
	bs, _ := json.Marshal(&v)
	ret := vocabRet{}
	json.Unmarshal(c.sendRequest("POST", "https://nlsapi.aliyun.com/asr/custom/vocabs", bs, ""), &ret)
	return ret.VocabularyID
}

// UpdateVocabs update id with vocab
// {"request_id":"96a9478fa4364db8a2b623eb2f50ddb2"}
// {"id":"856271ec9e9942c6b932fbdfc8af060e","request_id":"856271ec9e9942c6b932fbdfc8af060e","error_code":100307,"error_message":"\"VOCABULARY_NOT_FOUND(Vocabulary not exist! oid=1744412405438088, vid=85ea46c8a7c1456685b0df479d929472)\""}
func (c Client) UpdateVocabs(id string, v Vocab) bool {
	bs, _ := json.Marshal(&v)
	ret := vocabRet{}
	json.Unmarshal(c.sendRequest("PUT", "https://nlsapi.aliyun.com/asr/custom/vocabs/"+id, bs, ""), &ret)
	if ret.ErrorCode == 0 {
		return true
	}
	log.Println(ret.ErrorMessage)
	return false
}

// DeleteVocabs delete the id table
// {"request_id":"a6ad7f0436f4491b84601045affa9d82"}
// {"id":"9a3c74d2ffdb4b6d952de97c531dda90","request_id":"9a3c74d2ffdb4b6d952de97c531dda90","error_code":100307,"error_message":"\"VOCABULARY_NOT_FOUND(Vocabulary not exist! oid=1744412405438088, vid=85ea46c8a7c1456685b0df479d929472)\""}
func (c Client) DeleteVocabs(id string) bool {
	ret := vocabRet{}
	json.Unmarshal(c.sendRequest("DELETE", "https://nlsapi.aliyun.com/asr/custom/vocabs/"+id, nil, ""), &ret)
	if ret.ErrorCode == 0 {
		return true
	}
	log.Println(ret.ErrorMessage)
	return false
}

// Recognize one sentence recognize
// {"result":"后天下午5点25分，整形外科需要手术账号36，携带传送到手术室。","request_id":"51e1005bf454476a91319d5bd311075e"}
func (c Client) Recognize(localPath string, models ...string) (string, error) {
	model := ""
	if len(models) > 0 {
		model = models[0]
	} else {
		model = "customer-service"
	}
	bs, err := ioutil.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	ret := vocabRet{}
	// 这里不同文件要做不同的识别，但是我只用pcm，所以不管了。
	contentType := "audio/pcm; samplerate=16000"
	json.Unmarshal(c.sendRequest("POST", "https://nlsapi.aliyun.com/recognize?version=2.0&model="+model, bs, contentType), &ret)
	if ret.ErrorCode == 0 {
		return ret.Result, nil
	}
	return "", errors.New(ret.ErrorMessage)
}

// Transcriptions translate record file
func (c Client) Transcriptions(ossPath, vobID string) string {

	tb := TransBody{}
	tb.OssLink = ossPath
	tb.AppKey = "nls-service-multi-domain"
	tb.VocabularyID = vobID
	bs, _ := json.Marshal(&tb)

	id := transID{}
	json.Unmarshal(c.sendRequest("POST", "https://nlsapi.aliyun.com/transcriptions", bs, ""), &id)

	ret := transRet{}
	for {
		json.Unmarshal(c.sendRequest("GET", "https://nlsapi.aliyun.com/transcriptions/"+id.ID, nil, ""), &ret)
		if ret.Status == "SUCCEED" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	text := ret.Result[0].Text
	return text
}
