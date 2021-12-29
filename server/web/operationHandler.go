package web

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"mokapi/config/dynamic/openapi"
	"mokapi/models"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"net/http"
	"reflect"
	"strings"
)

type OperationHandler struct {
}

func NewOperationHandler() *OperationHandler {
	return &OperationHandler{}
}

func (handler *OperationHandler) ProcessRequest(context *HttpContext) {
	log.WithFields(log.Fields{
		"url":    context.Request.URL.String(),
		"host":   context.Request.Host,
		"method": context.Request.Method,
	}).Info("Processing http request")

	operation := context.Operation

	req := newRequest(context)

	if operation.RequestBody != nil {
		bodyParam, err := r(context)
		if err != nil {
			writeError(err.Error(), http.StatusInternalServerError, context)
			return
		} else if bodyParam == nil {
			writeError("request body expected", http.StatusBadRequest, context)
			return
		}
		req.Body = bodyParam
	}

	res := &Response{
		Headers:    make(map[string]string),
		StatusCode: int(context.statusCode),
	}

	if context.ContentType != nil {
		res.Headers["Content-Type"] = context.ContentType.String()
	}

	gen := openapi.NewGenerator()

	if context.Response != nil {
		if len(context.Response.Examples) > 0 {
			keys := reflect.ValueOf(context.Response.Examples).MapKeys()
			v := keys[rand.Intn(len(keys))].Interface().(*openapi.ExampleRef)
			res.Data = v.Value.Value
		} else if context.Response.Example != nil {
			res.Data = context.Response.Example
		} else {
			res.Data = gen.New(context.Response.Schema)
		}
	}

	for k, v := range context.Headers {
		data := gen.New(v.Value.Schema)
		res.Headers[k] = fmt.Sprintf("%v", data)
	}

	context.metric.EventSummary = context.eventHandler(req, res)

	if err := write(res, context); err != nil {
		writeError(err.Error(), http.StatusInternalServerError, context)
		return
	}
}

func r(ctx *HttpContext) (interface{}, error) {
	contentType := media.ParseContentType(ctx.Request.Header.Get("content-type"))
	body, err := readBody(ctx, contentType)
	if err != nil {
		return nil, err
	}
	if ctx.Operation.RequestBody.Value.Required && body == nil {
		return nil, fmt.Errorf("request body expected")
	}

	return body, nil
}

func readBody(ctx *HttpContext, contentType *media.ContentType) (interface{}, error) {
	if ctx.Request.ContentLength == 0 {
		return "", nil
	}

	media, ok := ctx.Operation.RequestBody.Value.GetMedia(contentType)
	if !ok {
		return nil, fmt.Errorf("content type '%v' of request body is not defined. Check your service configuration", contentType.String())
	}
	if media.Schema == nil || media.Schema.Value == nil {
		return nil, fmt.Errorf("schema of request body %q is not defined", contentType.String())
	}

	schema := media.Schema.Value

	if contentType.Key() == "multipart/form-data" {
		if schema.Type != "object" {
			return nil, fmt.Errorf("schema %q not support for content type multipart/form-data, expected 'object'", schema.Type)
		}
		if schema.Properties.Value == nil {
			// todo raw value
			return nil, nil
		}

		err := ctx.Request.ParseMultipartForm(512) // maxMemory 32MB
		defer func() {
			err := ctx.Request.MultipartForm.RemoveAll()
			if err != nil {
				log.Errorf("error on removing multipart form: %v", err)
			}
		}()
		if err != nil {
			return nil, err
		}

		o := make(map[string]interface{})
		raw := strings.Builder{}

		for name, values := range ctx.Request.MultipartForm.Value {
			raw.WriteString(fmt.Sprintf("%v: %v", name, values))
			p := schema.Properties.Get(name)
			if p == nil || p.Value == nil {
				continue
			}
			if p.Value.Type == "array" {
				a := make([]interface{}, 0, len(values))
				for _, v := range values {
					i, err := parse(v, p.Value.Items)
					if err != nil {
						return nil, err
					}
					a = append(a, i)
				}
				o[name] = a
			} else {
				i, err := parse(values[0], p)
				if err != nil {
					return nil, err
				}
				o[name] = i
			}
		}

		for name, files := range ctx.Request.MultipartForm.File {
			p := schema.Properties.Get(name)
			if p == nil || p.Value == nil {
				continue
			}
			if p.Value.Type == "array" {
				a := make([]interface{}, 0, len(files))
				for _, file := range files {
					i, err := parseFormFile(file)
					if err != nil {
						return nil, err
					}
					a = append(a, i)
				}
				o[name] = a
			} else {
				i, err := parseFormFile(files[0])
				if err != nil {
					return nil, err
				}
				o[name] = i
			}
			//raw.WriteString(fmt.Sprintf("%v: filename=%v, type=%v, size=%v\n", name, fh.Filename, http.DetectContentType(sniff), prettyByteCountIEC(fh.Size)))
		}

		b, _ := json.Marshal(o)
		ctx.metric.Parameters = append(ctx.metric.Parameters, models.RequestParamter{
			Name:  "Body",
			Type:  "Body",
			Value: string(b),
			Raw:   raw.String(),
		})

		return o, nil
	} else {
		data, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, err
		}

		body, err := parseBody(data, contentType, media.Schema)
		if err == nil {
			b, _ := json.Marshal(body)
			ctx.metric.Parameters = append(ctx.metric.Parameters, models.RequestParamter{
				Name:  "Body",
				Type:  "Body",
				Value: string(b),
				Raw:   string(data),
			})
		}
		return body, err
	}
}

func parseFormFile(fh *multipart.FileHeader) (interface{}, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Errorf("unable to close file: %v", err)
		}
	}()

	var sniff [512]byte
	_, err = f.Read(sniff[:])
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"filename": fh.Filename,
		"type":     http.DetectContentType(sniff[:]),
		"size":     prettyByteCountIEC(fh.Size),
	}, nil
}

func parseBody(data []byte, contentType *media.ContentType, schema *openapi.SchemaRef) (interface{}, error) {
	if schema.Value != nil && schema.Value.Type == "string" {
		return string(data), nil
	}

	switch contentType.Subtype {
	case "xml":
		return encoding.ParseXml(string(data), schema)
	case "json":
		return encoding.Parse(data, contentType, schema)
	default:
		log.Debugf("unsupported content type '%v' from body", contentType)
		return string(data), nil
	}
}

func write(r *Response, ctx *HttpContext) error {
	var body []byte
	contentType := media.ParseContentType(r.Headers["Content-Type"])

	if len(r.Body) > 0 {
		body = []byte(r.Body)
	} else if r.Data != nil {
		if bytes, ok := r.Data.([]byte); ok {
			if contentType.Subtype == "*" {
				// detect content type by data
				contentType = media.ParseContentType(http.DetectContentType(bytes))
			}
			body = bytes
		} else {
			if bytes, err := encoding.Encode(r.Data, contentType, ctx.Response.Schema); err != nil {
				return err
			} else {
				body = bytes
			}
		}
	}

	for k, v := range r.Headers {
		ctx.ResponseWriter.Header().Add(k, v)
	}

	if r.StatusCode > 0 {
		ctx.ResponseWriter.WriteHeader(r.StatusCode)
	}

	_, err := ctx.ResponseWriter.Write(body)

	ctx.updateMetric(r.StatusCode, contentType.String(), string(body))

	return err
}

func prettyByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
