---
title: Using Mokapi Dashboard
description: The dashboard helps you analyze your mocked APIs, how they are used and what data is transferred.
---
# Analyze your mocked HTTP API in Dashboard

The dashboard helps you analyze your mocked APIs, how they are used and what data is transferred.

## Overview of your mocked HTTP API

Mocked HTTP APIs are visible in the HTTP section and clicking on the entry in the
table displays a detailed view. In this detailed view you can see:
- a summary about this API
- server and base URLs where this API is available
- all API endpoints with the corresponding metrics.
- all configurations that affect this API
- recent HTTP requests

In the next image you can see an example of a REST API showing four Paths one is deprecated
and therefore marked with a warning icon. In the Recent Requests section, you can see
four HTTP requests, one of which has a validation error that is counted in the Errors metric.

<img src="/dashboard-mock-rest-api.jpg" width="700" alt="Dashboard shows a REST API including metric, recent requests and config file." title="" />
