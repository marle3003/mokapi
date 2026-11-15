---
title: Analyze your mocked HTTP API in Dashboard
description: The dashboard helps you analyze your mocked APIs, how they are used and what data is transferred.
---
# Analyze your mocked HTTP API in Dashboard

The dashboard helps you analyze your mocked APIs, how they are used and what data is transferred.

## Overview of your mocked HTTP API

Mocked HTTP APIs are listed in the HTTP group of the sidebar, with their number of requests. A
red dot marks an API that handled requests with errors. Click the entry to open the detailed
view in a tab. In this detailed view you can see:
- a summary about this API
- server and base URLs where this API is available
- all API endpoints with the corresponding metrics.
- all configurations that affect this API
- recent HTTP requests

Click the **HTTP** heading to see the metrics and the log of all HTTP APIs.

In the next image you can see an example of a REST API showing four Paths one is deprecated
and therefore marked with a warning icon. In the Recent Requests section, you can see
four HTTP requests, one of which has a validation error that is counted in the Errors metric.

<img src="/docs/http/dashboard-mock-rest-api.jpg" width="700" alt="Dashboard shows a REST API including metric, recent requests and config file." title="" />

## Find and open endpoints

If you mock many APIs, use the filter at the top of the sidebar. It matches the API name, the
path of an endpoint, the HTTP method (for example `delete`) and the tag. Each API shows at most
6 matching endpoints, followed by "N more in `<API>`" which opens the API.

A click on an endpoint row replaces the active tab. Use Ctrl+click, Cmd+click or middle click
to open the endpoint in a new tab, for example to compare two endpoints. The tab title starts
with the HTTP method. Click the star in the row or in the tab to pin the endpoint, so it's
always one click away in the **Pinned** section of the sidebar.

<!-- screenshot: HTTP endpoint tab with method badge, pinned star and the Pinned section in the sidebar -->

### Group endpoints by tag

The endpoint list has a switch **A-Z** / **By tag**. By tag groups the endpoints under the tags
of your OpenAPI specification and shows the tag description next to the heading. Endpoints
without a tag are listed under "Untagged". Mokapi remembers the choice for each API.

See [Using Mokapi Dashboard](/docs/get-started/dashboard.md) for tabs, Pinned and the other
parts of the layout.
