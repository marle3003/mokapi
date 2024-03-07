---
title: How to start Running Mokapi - A Beginners Guide
description: Here you will learn how to run Mokapi
---
# Running Mokapi

In this quick start we will run [Swagger](https://swagger.io/)'s Petstore in Mokapi

## Start Mokapi
Start Mokapi with the following command

```bash tab=Docker  
docker run -it --rm -p 80:80 -p 8080:8080 -e MOKAPI_Providers_Http_Url=https://petstore3.swagger.io/api/v3/openapi.json mokapi/mokapi:latest
```

```bash tab=NPM
npm install go-mokapi
npx mokapi --Providers.Http.Url=https://petstore3.swagger.io/api/v3/openapi.json
```
``` box=tip
When you use Mokapi behind a corporate proxy, you probably need 
to skip SSL verification: "-e MOKAPI_Providers_Http_TlsSkipVerify=true".
```

## Mokapi's Dashboard

You can now open a browser and go to Mokapi's Dashabord 
(`http://localhost:8080`) to browse the Petstore's API.

## Make HTTP request
Now make a HTTP request to the Petstore's API and Mokapi 
creates a response with randomly generated data.

```
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```
In Mokapi's Dashboard you can see your request and the response.