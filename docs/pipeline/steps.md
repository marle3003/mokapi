# Pipeline Steps Reference

## Delay Step
Delay execution of the pipeline

### Fixed length of time
```
delay time: 10 # 10 seconds
delay time: 10, unit: 'm' # 10 minutes
delay time: '10m' # 10 minutes
delay time: '1m30s' # 1min 30 seconds
```
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"

### Variable length of time
```
delay type: 'lognormal', sigma: 5, mean: 20
delay type: 'lognormal', sigma: 5, mean: 20, unit: 'ms' 
delay type: 'uniform', lower: 10, upper: 30
```
`sigma` standard deviation of logarithmic values<br />
`mean` mean of logarithmic values<br />
`lower` lower bound of the range, inclusive<br />
`upper` upper bound of the range, inclusive

## Echo Step
Prints a message to the log

```
echo 'Hello World'
```

## FileExists Step
Checks if the given file exists

```
fileExists "data/${params.id}.jpg"
fileExists file: "data/${params.id}.jpg"
```

## ReadFile Step
Reads a file and returns its content

```
readFile './data/file1.xml' # reads content as a plain string
readFile './data/file2.yml', AsYml: true # reads content as yml
readFile './data/file3.json', AsJson: true # reads content as json
```

## Mustache Step
Replaces tags in a template. See the [manpage](http://mustache.github.io/mustache.5.html)

```
s := mustache 'Say: {{ Label }}', [Label: 'Hello World'] # Say: Hello World
s := mustache template: 'Say: {{ Label }}', data: [Label: 'Hello World'] # Say: Hello World
```
Using double curly brackets inside mokapi file must be *escaped*: ```'Say: {{ `{{Label}}` }}'```