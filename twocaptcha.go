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
	req := url.Values{}
	req.Add("method", "base64")
	req.Add("body", base64Str)
	req.Add("soft_id", "2099")

	id, err := c.sendRequest(req, requestURL, 20, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("SolveImageCaptcha Send New Request: [%s]", err.Error())
	}

	resp := url.Values{}
	resp.Add("action", "get")
	resp.Add("id", id)
	result, err := c.sendRequest(resp, responseURL, 20, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("SolveImageCaptcha Send Request to Get Result: [%s]", err.Error())
	}
	return result, nil
}

//SolveRecaptchaV2 solve recaptcha v2
func (c *Client) SolveRecaptchaV2(siteURL, recaptchaKey string) (string, error) {
	req := url.Values{}
	req.Add("googlekey", recaptchaKey)
	req.Add("pageurl", siteURL)
	req.Add("method", "userrecaptcha")
	req.Add("soft_id", "2099")

	id, err := c.sendRequest(req, requestURL, 20, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("SolveRecaptchaV2 Send New Request: [%s]", err.Error())
	}

	resp := url.Values{}
	resp.Add("id", id)
	resp.Add("action", "get")
	result, err := c.sendRequest(resp, responseURL, 20, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("SolveRecaptchaV2 Send Request to Get Result: [%s]", err.Error())
	}
	return result, nil
}

//ReportAnswer report answer
func (c *Client) ReportAnswer(isGood bool, id string) error {
	req := url.Values{}
	action := ""
	if isGood {
		action = "reportgood"
	} else {
		action = "reportbad"
	}
	req.Add("action", action)
	req.Add("id", id)

	resp, err := c.Client.Get(fmt.Sprintf("%s/?key=%s&action=%s&id=%s", responseURL, c.APIKey, action, id))
	if err != nil {
		return fmt.Errorf("ReportAnswer Send New Request: [%s]", err.Error())
	}
	defer resp.Body.Close()
	bodyContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ReportAnswer Read Response: [%s]", err.Error())
	}

	content := string(bodyContent)
	if !strings.Contains(content, "OK_REPORT_RECORDED") {
		return fmt.Errorf("ReportAnswer: Unknown Response [%s] Received", content)
	}
	return nil
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
