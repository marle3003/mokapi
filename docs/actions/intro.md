# Mokapi Actions

**Customize your API responses with Mokapi Actions. You can create actions to perform any response like gathering data from a file or inject faulty behavior**

Mokapi Actions are event-driven, meaning that you can run series of commands after a specified event has occurred. For example, every time an HTTP requests is received, you can add a header to the response.

## Components

Below is a list that describes the components of Mokapi Actions.

### Workflows
A Workflow contains a collection of steps and can be scheduled or triggered by an event.

### Events
An event is a specific activity that triggers a workflow. For example, activity can originate
when someone makes an HTTP request to a specific endpoint.

### Steps
A step is an individual task that can run commands. A step can be either an action or a shell command. Steps can share data with each other

### Actions
Actions are builtin commands which are executed directly in Mokapi instead of an external shell script.

## Create an example workflow
Github Actions uses YAML syntax to define events and steps. Mokapi reads these YAML files from the providers, see [Dynamic Configuration](../config.md).

```yaml
mokapi: 1.0
workflows:
  - name: learn mokapi actions
    on: {service: Sample API, http: {get: /users}}
    steps:
      - uses: set-response
        with:
          statusCode: 500
```

Your workflow will run automatically each time someone makes a request to /users on our "Sample API" service.

## Understanding the workflow
This section explains each line of the introduction's example:

`name: learn mokapi actions` 
Optional - The name of the workflow

`on: {service: Sample API, http: {get: /users}}`
Specify the event that automatically triggers the workflow. This example uses the HTTP event when someone makes a GET-request to the endpoint '/users' on the service 'Sample API'

`steps`
Groups together all the steps that run in this workflow.

`uses: set-response`
The *uses* keyword tells Mokapi to use a builtin action named *set-response*. This is a action that manipulates the response triggered by the request.

`with:`
Groups together all parameter for a builtin command

`statusCode: 500`
Sets the value *500* to the parameter *statusCode*. This will change the HTTP status code of the response to 500