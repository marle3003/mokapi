# Mokapi Pipelines

A pipeline is one or more stages that describe a process. A process could be a gathering data from a file for a REST request. Stages can be used for different behaviors.

The structure of a pipeline in a YAML file is like:

- pipeline P
   - stage A
     condition Q
     step 1.1
     step 1.2
   - stage B
     ...

Simple pipelines don't require all of these levels.

The following sections covers the schema of a Mokapi Pipeline. To learn the basics of YAML, see https://learnxinyminutes.com/docs/yaml.

## Pipeline

```yaml
pipelines:
  - name: string # name of the pipeline
    stages: [ stage ]
```

if you have a single stage, you can omit the stages keyword and directly specify the steps keyword:

```yaml
pipelines:
  - name: string # name of the pipeline
    steps: string # the steps of the pipeline
```

## Stage

A stage is a collection of steps. By default, stages run sequentially. Use conditions to control when a stage should run.
Conditions are written as expressions. The result of these expression is a boolean value that determines if the steps of a stage
should run or not.

```yaml
stages:
  - name: string # name of the stage
	condition: string
    steps: string # the steps of the pipeline
```

## Steps

Steps tell Mokapi what to do and serve as the basic building block for declartive pipeline syntax.

```yaml
steps: echo 'Hello World!'
```

## Variables

When you define a variable, you can use different syntaxes and what syntax you use will determine how the variable will be processed. In addition to user-defined variables, Mokapi has system variables with predefined values. Some variables are only in specific context available. Environment variables are also available and depend on your operating system you are using.

```yaml
pipelines:
  - name: search
    variables:
	  - name: variable1
	    value: 10
	  - variable2: myValue
	  - variable3 := 10 + 5
```