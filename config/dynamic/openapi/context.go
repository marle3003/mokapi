package openapi

import (
	"context"
	"fmt"
	"mokapi/models/media"
	"mokapi/server/httperror"
	"net/http"
	"strings"
)

const operationKey = "operation"

func NewOperationContext(ctx context.Context, o *Operation) context.Context {
	return context.WithValue(ctx, operationKey, o)
}

func OperationFromContext(ctx context.Context) (*Operation, bool) {
	o, ok := ctx.Value(operationKey).(*Operation)
	return o, ok
}

func ContentTypeFromRequest(r *http.Request, res *Response) (*media.ContentType, *MediaType, error) {
	accept := r.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, mimeType := range strings.Split(accept, ",") {
			contentType := media.ParseContentType(mimeType)
			if mt := res.GetContent(contentType); mt != nil {
				return contentType, mt, nil
			}
		}
		return nil, nil, httperror.Newf(http.StatusUnsupportedMediaType,
			"none of requests content type(s) are supported: %v", accept)
	}

	for name, mt := range res.Content {
		// return first element
		return media.ParseContentType(name), mt, nil
	}

	return nil, nil, fmt.Errorf("no content type found for accept header %q", accept)
}
