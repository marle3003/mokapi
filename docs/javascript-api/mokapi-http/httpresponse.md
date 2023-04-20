# HttpResponse

HttpResponse object contains response data from a HTTP request 
from functions in the module mokapi/http.

| Name       | Type     | Description                                                           |
|------------|----------|-----------------------------------------------------------------------|
| body       | string   | Response body content                                                 |
| headers    | object   | Request headers as \[string\]string[]                                 |
| json()     | function | Parses the response body as JSON. Returns a JS object, array or value |
| statusCode | number   | HTTP status code returned by the server                               |
