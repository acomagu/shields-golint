package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/xerrors"
)

// BadgeSource represents the response of this application.
type BadgeSource struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
	IsError       bool   `json:"isError"`
}

func genBadgeSource(nErr int) *BadgeSource {
	if nErr == 0 {
		return &BadgeSource{
			SchemaVersion: 1,
			Label:         "golint",
			Message:       "no suggestions",
			Color:         "green",
			IsError:       false,
		}
	}

	var msg string
	if nErr == 1 {
		msg = "1 suggestion"
	} else {
		msg = fmt.Sprintf("%d suggestions", nErr)
	}
	return &BadgeSource{
		SchemaVersion: 1,
		Label:         "golint",
		Message:       msg,
		Color:         "yellow",
		IsError:       true,
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		importPath := req.URL.Path[1:]

		n, err := get(importPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		body, err := json.Marshal(genBadgeSource(n))
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Write(body)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", "8080"), nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func get(importPath string) (int, error) {
	ps := append([]string{"go-lint.appspot.com"}, strings.Split(importPath, "/")...)
	resp, err := http.Get(fmt.Sprintf("https://%s", path.Join(ps...)))
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, xerrors.Errorf("got %s", resp.Status)
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return 0, err
	}

	return goquery.NewDocumentFromNode(root).Find("body>p").Length() - 1, nil
}
