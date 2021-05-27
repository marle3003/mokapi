# Syntax of Mokapi Actions

## name
The name of your workflow.

## on
**Required**. One or more events which triggers this workflow. 

### Example using a single HTTP event
```yaml
on:
  http:
    get: /users
```

### Example using any HTTP GET or POST method
You must append a colon ( *:* ) to all events, including events without a configuration
```yaml
on:
  http:
    get:
    post:
```
The above example in a short form.
```yaml
on:
  http: [get, post]
```

### Example using multiple events
```yaml
on:
  http:
    get: /users
    post:
  service: Sample API
    http:
      get: [/users, /pets]
```

## on.schedule

## env
A *map* of environment variable that are available to all steps in the workflow. You can
also set environment variable that are only available to a single step.

When more than one environment variable is defined with the same name, Mokapi uses the most
specific environment variable. For example, an environment variable defined in a step will
override a workflow variable with the same name, while the step executes.

```yaml
env:
  SERVER: Mokapi
```

## steps
A workflow contains a sequence of tasks called *steps*. Steps can run commands, scripts, or
a builtin action. A command or a script runs in its own process.

### Single line
```yaml
steps:
  - name: Print hello world
    run: echo "Hello World"
```

### Multi line
```yaml
steps:
  - name: Print hello world
    run: |
      ((sum=25+35))
      echo $sum
```

## steps[*].id
A unique identifier for the step. You can use the *id* to reference the step in contexts.
For example to get a step output variable.

## steps[*].uses
Selects a builtin action to run in your step. Some actions require inputs that you must
set using the *with* keyword. 

```yaml
steps:
  - uses: set-response
    with:
      statusCode: 500
```

## steps[*].run
Runs command-line programs using the operating  systems's shell. Each *run* keyword represents
a new process. When you provide multi-line commands, each line runs in the same process.

### Using a specific shell
You can override the default shell using the *shell* keyword. 

#### bash
```yaml
steps:
- name: Display the path
  run: echo $PATH
  shell: bash
```

#### Windows *cmd*
```yaml
steps:
- name: Display the path
  run: echo %PATH%
  shell: cmd
```

#### Powershell
```yaml
steps:
- name: Display the path
  run: echo ${env:PATH}
  shell: cmd
```

## steps[*].with
A *map* of the input parameters defined by the action.

```yaml
steps:
  - uses: set-response
    with:
      statusCode: 500
```

## steps[*].env
Sets environment variable for steps.

