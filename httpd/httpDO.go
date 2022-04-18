package httpd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type requests struct {
	Client  *http.Client
	Cookie  []*http.Cookie
	Headers map[string]interface{}
}

func NewRequest(Client *http.Client, Cookie []*http.Cookie, Headers map[string]interface{}) *requests {
	return &requests{
		Client:  Client,
		Cookie:  Cookie,
		Headers: Headers,
	}
}

func (r *requests) BuildUrl(link string, params map[string]interface{}) string {
	u, _ := url.Parse(link)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v.(string))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (r *requests) HttpGet(link string, params map[string]interface{}) (*http.Response, []byte, error) {

	u, _ := url.Parse(link)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v.(string))
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, errors.New("new request is fail: %v \n")
	}

	if r.Cookie != nil {
		for _, v := range r.Cookie {
			req.AddCookie(v)
		}
	}

	//add headers
	if r.Headers != nil {
		for key, val := range r.Headers {
			req.Header.Add(key, val.(string))
		}
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	return resp, body, nil
}

func (r *requests) HttpPostJson(link string, data map[string]interface{}) (*http.Response, []byte, error) {

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", link, strings.NewReader(string(bytes)))
	if err != nil {
		return nil, nil, errors.New("new request is fail: %v \n")
	}
	// AddCookie
	if r.Cookie != nil {
		for _, v := range r.Cookie {
			req.AddCookie(v)
		}
	}

	//add headers
	if r.Headers != nil {
		for key, val := range r.Headers {
			req.Header.Add(key, val.(string))
		}
	} else {
		req.Header.Add("Content-Type", "application/json; charset=utf-8")
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	return resp, body, nil
}

func (r *requests) HttpPostFrom(link string, body map[string]interface{}) (*http.Response, []byte, error) {
	sendBody := http.Request{}
	sendBody.ParseForm()
	for k, v := range body {
		sendBody.Form.Add(k, v.(string))
	}
	sendData := sendBody.Form.Encode()

	req, err := http.NewRequest("POST", link, strings.NewReader(string(sendData)))
	if err != nil {
		return nil, nil, errors.New("new request is fail: %v \n")
	}
	// AddCookie
	if r.Cookie != nil {
		for _, v := range r.Cookie {
			req.AddCookie(v)
		}
	}

	//add headers
	if r.Headers != nil {
		for key, val := range r.Headers {
			req.Header.Add(key, val.(string))
		}
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	Body, err := ioutil.ReadAll(resp.Body)

	return resp, Body, nil
}
