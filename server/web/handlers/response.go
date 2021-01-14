package handlers

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v4"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/models"
	"mokapi/providers/encoding"
	"mokapi/server/web"
	"net/http"
)

type Response struct {
	httpContext *web.HttpContext
}

func (r *Response) AddHeader(key string, value string) {
	r.httpContext.Response.Header().Add(key, value)
}

func (r *Response) WriteString(s string, statusCode int, contentType string) error {
	bytes := []byte(s)

	return r.write(bytes, statusCode, contentType)
}

func (r *Response) Write(object interface{}, statusCode int, contentType string) error {
	bytes, ok := object.([]byte)
	if len(contentType) == 0 {
		contentType = r.httpContext.ContentType.String()
	}

	var err error
	if ok {
		if r.httpContext.ContentType.Subtype == "*" {
			// detect content type by data
			contentType = models.ParseContentType(http.DetectContentType(bytes)).String()
		}
	} else {
		bytes, err = r.encodeData(object)
		if err != nil {
			log.WithFields(log.Fields{"url": r.httpContext.Request.URL.String()}).Errorf(err.Error())
			http.Error(r.httpContext.Response, err.Error(), http.StatusInternalServerError)
			return err
		}
	}

	return r.write(bytes, statusCode, contentType)
}

func (r *Response) WriteRandom(statusCode int, contentType string) error {
	data := getRandomObject(r.httpContext.Schema)
	return r.Write(data, statusCode, contentType)
}

func (r *Response) write(bytes []byte, statusCode int, contentType string) error {
	if statusCode > 0 {
		r.httpContext.Response.WriteHeader(statusCode)
	}
	if len(contentType) == 0 {
		r.AddHeader("Content-Type", r.httpContext.ContentType.String())
	} else {
		r.AddHeader("Content-Type", contentType)
	}

	r.AddHeader("Content-Length", fmt.Sprint(len(bytes)))
	_, err := r.httpContext.Response.Write(bytes)

	return err
}

func (r *Response) encodeData(data interface{}) ([]byte, error) {
	switch r.httpContext.ContentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, r.httpContext.Schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, r.httpContext.Schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", r.httpContext.ContentType)
	}
}

func getRandomObject(schema *models.Schema) interface{} {
	if schema.Type == "object" {
		obj := make(map[string]interface{})
		for name, propSchema := range schema.Properties {
			value := getRandomObject(propSchema)
			obj[name] = value
		}
		return obj
	} else if schema.Type == "array" {
		length := rand.Intn(5)
		obj := make([]interface{}, length)
		for i := range obj {
			obj[i] = getRandomObject(schema.Items)
		}
		return obj
	} else {
		if len(schema.Faker) > 0 {
			switch schema.Faker {
			case "numbers.uint32":
				return gofakeit.Uint32()
			default:
				return gofakeit.Generate(fmt.Sprintf("{%s}", schema.Faker))
			}
		} else if schema.Type == "integer" {
			return gofakeit.Int32()
		} else if schema.Type == "string" {
			return gofakeit.Lexify("???????????????")
		}
	}
	return nil
}
