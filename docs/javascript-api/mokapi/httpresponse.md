# HttpResponse

HttpResponse is an object used by the [HttpEventHandler](/docs/javascript-api/mokapi/eventhandler.md)
that contains response-specific data such as HTTP headers.

| Name       | Type   | Description                                                              |
|------------|--------|--------------------------------------------------------------------------|
| statusCode | number | Specifies the http status used to select the OpenAPI response definition |
| headers    | object | Object contains header parameters specified by OpenAPI header parameters |
| body       | string | Response body. It has a higher precedence than data                      |
| data       | any    | Data will be encoded with the OpenAPI response definition                |


