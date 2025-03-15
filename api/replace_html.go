package api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"net/url"
	"regexp"
	"strings"
)

var (
	regexTitle         = regexp.MustCompile(`(<title>)(.*)(</title>)`)
	regexDescription   = regexp.MustCompile(`(<meta name="description" content=")(.*)(" />)`)
	regexOgTitle       = regexp.MustCompile(`(<meta property="og:title" content=")(.*)(">)`)
	regexOgDescription = regexp.MustCompile(`(<meta property="og:description" content=")(.*)(">)`)
)

func (h *handler) replaceMeta(u *url.URL, html string) string {
	title, description, err := h.getMetaInfo(u)
	if err != nil {
		log.Errorf("set http meta info failed for request %v: %v", u.String(), err)
		return html
	}
	if len(title) == 0 {
		return html
	}

	html = regexTitle.ReplaceAllString(html, fmt.Sprintf("${1}%s | mokapi.io${3}", title))
	html = regexDescription.ReplaceAllString(html, fmt.Sprintf("${1}%s${3}", description))
	html = regexOgTitle.ReplaceAllString(html, fmt.Sprintf("${1}%s | mokapi.io${3}", title))
	html = regexOgDescription.ReplaceAllString(html, fmt.Sprintf("${1}%s${3}", description))

	return html
}

func (h *handler) getMetaInfo(u *url.URL) (title string, description string, err error) {
	segments := strings.Split(u.EscapedPath(), "/")
	if len(segments) <= 3 {
		return
	}
	if segments[1] != "dashboard" || segments[3] != "services" {
		return
	}

	switch segments[2] {
	case "http":
		return getHttpMetaInfo(segments, h.app)
	}

	return
}

func getHttpMetaInfo(segments []string, app *runtime.App) (title, description string, err error) {
	name, err := url.PathUnescape(segments[4])
	if err != nil {
		err = fmt.Errorf("unescape path '%v' failed: %w", segments[4], err)
		return
	}

	c := app.Http.Get(name)
	if c == nil {
		return
	}
	title = c.Info.Name
	description = c.Info.Description

	if len(segments) == 5 {
		return
	}

	pathName, n, path, err := getPath(segments, c)

	title = fmt.Sprintf("%v - %v", pathName, c.Info.Name)

	if path == nil || path.Value == nil {
		description = "Path not found"
		return
	}

	if len(path.Value.Summary) > 0 {
		description = path.Value.Summary
	} else if len(path.Value.Description) > 0 {
		description = path.Value.Description
	}

	if len(segments) == n+1 {
		return
	}

	operationName := segments[n+1]
	op := path.Value.Operation(operationName)

	title = fmt.Sprintf("%v %v - %v", strings.ToUpper(operationName), pathName, c.Info.Name)

	if op == nil {
		description = "Operation not found"
	}

	if len(op.Summary) > 0 {
		description = op.Summary
	} else if len(op.Description) > 0 {
		description = op.Description
	}

	return
}

func getPath(segments []string, c *runtime.HttpInfo) (pathName string, n int, path *openapi.PathRef, err error) {
	currentPath := ""
	for i := 5; i < len(segments); i++ {
		var seg string
		seg, err = url.PathUnescape(segments[i])
		if err != nil {
			err = fmt.Errorf("unescape path '%v' failed: %w", segments[i], err)
			return
		}
		currentPath = fmt.Sprintf("%s/%s", currentPath, seg)
		p, found := c.Paths[currentPath]
		if found {
			pathName = currentPath
			path = p
			n = i
		}
	}
	return
}
