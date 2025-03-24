const nodemailer = require("nodemailer");

async function sendEmail() {
    let transporter = nodemailer.createTransport({
        host: "localhost",
        port: 2525, // Mokapi's SMTP server
        secure: true,
        auth: {
            user: "bob",
            pass: "secret",
        },
        tls: {
            rejectUnauthorized: false
        }
    });

    let info = await transporter.sendMail({
        from: '"Test Sender" <test@example.com>',
        to: "recipient@example.com",
        subject: "Hello from Mokapi",
        text: "This is a test email sent via Mokapi's mock SMTP server.",
    });

    console.log("Message sent: %s", info.messageId);
}

sendEmail().catch(console.error);