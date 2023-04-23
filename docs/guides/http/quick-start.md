# Quick Start
In this quick start we will run [Swagger](https://swagger.io/)'s Petstore in Mokapi

## Start Mokapi
Start Mokapi with the following command

```
docker run -it --rm -p 80:80 -p 8080:8080 -e MOKAPI_Providers_Http_Url=https://petstore3.swagger.io/api/v3/openapi.json mokapi/mokapi:v0.5.0-alpha
```
When you use Mokapi in a company, you probably will need to skip SSL verification: `-e MOKAPI_Providers_Http_TlsSkipVerify=true`.

You can open a browser and go to Mokapi's Dashabord (`http://localhost:8080`) to browse the Petstore's API.

## Make HTTP request
You can now make HTTP requests to the Petstore's API and Mokapi creates responses with randomly generated data.

```
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```
You can now see your request and the response on Mokapi's Dashboard.