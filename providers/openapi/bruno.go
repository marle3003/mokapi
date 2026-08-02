package openapi

import (
	"fmt"
	"maps"
	"mokapi/export/bruno"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	"mokapi/version"
	"net"
	"net/url"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (c *Config) ExportBruno(host string) (bruno.Collection, error) {
	e := bruno.Collection{
		Version: new(version.New("1.0.0")),
		Info: bruno.Info{
			Name:    c.Info.Name,
			Summary: c.Info.Description,
			Version: c.Info.Version,
		},
		Bundled: true,
	}
	if c.Info.Contact != nil {
		e.Info.Authors = append(e.Info.Authors, bruno.Author{
			Name:  c.Info.Contact.Name,
			Email: c.Info.Contact.Email,
			Url:   c.Info.Contact.Url,
		})
	}

	var validServers []*Server
	var resolvedURLs []string

	for _, server := range c.Servers {
		u, err := resolveServerURL(server.Url, host)
		if err != nil {
			log.Debugf("failed to parse url %s: %s", server.Url, err)
			continue
		}
		validServers = append(validServers, server)
		resolvedURLs = append(resolvedURLs, u)
	}

	if len(validServers) > 0 {
		names := environmentNames(validServers, resolvedURLs)
		e.Config = &bruno.Config{}

		for i, server := range validServers {
			resolvedURL := resolvedURLs[i]

			if i == 0 {
				e.Request = &bruno.RequestDefault{
					Variables: []bruno.Variable{{Name: "baseUrl", Value: resolvedURL}},
				}
			}

			e.Config.Environments = append(e.Config.Environments, bruno.Environment{
				Name:        names[i],
				Description: server.Description,
				Variables: []bruno.Variable{
					{Name: "baseUrl", Value: resolvedURL},
				},
			})
		}
	}

	for path, pi := range c.Paths {
		if pi.Value == nil {
			continue
		}
		for method, o := range pi.Value.Operations() {
			name := fmt.Sprintf("%s %s", method, path)
			if o.OperationId != "" {
				name = o.OperationId
			}
			item := bruno.HttpItem{
				Info: &bruno.HttpInfo{
					Name:        name,
					Description: o.Description,
					Type:        "http",
				},
				Http: &bruno.HttpDetail{
					Method:  method,
					Url:     "{{baseUrl}}" + path,
					Headers: nil,
					Params:  nil,
					Body:    nil,
				},
			}

			reqPath := strings.Split(path, "/")
			for i, seg := range reqPath {
				reqPath[i] = strings.Trim(seg, "{}")
			}

			newRandom := func(s *schema.Schema, additionalPath string) string {
				req := &generator.Request{
					Path:   reqPath,
					Schema: schema.ConvertToJsonSchema(s),
				}
				if additionalPath != "" {
					req.Path = append(req.Path, additionalPath)
				}

				v, err := generator.New(req)
				if err != nil {
					log.Debugf("failed to create random data for schema at %s %s: %v", method, path, err)
					return ""
				}
				return fmt.Sprintf("%v", v)
			}

			params := append(pi.Value.Parameters, o.Parameters...)
			addedQuery := false
			for _, ref := range params {
				if ref.Value == nil {
					continue
				}
				p := ref.Value
				switch p.Type {
				case ParameterHeader:
					item.Http.Headers = append(item.Http.Headers, bruno.HttpRequestHeader{
						Name:        p.Name,
						Value:       newRandom(p.Schema, p.Name),
						Description: p.Description,
						Disabled:    !p.Required,
					})
				case ParameterPath:
					// bruno does only encode query parameters but not path parameter
					v := url.PathEscape(newRandom(p.Schema, ""))
					item.Http.Params = append(item.Http.Params, bruno.HttpRequestParam{
						Name:        p.Name,
						Value:       v,
						Description: p.Description,
						Type:        string(ParameterPath),
					})
					item.Http.Url = strings.ReplaceAll(item.Http.Url, fmt.Sprintf("{%s}", p.Name), fmt.Sprintf(":%s", p.Name))
				case ParameterQuery:
					item.Http.Params = append(item.Http.Params, bruno.HttpRequestParam{
						Name:        p.Name,
						Value:       newRandom(p.Schema, p.Name),
						Description: p.Description,
						Type:        string(ParameterQuery),
						Disabled:    !p.Required,
					})
					if !addedQuery {
						item.Http.Url += "?"
						addedQuery = true
					}
					item.Http.Url += fmt.Sprintf("%s=", p.Name)
				default:
					log.Debugf("unsupported type %s for parameter %s", p.Type, p.Name)
				}
			}

			if o.RequestBody != nil && o.RequestBody.Value != nil {
				rb := o.RequestBody.Value
				var result []bruno.HttpRequestBodyVariant

				keys := slices.Collect(maps.Keys(rb.Content))
				slices.SortFunc(keys, func(a, b string) int {
					return strings.Compare(a, b)
				})

				for _, key := range keys {
					mt := rb.Content[key]
					typeName := "text"
					ct := media.ParseContentType(key)
					if ct.Subtype == "json" {
						typeName = "json"
					}
					if ct.Subtype == "xml" {
						typeName = "xml"
					}

					var b []byte
					s := schema.ConvertToJsonSchema(mt.Schema)
					v, err := generator.New(&generator.Request{Path: reqPath, Schema: s})
					if err != nil {
						log.Debugf("failed to create random data for body at %s %s: %v", method, reqPath, err)
					} else {
						b, err = encoding.NewEncoder(s).Write(v, ct)
					}

					result = append(result, bruno.HttpRequestBodyVariant{
						Title: key,
						Body: bruno.HttpRequestBodyRaw{
							Type: typeName,
							Data: string(b),
						},
					})
					if len(result) == 1 && rb.Required {
						result[0].Selected = true
					}
				}

				if len(result) > 1 {
					item.Http.Body = &bruno.HttpRequestBody{
						Variant: result,
					}
				} else if len(result) == 1 {
					item.Http.Body = &bruno.HttpRequestBody{
						Body: &result[0].Body,
					}
				}
			}

			e.Items = append(e.Items, item)
		}
	}

	return e, nil
}

func resolveServerURL(rawURL, host string) (string, error) {
	h, _, _ := net.SplitHostPort(host)

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if u.Hostname() == "" {
		if port := u.Port(); port != "" {
			u.Host = net.JoinHostPort(h, port)
		} else {
			u.Host = host
		}
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	s := u.String()
	s = strings.TrimSuffix(s, "/")

	return s, nil
}

func environmentNames(servers []*Server, resolvedURLs []string) []string {
	names := make([]string, len(servers))

	// Prefer a description, if available
	for i, s := range servers {
		if name := format(s.Description, 45); name != "" {
			names[i] = name
		}
	}

	// remove common prefix for the rest
	prefix := longestCommonPrefix(resolvedURLs)
	if prefix != "" {
		for i, name := range names {
			if name != "" {
				continue
			}
			trimmed := strings.TrimPrefix(resolvedURLs[i], prefix)
			names[i] = format(trimmed, 45)
		}
	} else {
		// try to remove scheme
		urls := tryRemoveScheme(resolvedURLs, "http")
		if urls == nil {
			urls = tryRemoveScheme(resolvedURLs, "https")
		}
		if urls != nil {
			resolvedURLs = urls
		}
		for i, name := range names {
			if name != "" {
				continue
			}
			trimmed := strings.TrimPrefix(resolvedURLs[i], prefix)
			names[i] = format(trimmed, 45)
		}
	}

	return uniqueNames(names)
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}

		if strings.TrimPrefix(s, prefix) == "" {
			return ""
		}

		if !strings.HasSuffix(prefix, "://") {
			prefix = strings.TrimSuffix(prefix, "/")
		}
	}

	if strings.TrimPrefix(strs[0], prefix) == "" {
		return ""
	}

	return prefix
}

