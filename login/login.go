package login

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
)

var ErrIncorrectLoginInfo = errors.New("Username/Password is Incorrect")

func extractCSRFToken(page string) string {
	rgx := regexp.MustCompile(`<input type="hidden" name="execution" value="(.*?)"`)
	return rgx.FindStringSubmatch(page)[1]
}

func Login(url, username, password string) (client *http.Client, source string, err error) {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}

	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	resp.Body.Close()

	csrf := extractCSRFToken(string(content))

	resp, err = client.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte("username="+username+"&password="+password+"&execution="+csrf+"&_eventId=submit&geolocation=")))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if strings.Contains(string(buf), "نام کاربری یا کلمه عبور اشتباه است") {
		return nil, "", ErrIncorrectLoginInfo
	}

	return client, string(buf), nil
}
