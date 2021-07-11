# Actions
Actions are builtin commands which are executed directly in Mokapi instead of an external shell script.

- [read-file](./#read-file-action)
- [mustache](#mustache-action)
- [parse-yaml](#parse-yaml-action)
- [split](#split-action)
- [xpath](#xpath-action)

## read-file action
Read file contents

### Usage
```yaml
steps:
  - uses: read-file
    id: data
    with:
      path: ./data.yml
  - run: echo "${{ steps.data.outputs.content }}"
```

## mustache action
Replace tags in a template. See the [manpage](http://mustache.github.io/mustache.5.html)

### Usage
```yaml
steps:
  - uses: mustache
    id: demo
    with:
      template: 'Say: {{ Label }}'
      data: "${{ [Label: 'Hello World'] }}"
  - run: echo "${{ steps.demo.outputs.result }}"
```

## parse-yaml action
Returns a YAML object for the given input.

### Usage
```yaml
steps:
  - uses: read-file
    id: data
    with:
      path: ./data.yml
  - uses: parse-yaml
    id: demo
    with:
      content: ${{ steps.data.outputs.content }}
  - run: echo "${{ steps.demo.outputs.result }}"
```

## split action
Split a string

### Inputs
- *s*: String to split
- *separator*: The delimiter to split the string.
- *n*: Maximum number of splits. Default: *-1* (no limit)

### Outputs
- *_0*, *_1*,...,*_n*: Each result of a split

### Usage
```yaml
steps:
  - uses: split
    id: demo
    with:
      s: "Hello World"
      separator: " "
  - run: echo "${{ steps.demo.outputs._0 + " " +  steps.demo.outputs._1 }}"
```

## xpath action
Return the string value of the first node matched by the expression

### Usage
```yaml
steps:
  - uses: xpath
    id: demo
    with:
      content: '<book category="cooking"><title lang="en">Everyday Italian</title><author>Giada De Laurentiis</author><year>2005</year><price>30.00</price></book>'
      expression: /bookstore/book/title
  - run: echo "${{ steps.demo.outputs.result }}"
```

## delay action
Pauses the current Mokapi Action for at least the duration.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", default is "s"

### Fixed length of time

```yaml
steps:
  - uses: delay
    with:
      time: 10 # 10 seconds
  - uses: delay
    with:
      time: 10
      unit: s
  - uses: delay
      with:
        time: 10m # 10 minutes
  - uses: delay
      with:
        time: 1m30s # 1min 30 seconds
```

### Variable length of time

`sigma` standard deviation of logarithmic values<br />
`mean` mean of logarithmic values<br />
`lower` lower bound of the range, inclusive<br />
`upper` upper bound of the range, inclusive

```yaml
steps:
  - uses: delay
    with:
      type: lognormal
      sigma: 5
      mean: 20
      unit: ms
  - uses: delay
    with:
      type: uniform
      lower: 10
      upper: 30
```

## set-response
Manipulates the HTTP response of the current request.

`body` response body as plain text, overrides data.<br />
`data` any object marshalling to schema and encoding to content-type.<br />
`statusCode` HTTP status code of the response.<br />
`headers` add additionally or manipulate headers of the response.<br />
`contentType` sets the content type of the response body.<br />

### Usage
```yaml
steps:
  - uses: set-response
    with:
      body: Hello World
      statusCode: 200
      headers: "${{ {'access-control-allow-origin': '*'}}"
```

## send-mail
Sends the specified message to an SMTP server for delivery

`server` a string that contains the name or IP address with optional port. Default port is 25. <br />
`from` a string that contains the from address. <br />
`to` a string that contains the to address. <br />
`contentType` HTTP status code of the response. <br />
`encoding` the encoding used to encode body. <br />
`subject` a string that contains the subject for this mail.<br />
`body` a string that contains the body for this mail. <br />

### Usage
```yaml
steps:
  - uses: send-mail
    with:
      server: localhost
      from: demo@mokapi.io
      to: test@mokapi.io
      body: Hello World
      contentType: text/plain
```

## kafka-producer
With Mokapi Actions you can produce scheduled messages to your Kafka channels.

`broker` broker server address. <br />
`topic` name of the topic. <br />
`broker` broker server address. <br />
`key` key of the message, if null Mokapi generates a random one. <br />
`message` the message, if null Mokapi generates a random one. <br />
`partition` the partition to write the message to. Default is -1, indicating to choose a random one. <br />

```yaml
steps:
  - uses: kafka-producer
    with:
      broker: localhost:9092
      topic: message
```