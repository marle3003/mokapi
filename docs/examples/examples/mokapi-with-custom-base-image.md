---
title: Mokapi with custom base image
description: How to use Mokapi with your custom base Docker image
---

# Use Mokapi with a custom base Docker image 


```Dockerfile tab=Dockerfile  
FROM mokapi/mokapi as mokapi

FROM ubuntu:jammy

COPY --from=mokapi /mokapi /mokapi

ENTRYPOINT ["/mokapi"]
```