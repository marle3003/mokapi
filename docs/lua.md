# Lua

You can extend Mokapi with scripts written in the Lua programming language. With
extensions points you can map requests to files or generate dynamic responses.

More information on the Lua programming language can be found at the [lua website](http://www.lua.org/).

## Create an HTTP handler
```lua
log = require("log")
yaml = require("yaml")

local function eventHandler(self, event, ...)
    local request, response = ...

    if request.url.path == "/models" then
        local m = yaml.read_file(script_dir .. "/bikes.yml")
        bikes = m.bikes
        log.debug("number of bikes in file: " .. #bikes)
        
        return true
    end

    return false
end

demo = workflow.new("demo")
demo:event("http", eventHandler)
```

## Produce Kafka message
```lua
kafka = require("kafka")

repeat
    --repeat until broker is running
   k, m, err = kafka.produce("demo:9092", "message", nil, {id=12345, price= 12, shipTo= {name= "Bern"}})
until err == nil
```