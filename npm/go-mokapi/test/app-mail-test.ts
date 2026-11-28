import { app, MailEventMessage, MailMessage, MailRouter, MailboxRouter } from "mokapi";

const router: MailRouter = app.mail();
const scoped: MailRouter = app.api("Mail Server").mail();
const mailbox: MailboxRouter = scoped.mailbox("alice@example.com");

router.message((m: MailEventMessage) => {
    const api: string = m.api;
    const server: string = m.server;
    m.subject = "changed";
    m.body = "changed";
    // @ts-expect-error
    m.api = "";
});
router.message((m) => {
    const t: string = m.subject;
});
router.message(async (m) => {}, { priority: 1, tags: { name: "alice" }, track: true });
router.message(() => {}, { track: (m) => m.subject === "foo" });
// @ts-expect-error
router.message(() => {}, { priority: "high" });
mailbox.message((m) => {}).message((m) => {});
mailbox.message(() => {}, { tags: { name: "alice" } });
// @ts-expect-error
mailbox.send({});
// @ts-expect-error
mailbox.sendAsync({});

const mail: MailMessage = {
    from: [{ address: "bob@example.com" }],
    to: [{ address: "alice@example.com" }],
    messageId: "id",
    subject: "Hello",
    contentType: "text/plain",
    encoding: "utf-8",
    body: "Hi Alice",
    attachments: [],
};
router.send(mail).send(mail);
router.sendAsync(mail).then(r => r.send(mail));
scoped.send({ ...mail, cc: [{ address: "carol@example.com" }], time: new Date() });
// @ts-expect-error
router.send({ ...mail, server: "localhost" });
// @ts-expect-error
router.send({ ...mail, subject: 1 });

// the minimal mail needs only from, to, subject and body
router.send({
    from: [{ address: "bob@example.com" }],
    to: [{ address: "alice@example.com" }],
    subject: "Hello",
    body: "Hi Alice",
});
// @ts-expect-error
router.send({ from: [{ address: "bob@example.com" }], subject: "Hello", body: "Hi Alice" });