func format(s string, maxLen int) string {
	if s == "" {
		return ""
	}

	if idx := strings.IndexAny(s, "\n\r"); idx != -1 {
		s = s[:idx]
	}

	s = strings.ToLower(s)
	var b strings.Builder
	skipNextDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.':
			b.WriteRune(r)
			skipNextDash = false
		default:
			if !skipNextDash {
				b.WriteRune('-')
				skipNextDash = true
			}
		}
	}
	s = strings.Trim(b.String(), "-")

	if len(s) > maxLen {
		s = s[:maxLen]
		if idx := strings.LastIndex(s, "-"); idx > maxLen/2 {
			s = s[:idx]
		}
		s = strings.Trim(s, "-")
	}
	return s
}

func uniqueNames(names []string) []string {
	seen := map[string]int{}
	result := make([]string, len(names))
	for i, n := range names {
		if n == "" {
			n = "env"
		}
		seen[n]++
		if seen[n] == 1 {
			result[i] = n
		} else {
			result[i] = fmt.Sprintf("%s-%d", strings.TrimSuffix(n, "-"), seen[n])
		}
	}
	return result
}

func tryRemoveScheme(urls []string, scheme string) []string {
	results := make([]string, len(urls))
	prefix := scheme + "://"
	for i, u := range urls {
		if !strings.HasPrefix(u, prefix) {
			return nil
		}
		results[i] = strings.TrimPrefix(u, prefix)
	}
	return results
}
