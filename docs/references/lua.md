---
title: Lua
description: You can extend Mokapi with scripts written in the Lua programming language.
---
# Lua

You can extend Mokapi with scripts written in the Lua programming language. With
extensions points you can map requests to files or generate dynamic responses.

More information on the Lua programming language can be found at the [lua website](http://www.lua.org/).

## Create an HTTP handler
```lua
local log = require "log"
local yaml = require "yaml"
local mokapi = require "mokapi"

mokapi.on("foo", function() {
    local request, response = ...

    if request.url.path == "/models" then
        local data = open("bikes.yml")
        local m = yaml.parse(data)
        log.debug("number of bikes in file: " .. #m.bikes)
        response.data = m.bikes
        return true
    end

    return false
end
```

## Produce Kafka message
```lua
kafka = require "kafka"

k, m, err = kafka.produce("demo:9092", "message", nil, {id=12345, price= 12, shipTo= {name= "Bern"}})
```