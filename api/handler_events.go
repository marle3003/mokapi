package api

import (
	"fmt"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type httpFilter = func(d *openapi.HttpLog) bool

type EventInfo struct {
	Id     string        `json:"id"`
	Traits events.Traits `json:"traits"`
	Time   time.Time     `json:"time"`
	Data   any           `json:"data,omitempty"`
}

type EventPage struct {
	Items []EventInfo `json:"items"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int         `json:"total"`
}

type eventQuery struct {
	traits events.Traits
	http   map[string][]string
	page   int
	size   int
}

const maxEventPageSize = 1000

func (h *handler) setupEvents() {
	r := h.router.PathPrefix("/api/events").Subrouter()

	r.HandleFunc("", h.getEvents).Methods(http.MethodGet)
	r.HandleFunc("/{id}", h.getEvent).Methods(http.MethodGet)
}

func (h *handler) getEvents(w http.ResponseWriter, r *http.Request) {
	q, err := parseEventFilter(r.URL.Query())
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	result := EventPage{Page: q.page, Size: q.size}

	data := h.app.Events.GetEvents(q.traits)
	data, err = filterEvents(data, q)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	result.Total = len(data)
	start, end := paginate(q.page, q.size, len(data))

	result.Items = make([]EventInfo, 0, min(q.size, len(data)))
	for _, e := range data[start:end] {
		result.Items = append(result.Items, eventInfo(e))
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	e := h.app.Events.GetEvent(vars["id"])
	if !e.IsValid() {
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, e)
}

func paginate(page int, size int, len int) (int, int) {
	start := min((page-1)*size, len)
	return start, min(start+size, len)
}

func filterEvents(data []events.Event, q eventQuery) ([]events.Event, error) {
	var result []events.Event

	filterHttp, err := matchHttpEvent(q)
	if err != nil {
		return nil, err
	}

	for _, e := range data {
		switch d := e.Data.(type) {
		case *openapi.HttpLog:
			if filterHttp == nil || filterHttp(d) {
				result = append(result, e)
			}
		default:
			result = append(result, e)
		}
	}
	return result, nil
}

func matchHttpEvent(q eventQuery) (httpFilter, error) {
	if q.http == nil {
		return nil, nil
	}

	var filters []func(d *openapi.HttpLog) bool
	for k, v := range q.http {
		switch k {
		case "method":
			filters = append(filters, func(d *openapi.HttpLog) bool {
				match := false
				for _, m := range v {
					if strings.EqualFold(d.Request.Method, m) {
						match = true
					}
				}
				return match
			})
		case "url":
			if v[0] != "" {
				regex, err := regexp.Compile(v[0])
				if err != nil {
					return nil, fmt.Errorf("invalid url regex: %w", err)
				}
				filters = append(filters, func(d *openapi.HttpLog) bool {
					return regex.MatchString(d.Request.Url)
				})
			}
		case "requestHeader":
			return func(d *openapi.HttpLog) bool {
				for _, val := range v {
					kv := strings.SplitN(val, ":", 2)
					match := false
					for _, p := range d.Request.Parameters {
						if p.Type != string(openapi.ParameterHeader) {
							continue
						}
						if p.Name == kv[0] {
							if len(kv) > 1 {
								if p.Raw != nil && kv[1] == *p.Raw {
									match = true
									break
								}
							}
						}
					}
					if !match {
						return false
					}
				}
				return true
			}, nil
		case "responseHeader":
			return func(d *openapi.HttpLog) bool {
				for _, val := range v {
					kv := strings.SplitN(val, ":", 2)
					hv, ok := d.Response.Headers[kv[0]]
					if len(kv) == 1 && !ok {
						return false
					} else if kv[1] != hv {
						return false
					}
				}
				return true
			}, nil
		case "statusCode":
			var values []string
			for _, val := range v {
				values = append(values, strings.Split(val, ",")...)
			}
			var checks []httpFilter
			for i, val := range values {
				if val == "" {
					continue
				}

				not := false
				match := false
				if val[0] == '!' {
					not = true
					val = val[1:]
				}

				if len(val) == 3 && strings.HasSuffix(val, "xx") {
					val = fmt.Sprintf("%c00-%c99", val[0], val[0])
				}
				if strings.Contains(val, "-") {
					parts := strings.Split(val, "-")
					if len(parts) != 2 {
						continue
					}
					minVal, err := strconv.Atoi(parts[0])
					if err != nil {
						return nil, fmt.Errorf("invalid min value in '%s'", values[i])
					}
					maxValue, err := strconv.Atoi(parts[1])
					if err != nil {
						return nil, fmt.Errorf("invalid max value in '%s'", values[i])
					}
					checks = append(checks, func(d *openapi.HttpLog) bool {
						match = d.Response.StatusCode >= minVal && d.Response.StatusCode <= maxValue
						return not != match
					})
				} else {
					expect, err := strconv.Atoi(val)
					if err != nil {
						return nil, fmt.Errorf("invalid value in '%s'", values[i])
					}
					checks = append(checks, func(d *openapi.HttpLog) bool {
						match = d.Response.StatusCode == expect
						return not != match
					})
				}

			}

			return func(d *openapi.HttpLog) bool {
				for _, check := range checks {
					if !check(d) {
						return false
					}
				}
				return true
			}, nil
		case "clientIP":
			var values []string
			for _, val := range v {
				values = append(values, strings.Split(val, ",")...)
			}
			var checks []httpFilter
			for i, val := range values {
				if val == "" {
					continue
				}

				not := false
				match := false
				if val[0] == '!' {
					not = true
					val = val[1:]
				}

				if strings.Contains(val, "/") {
					prefix, err := netip.ParsePrefix(val)
					if err != nil {
						return nil, fmt.Errorf("invalid prefix in '%s'", values[i])
					}
					checks = append(checks, func(d *openapi.HttpLog) bool {
						ip := netip.MustParseAddr(d.ClientIP)
						return not != prefix.Contains(ip)
					})
				} else {
					checks = append(checks, func(d *openapi.HttpLog) bool {
						match = d.ClientIP == val
						return not != match
					})
				}

			}

			return func(d *openapi.HttpLog) bool {
				for _, check := range checks {
					if !check(d) {
						return false
					}
				}
				return true
			}, nil
		}
	}
	if len(filters) == 0 {
		return nil, nil
	}
	return func(d *openapi.HttpLog) bool {
		for _, f := range filters {
			if b := f(d); !b {
				return false
			}
		}
		return true
	}, nil
}

func parseEventFilter(v url.Values) (eventQuery, error) {
	q := eventQuery{
		page:   1,
		size:   25,
		traits: events.NewTraits(),
	}
	var err error

	s := v.Get("page")
	if s != "" {
		q.page, err = strconv.Atoi(s)
		if err != nil || q.page < 1 {
			return q, fmt.Errorf("invalid query parameter 'page'")
		}
	}

	s = v.Get("size")
	if s != "" {
		q.size, err = strconv.Atoi(s)
		if err != nil || q.size < 1 || q.size > maxEventPageSize {
			return q, fmt.Errorf("invalid query parameter 'size'")
		}
	}

	q.http = deepObject(v, "http")

	/*if h := deepObject(q, "http"); len(h["method"]) > 0 {
		f.HTTPMethod = strings.ToUpper(h["method"][0])
	}*/
	for name, values := range deepObject(v, "traits") {
		q.traits[name] = values[0]
	}
	return q, nil
}

func deepObject(q url.Values, name string) map[string][]string {
	prefix := name + "["
	result := make(map[string][]string)

	for key, values := range q {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, "]") {
			continue
		}
		prop := key[len(prefix) : len(key)-1]
		if prop == "" || strings.ContainsAny(prop, "[]") {
			continue
		}
		values = splitQueryValues(values)
		result[prop] = append(result[prop], values...)
	}
	return result
}

func splitQueryValues(values []string) []string {
	var result []string
	for _, v := range values {
		for part := range strings.SplitSeq(v, ",") {
			if part = strings.TrimSpace(part); part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}

func eventInfo(e events.Event) EventInfo {
	return EventInfo{
		Id:     e.Id,
		Traits: e.Traits,
		Time:   e.Time,
		Data:   e.Data.Info(),
	}
}
