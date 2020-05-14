package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

func NewService() Service {
	cookieJar, _ := cookiejar.New(nil)

	return Service{client: &http.Client{
		Jar: cookieJar,
	}}
}

type Service struct {
	client *http.Client
}

func (s Service) Login(username string, password string) error {
	payload := url.Values{}
	payload.Set("vfw_form", "login_check_password")
	payload.Set("vfw_coll", "direct")
	payload.Set("username", username)
	payload.Set("password", password)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.dijnet.hu/ekonto/login/login_check_password",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !strings.Contains(string(body), "Bejelentkez&eacute;si n&eacute;v: <strong>"+username+"</strong>") {
		return fmt.Errorf("wrong username or password")
	}
	return nil
}

func main() {
	username := os.Getenv("DIJNET_USERNAME")
	password := os.Getenv("DIJNET_PASSWORD")
	srv := NewService()
	err := srv.Login(username, password)
	fmt.Println(err)
}
