local mokapi = require "mokapi"

local store = {}
store["1"] = {
name = "Gidget",
photoUrls = {"http://www.pets.com/gidget.png"},
status = "pending",
}
store["2"] = {
name = "Max",
photoUrls = {"http://www.pets.com/max.png"},
status = "available",
}
store["3"] = {
name = "Rooster",
photoUrls = {"http://www.pets.com/rooster.png"},
status = "sold",
}

local function getPetsByStatus (status)
    local result = {}
    for _, v in pairs (store) do
        local match = false
        for i = 1, #status do
            if v.status == status[i] then
                match = true
            end
        end
        if match then
            table.insert(result, v)
        end
    end
    return result
end

mokapi.on("http", function(request, response)
    if request.operationId == "findPetsByStatus" then
        response.data = getPetsByStatus(request.query.status)
        return true
     end
     return false
end
)