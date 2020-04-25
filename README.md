# TwoCaptcha

Two Captcha Client

```
go get github.com/Z-M-Huang/Go2Captcha
```

```
import twocaptcha "github.com/Z-M-Huang/Go2Captcha"

captchaClient := &twocaptcha.Client{
  APIKey: "Your 2Captcha API Key",
  Client: &http.Client{},
}

response, err := captchaClient.SolveRecaptchaV2("url", "siteKey")
if err != nil {
  log.Fatal(err)
}
```
