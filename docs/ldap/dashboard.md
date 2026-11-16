---
title: Analyze your mocked LDAP API in Dashboard
description: The LDAP request log of the Mokapi dashboard lists the requests to your mocked LDAP directory, shows a request next to the table and filters by service and operation.
---
# Analyze your mocked LDAP API in Dashboard

The dashboard shows the requests that clients sent to your mocked LDAP directory, with operation,
DN, status and duration. Use it to check that a bind or search arrived, to read its parameters and
the result, and to find the requests of one service or one operation.

## Metrics

In the new layout, the LDAP page starts with three tiles: **Requests**, **Search Errors** and **Last Request**. **Search Errors** counts the searches that ended with a result code other than `success`, for example an invalid filter, the server size limit or a code set by a script. **Last Request** shows the time of the latest request as relative time (for example "2 minutes ago") with the absolute time below it, or `-` if there was no activity yet. The values cover all services of the protocol, and a value that doesn't exist yet shows `0`. The values update with the normal dashboard refresh. The Classic view keeps the metric card.

## Request log

The request log lists the requests of your LDAP services, newest first, with operation, DN,
criteria, status, time and duration. The log shows the requests that Mokapi still keeps in memory.

You find the log in these places:

- the **Requests** sub tab of an LDAP service. It's the first sub tab and opens by default. The other sub tabs are **Servers** and **Configs**.
- the LDAP page, which you open with the **LDAP** heading of the sidebar. It shows the log of all
  services.
- `/dashboard/ldap/log/<service>`, the log of one service, for example
  `/dashboard/ldap/log/HR Employee Directory`
- `/dashboard/ldap/log`, the log of all LDAP services

In the Requests sub tab, click **Open as tab** to open the log of this service in its own dashboard
tab. A click replaces the active tab or activates the tab if it's already open. Ctrl+click,
Cmd+click or middle click opens a new tab. Click the star in the tab to pin the log, so it's
always in the **Pinned** section of the sidebar. The tab shows "Requests" for the log of all
services and "`<service>` requests" for one service.

In the Classic view, the log has the same panel, filters and **Pause** button. Only the link in a
row opens a browser tab with Ctrl+click, Cmd+click or middle click.

<!-- screenshot: Requests sub tab of an LDAP service with the log, the chips and the Open as tab link -->

### Request details

Click a row to open the details of the request next to the table. The selected row is
highlighted. The top of the panel shows:

- the operation, as a badge, followed by the filter of a Search, the name of a Bind or the DN of
  the other operations. An Unbind has no text.
- the status of the response
- the time

Below the head, the panel has these sub tabs:

| Sub tab  | Content                                                                                      |
|----------|----------------------------------------------------------------------------------------------|
| Request  | The fields of the request. This is the default. Not shown for Unbind and Delete.              |
| Response | The results of a Search. Only shown for Search.                                               |
| Details  | The duration, and the event handlers that ran for the request.                                |

The sub tab stays the same when you select the next request, as long as the next request has it.
Otherwise the panel shows the first sub tab.

In the Response sub tab, click a result to open a dialog with its attributes.

If the response of a failed request has a message or a matched DN, the Details sub tab shows them
as **Message** and **Matched DN**. Requests without them, and an Unbind, which has no response,
show neither field. The Event Handlers section only appears if a handler ran for the request.

Click **Open as tab** to show the request in the full view. A click replaces the active tab,
Ctrl+click or Cmd+click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click
on a row does the same without selecting the row. In the Classic view, only the link in a row
opens a browser tab this way.

Use the arrow keys up and down to select the previous or next request of the current page. Press
`Escape` or click **Close** to close the panel. The keys are ignored while you type in an input or a dialog is open.

The selected request is part of the address (`?request=<id>`). After a reload, the panel opens
again, and you can copy the address to share the request with a colleague. Mokapi only keeps
a limited number of requests in memory. If Mokapi has dropped the request, the panel shows
"LDAP Request not found".

In windows narrower than 768 px, the panel is shown below the table.

<!-- screenshot: request log with the detail panel open on the Request sub tab of a Bind -->
<!-- screenshot: panel of a Search with the Response sub tab and the attributes dialog -->

### Filter requests

Active filters are shown as chips above the table. Click a chip to change its value in a popover,
click the x of a chip to remove the filter. **+ Add filter** lists the filters that aren't active yet:

| Field     | Description                                                                                  |
|-----------|----------------------------------------------------------------------------------------------|
| Service   | Only in the log of all services. Select one service.                                         |
| Operation | Bind, Unbind, Search, Add, Modify, Delete, ModifyDN or Compare                               |

Each filter takes a single value. The log of a service doesn't offer the Service filter, because
it already fixes it.

Mokapi saves the filters in the browser, separately for the log of all services and for each
service. It restores them after a reload.

<!-- screenshot: filter chips, an open popover for Operation and the + Add filter menu -->

### Pause the log

While auto refresh is on, the log reloads by itself. Open the dashboard with `?refresh=<seconds>`,
for example `/dashboard?refresh=20`, to turn it on. Click **Pause** to stop reloading the log, for
example to read a request without new ones pushing it down the table. **Pause** is only shown while auto refresh is on, that is when the URL has `?refresh=<seconds>` with a number above 0. Without it, the log doesn't reload by itself and the button is hidden. The log shows "Paused" until you click
**Resume**, which reloads it at once. Pause only affects this log, other views keep updating.

New requests appear on top while the log runs, and the selected request stays selected. Mokapi
doesn't save the pause. A reload of the page or a switch to another tab runs the log again. If you
change a filter or the page while paused, the log loads the new result and runs again. The demo
dashboard has no Pause button, because it has no live requests.

See [Using Mokapi Dashboard](/docs/get-started/dashboard.md) for tabs, Pinned and the other
parts of the layout.
