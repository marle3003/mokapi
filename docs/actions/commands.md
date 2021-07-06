# Commands for Mokapi Actions

Actions can communicate with Mokapi to set output values used by other actions. 

## Setting an output parameter

`::set-output name={name}::{value}`

Sets an action's output parameter

### Example

Using shell
```shell
echo "::set-output name=msg::hello world"
```

Using Powershell
```powershell
Write-Host "::set-output name=msg::hello world"
```