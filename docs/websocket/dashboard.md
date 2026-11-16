---
title: Analyze your mocked WebSocket API in Dashboard
description: The WebSocket message log of the Mokapi dashboard lists the messages of your mocked channels, shows a message next to the table, filters by service, path and client, and logs connection requests.
---
# Analyze your mocked WebSocket API in Dashboard

The dashboard shows the messages that clients and Mokapi exchanged on your mocked WebSocket
channels, with value, direction and the client involved. It also lists the connection requests
(connect and close) of each client. Use it to check that a message arrived, to read its value, to
see why a message was rejected and to find the messages of one path or client.

## Message log

The message log lists the messages of your channels, newest first, with path, value and time.
The log shows the messages that Mokapi still keeps in memory. Messages that failed validation are
listed too and marked red.

You find the log in these places:

- the **Messages** sub tab of a WebSocket service
- the view of a channel, which shows only the messages of this channel. For a channel with a
  template path such as `/chats/{roomId}`, the view of the template shows the messages of all its
  instances and the view of an instance shows only its own.
- the view of a client, which shows only the messages of this client. The log is hidden if the
  client has no messages and no filter is active.
- the WebSocket page, which you open with the **WebSocket** heading of the sidebar. It shows the
  log of all services.
- `/dashboard/websocket/log/<service>`, the log of one service, for example
  `/dashboard/websocket/log/WebSocket Chat API`
- `/dashboard/websocket/log`, the log of all WebSocket services

In the Messages sub tab, click **Open as tab** to open the log of this service in its own dashboard
tab. A click replaces the active tab or activates the tab if it's already open. Ctrl+click,
Cmd+click or middle click opens a new tab. Click the star in the tab to pin the log, so it's
always in the **Pinned** section of the sidebar. The tab shows "Messages" for the log of all
services and "`<service>` messages" for one service.

The Channels sub tab lists the channels only.

In the Classic view, the log has the same panel, filters and **Pause** button. Only the link in a
row opens a browser tab with Ctrl+click, Cmd+click or middle click.

<!-- screenshot: Messages sub tab of a WebSocket service with the log, the chips and the Open as tab link -->

### Message details

Click a row to open the details of the message next to the table. The selected row is
highlighted. The top of the panel shows:

- the path (a link to the channel)
- the time
- the direction, as a badge
- an **Invalid** badge if the message failed validation

The direction badge shows `send` for a message that the client sent to Mokapi and `receive` for a
message that Mokapi sent to the client. Hover over the badge to see this as text.

Below the head, the panel has these sub tabs:

| Sub tab    | Content                                                                                      |
|------------|----------------------------------------------------------------------------------------------|
| Value      | The value of the message. This is the default.                                               |
| Details    | Client, source, destination and content type. Also the event handlers that ran for the message. |
| Validation | The validation errors of the message. Only shown for invalid messages.                       |

The Details sub tab also shows a **Script** field for messages that a script sent with
`mokapi.websocket(...).send()`. It links to the script. Messages that an event handler sent with
`client.send`, `reply` or `broadcast` have no Script field. The Event Handlers section only appears
if a handler ran for the message.

Click **Open as tab** to show the message in the full view. A click replaces the active tab,
Ctrl+click or Cmd+click opens a new one. In the new layout, Ctrl+click, Cmd+click or middle click
on a row does the same without selecting the row. In the Classic view, only the link in a row
opens a browser tab this way.

Use the arrow keys up and down to select the previous or next message of the current page. Press
`Escape` or click **Close** to close the panel. The keys are ignored while you type in an input or a dialog is open.

The selected message is part of the address (`?message=<id>`). After a reload, the panel opens
again, and you can copy the address to share the message with a colleague. Mokapi only keeps
a limited number of messages in memory. If Mokapi has dropped the message, the panel shows
"WebSocket Message not found".

In windows narrower than 768 px, the panel is shown below the table.

<!-- screenshot: message log with the detail panel open on the Value sub tab -->
<!-- screenshot: panel of an invalid message with the Validation sub tab -->

### Filter messages

Active filters are shown as chips above the table. Click a chip to change its value in a popover,
click the x of a chip to remove the filter. **+ Add filter** lists the filters that aren't active yet:

| Field   | Description                                                      |
|---------|------------------------------------------------------------------|
| Service | Only in the log of all services. Select one service.             |
| Path    | The path of a channel. The input suggests the known template paths. |
| Client  | The client id                                                    |

Each filter takes a single value. A view doesn't offer the fields it already fixes: the log of a
service has no Service filter, the view of a channel has no Path filter and the client view has no
Client filter.

For a channel with a template path such as `/chats/{roomId}`, the Path filter takes the template
path and matches the messages of all its instances. A concrete path like `/chats/general` matches
nothing.

Mokapi saves the filters in the browser, separately for the log of all services, for each
service, for each channel (and instance of a template channel) and for each client. It restores
them after a reload.

<!-- screenshot: filter chips, an open popover for Path and the + Add filter menu -->

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

## Connection request log

The view of a WebSocket client also lists the connection requests of this client in the
"Connection Requests" table. A row is a `Connect` or a `Close` event with channel, reason, who
closed the connection and time. Each row opens in a detail panel, in the same way as a message.

### Request details

Click a row to open the details of the request next to the table. The top of the panel shows the
type as a badge, for example `Connect`, and the time. Below it, the panel shows:

- the channel (a link to the channel)
- the client
- for a `Close` event, the reason and who closed the connection (**Closed By**)
- the event handlers that ran, if there are any

Click **Open as tab** to show the request in the full view. A click replaces the active tab,
Ctrl+click, Cmd+click or middle click opens a new one. In the new layout, Ctrl+click, Cmd+click or
middle click on a row does the same without selecting the row. In the Classic view, only the link
in a row opens a browser tab this way. Arrow keys up and down and `Escape` work as in the
[message details](#message-details).

The selected request is part of the address (`?request=<id>`). After a reload, the panel opens
again, and you can copy the address to share the request. If Mokapi has dropped the request, the
panel shows "WebSocket Event not found".

A client view has one selection. Selecting a request closes the message panel and selecting a
message closes the request panel. **Pause** and **Resume** work as described in
[Pause the log](#pause-the-log) and also apply to the request log.

<!-- screenshot: client view with the Connection Requests table and the detail panel of a Close event -->

See [Using Mokapi Dashboard](/docs/get-started/dashboard.md) for tabs, Pinned and the other
parts of the layout.
