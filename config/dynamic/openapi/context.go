package openapi

import (
	"context"
	"mokapi/media"
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

func ContentTypeFromRequest(r *http.Request, res *Response) (media.ContentType, *MediaType, error) {
	accept := r.Header.Get("accept")
	ct, mt := negotiateContentType(accept, res)
	if ct.IsEmpty() {
		return media.Empty, nil, httperror.Newf(http.StatusUnsupportedMediaType,
			"none of requests content type(s) are supported: %q", accept)
	} else if ct.IsRange() {
		return media.GetRandom(accept), mt, nil
	}

	return ct, mt, nil
}

func negotiateContentType(accept string, res *Response) (media.ContentType, *MediaType) {
	if accept == "" || accept == "*" {
		accept = "*/*"
	}

	best := media.Empty
	bestSpec := media.Empty
	var bestMediaType *MediaType
	bestQ := -1.0
	for _, spec := range parseAccept(accept) {
		for _, mt := range res.Content {
			if spec.Match(mt.ContentType) {
				if bestQ > spec.Q {
					continue
				}
				if !best.IsEmpty() && !best.IsRange() {
					continue
				}
				if !best.IsEmpty() && len(best.Parameters) > len(mt.ContentType.Parameters) {
					continue
				}
				best = mt.ContentType
				bestQ = spec.Q
				bestSpec = spec
				bestMediaType = mt
			}
		}
	}

	if best.String() != media.Empty.String() && best.IsRange() {
		return bestSpec, bestMediaType
	}

	return best, bestMediaType
}

func parseAccept(s string) []media.ContentType {
	var ret []media.ContentType
	for _, v := range strings.Split(s, ",") {
		ret = append(ret, media.ParseContentType(v))
	}
	return ret
}
