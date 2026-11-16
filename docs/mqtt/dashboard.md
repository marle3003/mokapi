---
title: Analyze your mocked MQTT API in Dashboard
description: The MQTT message log of the Mokapi dashboard lists the messages published to your mocked topics, shows a message next to the table and filters by cluster, topic and client.
---
# Analyze your mocked MQTT API in Dashboard

The dashboard shows the messages that were published to your mocked MQTT topics, with value,
retain flag and the client that sent them. Use it to check that a message arrived, to read its
value and to find messages of one topic or client.

## Metrics

In the new layout, the MQTT page starts with three tiles: **Messages**, **Topics** and **Last Message**. **Topics** is the number of topics. **Last Message** shows the time of the latest message as relative time (for example "2 minutes ago") with the absolute time below it, or `-` if there was no activity yet. The values cover all services of the protocol, and a value that doesn't exist yet shows `0`. The values update with the normal dashboard refresh. The Classic view keeps the metric card.

## Message log

The message log lists the messages of your topics, newest first, with topic, value and time.
The log shows the messages that Mokapi still keeps in memory.

You find the log in these places:

- the **Messages** sub tab of an MQTT cluster
- the **Messages** tab of a topic, which shows only the messages of this topic
- the view of a client, which shows only the messages this client published. The log is hidden if
  the client has no messages and no filter is active.
- the MQTT page, which you open with the **MQTT** heading of the sidebar. It shows the log of
  all clusters.
- `/dashboard/mqtt/log/<cluster>`, the log of one cluster, for example
  `/dashboard/mqtt/log/Sensor Service`
- `/dashboard/mqtt/log`, the log of all MQTT clusters

In the Messages sub tab, click **Open as tab** to open the log of this cluster in its own dashboard
tab. A click replaces the active tab or activates the tab if it's already open. Ctrl+click,
Cmd+click or middle click opens a new tab. Click the star in the tab to pin the log, so it's
always in the **Pinned** section of the sidebar. The tab shows "Messages" for the log of all
clusters and "`<cluster>` messages" for one cluster.

The Topics sub tab lists the topics only. Selecting tags there filters the topic table, not the
message log.

<!-- screenshot: Messages sub tab of an MQTT cluster with the log, the chips and the Open as tab link -->

### Message details

Click a row to open the details of the message next to the table. The selected row is
highlighted. The top of the panel shows the topic (a link to the topic) and the time. Messages
that were published with the retain flag also show a **Retained** badge. Below it, the panel has
two sub tabs:

| Sub tab | Content                                                                                      |
|---------|----------------------------------------------------------------------------------------------|
| Value   | The value of the message. This is the default.                                               |
| Details | Client and content type. Also the event handlers that ran for the message.                   |

Click **Open as tab** to show the message in the full view. A click replaces the active tab,
Ctrl+click or Cmd+click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click
on a row does the same without selecting the row. In the Classic view, only the link in a row
opens a browser tab this way.

Use the arrow keys up and down to select the previous or next message of the current page. Press
`Escape` or click **Close** to close the panel. The keys are ignored while you type in an input or a dialog is open.

The selected message is part of the address (`?message=<id>`). After a reload, the panel opens
again, and you can copy the address to share the message with a colleague. Mokapi only keeps
a limited number of messages in memory. If Mokapi has dropped the message, the panel shows
"MQTT Message not found".

In windows narrower than 768 px, the panel is shown below the table.

<!-- screenshot: message log with the detail panel open on the Value sub tab -->

### Filter messages

Active filters are shown as chips above the table. Click a chip to change its value in a popover,
click the x of a chip to remove the filter. **+ Add filter** lists the filters that aren't active yet:

| Field   | Description                                                     |
|---------|-----------------------------------------------------------------|
| Cluster | Only in the log of all clusters. Select one cluster.            |
| Topic   | The name of a topic. The input suggests the known topics.       |
| Client  | The client id of the publisher                                  |

Each filter takes a single value. A view doesn't offer the fields it already fixes: the log of a
cluster has no Cluster filter, the topic view has no Topic filter and the client view has no Client
filter.

Two cases to keep in mind:

- For a topic with a template name such as `sensors/{sensorId}/data`, the Topic filter takes the
  template name and matches the messages of all its instances. A concrete name like
  `sensors/12345/data` matches nothing.
- Messages that a script published, for example with `publish()` from `mokapi/mqtt`, have the client
  `mokapi-script`. Use it as the value of the Client filter to find them.

Mokapi saves the filters in the browser, separately for the log of all clusters, for each
cluster, for each topic (and instance of a template topic) and for each client. It restores them
after a reload.

<!-- screenshot: filter chips, an open popover for Topic and the + Add filter menu -->

### Pause the log

While auto refresh is on, the log reloads by itself. Open the dashboard with `?refresh=<seconds>`,
for example `/dashboard?refresh=20`, to turn it on. Click **Pause** to stop reloading the log, for
example to read a message without new ones pushing it down the table. **Pause** is only shown while auto refresh is on, that is when the URL has `?refresh=<seconds>` with a number above 0. Without it, the log doesn't reload by itself and the button is hidden. The log shows "Paused" until you click
**Resume**, which reloads it at once. Pause only affects this log, other views keep updating.

New messages appear on top while the log runs, and the selected message stays selected. Mokapi
doesn't save the pause. A reload of the page or a switch to another tab runs the log again. If you
change a filter or the page while paused, the log loads the new result and runs again. The demo
dashboard has no Pause button, because it has no live messages.

## Client request log

The view of an MQTT client also lists the requests of this client in the "Connection & Subscription Requests" table. Each row opens in a detail panel, in the same way as a message.

### Request details

Click a row to open the details of the request next to the table. The top of the panel shows the type, for example `Connect`, and the time. Below it, the panel has the sub tabs **Request** and **Response**. A `Disconnect` request has no response, so the panel shows only the Request sub tab. The sub tab stays the same when you select the next request.

Click **Open as tab** to show the request in the full view. A click replaces the active tab, Ctrl+click, Cmd+click or middle click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click on a row does the same without selecting the row. In the Classic view, only the link in a row opens a browser tab this way. Arrow keys up and down and `Escape` work as in the [message details](#message-details).

The selected request is part of the address (`?request=<id>`). After a reload, the panel opens again, and you can copy the address to share the request. If Mokapi has dropped the request, the panel shows "MQTT Request not found".

A client view has one selection. Selecting a request closes the message panel and selecting a message closes the request panel. **Pause** and **Resume** work as described in [Pause the log](#pause-the-log) and also apply to the request log.

<!-- screenshot: client view with the Connection & Subscription Requests table and the detail panel on the Request sub tab -->

See [Using Mokapi Dashboard](/docs/get-started/dashboard.md) for tabs, Pinned and the other
parts of the layout.
