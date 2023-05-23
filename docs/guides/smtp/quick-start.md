# Quick Start

A simple Use Case to use Mokapi as a fake smtp server

## Define SMTP server
First, create a configuration file `smtp.yaml`

```yaml
smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtp://127.0.0.1:25
```

With these configuration Mokapi will start an smtp server on port 25.

## Create a Dockerfile
Next create a `Dockerfile` to configure Mokapi
```dockerfile
FROM mokapi/mokapi:latest

COPY ./smtp.yaml /demo/

CMD ["--Providers.File.Directory=/demo"]
```

## Start Mokapi

```
docker run -p 8080:8080 -p 8025:25 --rm -it $(docker build -q .)
```

You can now open a browser and go to Mokapi's Dashboard (`http://localhost:8080`) to see the Mokapi's SMTP server and the received messages.

## Use a SMTP client

```c#
using System.Net.Mail;

string to = "alice@mokapi.io";
string from = "bob@mokapi.io";
string subject = "Using the new SMTP client.";
string body = "Using Mokapi SMTP server, you can send an email message from any application very easily.";

MailMessage message = new(from, to, subject, body);

using SmtpClient client = new SmtpClient("127.0.0.1", 8025);
client.Send(message);
```