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
	"path"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (c *Config) ExportBruno(baseUrl string) (bruno.Collection, error) {
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
		u, err := resolveURL(server.Url, baseUrl, "http")
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

	if c.Paths != nil && c.Paths.Len() > 0 {
		items := groupByTag(c.Paths, tagsByName(c.Tags))
		if len(items) > 0 {
			e.Items = items
		}
	}

	return e, nil
}

func resolveURL(rawURL, baseUrl, defaultScheme string) (string, error) {
	if baseUrl == "" {
		return rawURL, nil
	}

	base, err := parseBaseUrl(baseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid baseUrl: %w", err)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if u.Hostname() == "" {
		if port := u.Port(); port != "" {
			u.Host = net.JoinHostPort(base.Hostname(), port)
		} else {
			u.Host = base.Host
		}
		u.Path = path.Join(base.Path, u.Path)
	}

	if u.Scheme == "" {
		u.Scheme = defaultScheme
	}

	s := strings.TrimSuffix(u.String(), "/")

	return s, nil
}

func parseBaseUrl(baseUrl string) (*url.URL, error) {
	if !strings.Contains(baseUrl, "://") {
		baseUrl = "http://" + baseUrl
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid base url %q: %w", baseUrl, err)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("invalid base url %q: no hostname", baseUrl)
	}
	return u, nil
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

func groupByTag(paths *PathItems, tagsByName map[string]*Tag) []any {
	root := &folderBuilder{items: map[string]*folderBuilder{}}
	var untagged []*Operation

	for it := paths.Iter(); it.Next(); {
		ref := it.Value()
		pathItem := ref.Value
		if pathItem == nil {
			continue
		}

		lookup := pathItem.Operations()
		for _, method := range pathItem.MethodOrder {
			op := lookup[method]
			if len(op.Tags) == 0 {
				untagged = append(untagged, op)
				continue
			}
			path := tagPath(tagsByName, op.Tags[0])
			if len(path) == 0 {
				// tag referenced but not declared in spec's top-level tags array
				untagged = append(untagged, op)
				continue
			}
			root.insert(path, op)
		}
	}

	items := root.build()
	items = append(items, buildItems(untagged, len(items)+1)...)

	return items
}

func buildItems(operations []*Operation, seq int) []any {
	var items []any

	for _, o := range operations {
		pi := o.Path
		path := pi.Path
		method := o.Method

		name := fmt.Sprintf("%s %s", method, path)
		if o.OperationId != "" {
			name = o.OperationId
		}
		item := bruno.HttpItem{
			Info: &bruno.HttpInfo{
				Name:        name,
				Description: o.Description,
				Type:        "http",
				Sequence:    seq,
			},
			Http: &bruno.HttpDetail{
				Method:  method,
				Url:     "{{baseUrl}}" + path,
				Headers: nil,
				Params:  nil,
				Body:    nil,
			},
		}
		if item.Info.Description == "" {
			item.Info.Description = o.Summary
		}

		reqPath := strings.Split(path, "/")
		for i, seg := range reqPath {
			reqPath[i] = strings.Trim(seg, "{}")
		}

		params := append(pi.Parameters, o.Parameters...)
		buildParams(item.Http, params, reqPath)

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
				if len(result) == 1 && rb.IsRequired() {
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

		items = append(items, item)
		seq++
	}
	return items
}

type folderBuilder struct {
	ops   []*Operation
	order []*Tag
	items map[string]*folderBuilder
}

func (f *folderBuilder) insert(path []*Tag, op *Operation) {
	if len(path) == 1 {
		child := f.child(path[0])
		child.ops = append(child.ops, op)
		return
	}
	f.child(path[0]).insert(path[1:], op)
}

func (f *folderBuilder) child(tag *Tag) *folderBuilder {
	if _, ok := f.items[tag.Name]; !ok {
		f.items[tag.Name] = &folderBuilder{items: map[string]*folderBuilder{}}
		f.order = append(f.order, tag)
	}
	return f.items[tag.Name]
}

func (f *folderBuilder) build() []any {
	var result []any
	seq := 1

	for _, tag := range f.order {
		child := f.items[tag.Name]
		item := bruno.FolderItem{
			Info: &bruno.FolderInfo{
				Name:        tag.Name,
				Description: tag.Description,
				Type:        "folder",
				Sequence:    seq,
			},
			Items: child.build(),
		}
		if item.Info.Description == "" {
			item.Info.Description = tag.Summary
		}
		result = append(result, item)
		seq++
	}

	result = append(result, buildItems(f.ops, seq)...)

	return result
}
