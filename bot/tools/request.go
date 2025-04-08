package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Request http.Request

var ErrorEmptyPath = fmt.Errorf("path cannot be empty")

func Get(path string) *Request {
	req, _ := http.NewRequest("GET", path, nil)

	return (*Request)(req)
}

func Post(path string, body io.Reader) *Request {
	req, _ := http.NewRequest("POST", path, body)

	return (*Request)(req)
}

func (req *Request) AddHeader(key, value string) *Request {
	(*http.Request)(req).Header.Set(key, value)

	return req
}

func (req *Request) AddCookie(key, value string) *Request {
	(*http.Request)(req).AddCookie(&http.Cookie{Name: key, Value: value})

	return req
}

func (req *Request) AddQuery(key, value string) *Request {
	query := (*http.Request)(req).URL.Query()
	query.Add(key, value)
	(*http.Request)(req).URL.RawQuery = query.Encode()

	return req
}

func (req *Request) Do() (io.ReadCloser, error) {
	if req.URL.String() == "" {
		return nil, ErrorEmptyPath
	}

	resp, err := http.DefaultClient.Do((*http.Request)(req))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request!\n%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func ToString(reader io.ReadCloser) (string, error) {
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse into string!\n%w", err)
	}

	return string(body), nil
}

func ToJson(reader io.ReadCloser) (*Json, error) {
	defer reader.Close()

	var jsonData Json

	err := json.NewDecoder(reader).Decode(&jsonData.Data)
	if err != nil {
		return &Json{}, fmt.Errorf("failed to parse into JSON!\n%w", err)
	}

	return &jsonData, nil
}

func ToDocument(reader io.ReadCloser) (*goquery.Document, error) {
	defer reader.Close()

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return &goquery.Document{}, fmt.Errorf("failed to parse into document!\n%w", err)
	}

	return doc, nil
}

func GetTitle(path string) (string, error) {
	reader, err := Get(path).AddHeader("User-Agent", UserAgent).Do()
	if err != nil {
		return path, err
	}

	doc, err := ToDocument(reader)
	if err != nil {
		return "", err
	}

	title := doc.Find("title").Text()

	return title, nil
}
