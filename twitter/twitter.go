package twitter

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed auth.json
var auth []byte

const (
	POST_TWEET_PATH = "https://api.twitter.com/2/tweets"
	POST_IMAGE_PATH = "https://upload.twitter.com/1.1/media/upload.json"
)

type tweet struct {
	Text  string `json:"text"`
	Media struct {
		MediaIds []string `json:"media_ids"`
	} `json:"media"`
}

type postImageResponse struct {
	MediaID          int64  `json:"media_id"`
	MediaIDString    string `json:"media_id_string"`
	Size             int    `json:"size"`
	ExpiresAfterSecs int    `json:"expires_after_secs"`
	Image            struct {
		ImageType string `json:"image_type"`
		W         int    `json:"w"`
		H         int    `json:"h"`
	} `json:"image"`
}

type twitterAuth struct {
	ConsumerKey          string
	ConsumerSecret       string
	AccessToken          string
	TokenSecret          string
	OAuthSignatureMethod string
	OAuthVersion         string
	OAuthTimestamp       int
	OAuthNonce           string
	OAuthSignature       string
}

func PostTweet(title, author string, mediaids []string) error {
	t := tweet{
		Text: fmt.Sprintf("作品名 %s\n著者   %s", title, author),
		Media: struct {
			MediaIds []string "json:\"media_ids\""
		}{
			MediaIds: mediaids,
		},
	}

	j, err := json.Marshal(t)
	if err != nil {
		return err
	}

	payload := strings.NewReader(string(j))

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, POST_TWEET_PATH, payload)
	if err != nil {
		return err
	}

	ah, err := generateAuthorizationHeader(http.MethodPost, POST_TWEET_PATH)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", ah)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("cloud not post tweet : %d %s", res.StatusCode, string(body))
	}

	return nil
}

func PostImage(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	payload := &bytes.Buffer{}
	mw := multipart.NewWriter(payload)
	fw, err := mw.CreateFormFile("media", string(path[2:]))
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return "", err
	}

	ct := mw.FormDataContentType()
	err = mw.Close()
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, POST_IMAGE_PATH, payload)
	if err != nil {
		return "", err
	}

	ah, err := generateAuthorizationHeader(http.MethodPost, POST_IMAGE_PATH)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", ah)
	req.Header.Add("Content-Type", ct)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloud not post image : %d %s", res.StatusCode, string(body))
	}

	pir := postImageResponse{}
	if err := json.Unmarshal(body, &pir); err != nil {
		return "", err
	}
	return pir.MediaIDString, nil
}

func generateOAuthSignature(ta *twitterAuth, method, requrl string) (string, error) {
	parm := fmt.Sprintf(
		"oauth_consumer_key=%s&oauth_nonce=%s&oauth_signature_method=%s&oauth_timestamp=%d&oauth_token=%s&oauth_version=%s",
		ta.ConsumerKey, ta.OAuthNonce, ta.OAuthSignatureMethod, ta.OAuthTimestamp, ta.AccessToken, ta.OAuthVersion,
	)
	base := fmt.Sprintf("%s&%s&%s", url.QueryEscape(method), url.QueryEscape(requrl), url.QueryEscape(parm))
	secret := fmt.Sprintf("%s&%s", url.QueryEscape(ta.ConsumerSecret), url.QueryEscape(ta.TokenSecret))

	key := []byte(secret)
	hash := hmac.New(sha1.New, key)
	_, err := hash.Write([]byte(base))
	if err != nil {
		return "", err
	}

	return url.QueryEscape(base64.StdEncoding.EncodeToString(hash.Sum(nil))), nil
}

func generateOAuthNonce() (nonce string) {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < 11; i++ {
		nonce += string(str[r.Intn(len(str)-1)])
	}
	return nonce
}

func generateAuthorizationHeader(mehod, requrl string) (ah string, err error) {
	ta := twitterAuth{}
	//認証情報をセット
	if err = json.Unmarshal(auth, &ta); err != nil {
		return "", err
	}
	ta.OAuthTimestamp = int(time.Now().Unix())
	ta.OAuthNonce = generateOAuthNonce()
	ta.OAuthSignature, err = generateOAuthSignature(&ta, mehod, requrl)
	if err != nil {
		return "", err
	}

	ah = fmt.Sprintf(
		"OAuth oauth_consumer_key=\"%s\",oauth_token=\"%s\",oauth_signature_method=\"%s\",oauth_timestamp=\"%d\",oauth_nonce=\"%s\",oauth_version=\"%s\",oauth_signature=\"%s\"",
		ta.ConsumerKey, ta.AccessToken, ta.OAuthSignatureMethod, ta.OAuthTimestamp, ta.OAuthNonce, ta.OAuthVersion, ta.OAuthSignature,
	)
	return ah, nil
}
