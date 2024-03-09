---
title: HTTP Quick Start - Mock an HTTP API that don't exists yet
description: A quick tutorial how to run Swagger's Petstore in Mokapi
---
# Quick Start
In this quick start we will run [Swagger](https://swagger.io/)'s Petstore in Mokapi

## Start Mokapi
Start Mokapi with the following command

```
docker run -it --rm -p 80:80 -p 8080:8080 -e MOKAPI_Providers_Http_Url=https://raw.githubusercontent.com/marle3003/mokapi/v0.9/examples/swagger/petstore-3.0.json mokapi/mokapi:latest
```
When you use Mokapi in a company, you probably will need to skip SSL verification: `-e MOKAPI_Providers_Http_TlsSkipVerify=true`.

You can open a browser and go to Mokapi's Dashabord (`http://localhost:8080`) to browse the Petstore's API.

## Make HTTP request
You can now make HTTP requests to the Petstore's API and Mokapi creates responses with randomly generated data.
Your request and the response is shown on Mokapi's Dashboard.

```
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```

A possible response could be as shown below.

```json
{
  "id": 10,
  "name": "Waldo",
  "category": {
    "id": 1,
    "name": "Dogs"
  },
  "photoUrls": [
    "http://www.chiefcontent.name/cross-media/out-of-the-box/metrics"
  ],
  "tags": [
    {
      "id": 7100881306113605000,
      "name": " WPMazgdHUX"
    },
    {
      "id": 5550604200940089000,
      "name": "liIe4zU0aX0WUp"
    },
    {
      "id": 8592233813571558000,
      "name": "JFRzXlfBOL"
    },
    {
      "id": 5949990397095072000,
      "name": "ZJtLxPTxNsfvLo"
    }
  ],
  "status": "pending"
}
```

