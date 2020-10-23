package twocaptcha

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var requestURL string = "https://2captcha.com/in.php"
var responseURL string = "https://2captcha.com/res.php"

//Client main
type Client struct {
	APIKey string
	Client *http.Client
}

//SolveImageCaptcha solve image captcha
func (c *Client) SolveImageCaptcha(base64Str string) (string, error) {
	requestForm := url.Values{}
	requestForm.Add("method", "base64")
	requestForm.Add("body", base64Str)
	requestForm.Add("soft_id", "2099")

	id, err := c.sendRequest(requestForm, requestURL, 20, 5*time.Second)
	if err != nil {
		return "", err
	}

	responseForm := url.Values{}
	responseForm.Add("action", "get")
	responseForm.Add("id", id)
	return c.sendRequest(responseForm, responseURL, 20, 5*time.Second)
}

//SolveRecaptchaV2 solve recaptcha v2
func (c *Client) SolveRecaptchaV2(siteURL, recaptchaKey string) (string, error) {
	requestForm := url.Values{}
	requestForm.Add("googlekey", recaptchaKey)
	requestForm.Add("pageurl", siteURL)
	requestForm.Add("method", "userrecaptcha")
	requestForm.Add("soft_id", "2099")

	id, err := c.sendRequest(requestForm, requestURL, 20, 5*time.Second)
	if err != nil {
		return "", err
	}

	responseForm := url.Values{}
	responseForm.Add("id", id)
	responseForm.Add("action", "get")
	return c.sendRequest(responseForm, responseURL, 20, 5*time.Second)
}

func (c *Client) sendRequest(form url.Values, URL string, retry int, delay time.Duration) (string, error) {
	if retry <= 0 {
		return "", errors.New("Retry is 0")
	}
	if form.Get("key") == "" {
		form.Add("key", c.APIKey)
	} else {
		time.Sleep(delay)
	}

	req, err := http.NewRequest("POST", URL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("Failed to initiate TwoCaptcha request, please try again. Error Message: %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error encountered while sending TwoCaptcha request, please try again. Error Message: %s", err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error encountered while reading TwoCaptcha response, please try again. Error Message: %s", err.Error())
	}
	resp.Body.Close()

	stringBody := string(body)

	if strings.Contains(stringBody, "OK|") { //Captcha Solved with answer
		return string(body[3:]), nil
	} else if strings.Contains(stringBody, "CAPCHA_NOT_READY") { //Captcha response not yet ready
		return c.sendRequest(form, URL, retry-1, delay)
	}
	return "", fmt.Errorf("Error response from 2Captcha: %s", stringBody)
}
