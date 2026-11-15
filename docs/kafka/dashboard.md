---
title: Analyze your mocked Kafka API in Dashboard
description: The Kafka message log of the Mokapi dashboard lists the messages of your mocked topics, shows a message next to the table and filters by cluster, topic, partition and client.
---
# Analyze your mocked Kafka API in Dashboard

The dashboard shows the messages that were produced to your mocked Kafka topics, with key, value,
headers and the client that sent them. Use it to check that a message arrived, to read its value
and to find messages of one topic, partition or client.

## Message log

The message log lists the messages of your topics, newest first, with topic, key, value and time.
The log shows the messages that Mokapi still keeps in memory.

You find the log in these places:

- the **Messages** sub tab of a Kafka cluster
- the **Messages** tab of a topic, which shows only the messages of this topic
- the view of a client, which shows only the messages this client produced. The log is hidden if
  the client has no messages and no filter is active.
- the Kafka page, which you open with the **Kafka** heading of the sidebar. It shows the log of
  all clusters.
- `/dashboard/kafka/log/<cluster>`, the log of one cluster, for example
  `/dashboard/kafka/log/Orders`
- `/dashboard/kafka/log`, the log of all Kafka clusters

In the Messages sub tab, click **Open as tab** to open the log of this cluster in its own dashboard
tab. A click replaces the active tab or activates the tab if it's already open. Ctrl+click,
Cmd+click or middle click opens a new tab. Click the star in the tab to pin the log, so it's
always in the **Pinned** section of the sidebar. The tab shows "Messages" for the log of all
clusters and "`<cluster>` messages" for one cluster.

The Topics sub tab lists the topics only. Selecting tags there filters the topic table, not the
message log.

<!-- screenshot: Messages sub tab of a Kafka cluster with the log, the chips and the Open as tab link -->

### Message details

Click a row to open the details of the message next to the table. The selected row is
highlighted. The top of the panel shows the key, the topic (a link to the topic), the offset, the
partition and the time. Below it, the panel has three sub tabs:

| Sub tab | Content                                                                                      |
|---------|----------------------------------------------------------------------------------------------|
| Value   | The value of the message. This is the default.                                               |
| Headers | The headers of the message                                                                   |
| Details | Key type, client, content type and, if set, producer id, epoch and sequence number. Also the event handlers that ran for the message. |

Click **Open as tab** to show the message in the full view. A click replaces the active tab,
Ctrl+click or Cmd+click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click
on a row does the same without selecting the row.

Use the arrow keys up and down to select the previous or next message of the current page. Press
`Escape` or click **Close** to close the panel. The keys are ignored while you type in an input.

The selected message is part of the address (`?message=<id>`). After a reload, the panel opens
again, and you can copy the address to share the message with a colleague. Mokapi only keeps
a limited number of messages in memory. If Mokapi has dropped the message, the panel shows
"Kafka Message not found".

In windows narrower than 768 px, the panel is shown below the table.

<!-- screenshot: message log with the detail panel open on the Value sub tab -->

### Filter messages

Active filters are shown as chips above the table. Click a chip to change its value in a popover,
click the x of a chip to remove the filter. **+ Add filter** lists the filters that aren't active yet:

| Field     | Description                                                     |
|-----------|-----------------------------------------------------------------|
| Cluster   | Only in the log of all clusters. Select one cluster.            |
| Topic     | The name of a topic. The input suggests the known topics.       |
| Partition | The number of a partition                                       |
| Client    | The client id of the producer                                   |

Each filter takes a single value. A view doesn't offer the fields it already fixes: the log of a
cluster has no Cluster filter, the topic view has no Topic filter and the client view has no Client
filter.

Two cases to keep in mind:

- The Client filter doesn't match messages that a script produced without a client, for example
  with `produce()` from `mokapi/kafka`, because those messages have no client id.
- For a topic with a template name such as `shipments.{region}.updates`, the Topic filter takes the
  template name and matches the messages of all its instances.

Mokapi saves the filters in the browser, separately for the log of all clusters, for each
cluster, for each topic (and instance of a template topic) and for each client. It restores them
after a reload.

<!-- screenshot: filter chips, an open popover for Partition and the + Add filter menu -->

### Pause the log

While auto refresh is on, the log reloads by itself. Open the dashboard with `?refresh=<seconds>`,
for example `/dashboard?refresh=20`, to turn it on. Click **Pause** to stop reloading the log, for
example to read a message without new ones pushing it down the table. Without auto refresh, the log
doesn't reload by itself and **Pause** has no effect. The log shows "Paused" until you click
**Resume**, which reloads it at once. Pause only affects this log, other views keep updating.

New messages appear on top while the log runs, and the selected message stays selected. Mokapi
doesn't save the pause. A reload of the page or a switch to another tab runs the log again. If you
change a filter or the page while paused, the log loads the new result and runs again. The demo
dashboard has no Pause button, because it has no live messages.

## Client request log

The view of a Kafka client also lists the requests of this client in the "Recent Requests" table. Each row opens in a detail panel, in the same way as a message.

### Request details

Click a row to open the details of the request next to the table. The top of the panel shows the API key with its number, the version and the time. Below it, the panel has the sub tabs **Request** and **Response**. The sub tab stays the same when you select the next request.

Click **Open as tab** to show the request in the full view. A click replaces the active tab, Ctrl+click, Cmd+click or middle click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click on a row does the same without selecting the row. Arrow keys up and down and `Escape` work as in the [message details](#message-details).

The selected request is part of the address (`?request=<id>`). After a reload, the panel opens again, and you can copy the address to share the request. If Mokapi has dropped the request, the panel shows "Kafka Request not found".

A client view has one selection. Selecting a request closes the message panel and selecting a message closes the request panel. **Pause** and **Resume** work as described in [Pause the log](#pause-the-log) and also apply to the request log.

<!-- screenshot: client view with the Recent Requests table and the detail panel on the Request sub tab -->

See [Using Mokapi Dashboard](/docs/get-started/dashboard.md) for tabs, Pinned and the other
parts of the layout.
