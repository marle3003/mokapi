import nodemailer from 'nodemailer';

const transporter = nodemailer.createTransport({
  host: 'localhost',
  port: 8025,
  secure: false,
  tls: { rejectUnauthorized: false }
});

export async function driveMail() {
    await sendNewsletter();
    await sendForgetPassword();
}

async function sendNewsletter() {
    const body = `
        <!-- Card -->
        <table width="600" cellspacing="0" cellpadding="0" style="background:#ffffff; border-radius:8px; overflow:hidden;">

            <!-- Header -->
            <tr>
            <td>
                <img src="cid:headerimg" alt="New Product Banner" width="600" style="display:block; width:100%; height:auto;">
            </td>
            </tr>

            <!-- Title -->
            <tr>
            <td style="padding:30px 40px 10px 40px; text-align:center;">
                <h1 style="margin:0; font-size:26px; color:#222;">New Arrivals Just Landed!</h1>
            </td>
            </tr>

            <tr>
            <td style="padding:0 40px 20px 40px; text-align:center;">
                <p style="margin:0; font-size:16px; color:#555;">
                Fresh styles, cool tech, and exclusive releases — handpicked for you.
                </p>
            </td>
            </tr>

            <!-- Product Grid -->
            <tr>
            <td style="padding:0 40px 30px 40px;">

                <!-- Product 1 -->
                <table width="100%" style="margin-bottom:25px;">
                <tr>
                    <td width="180">
                    <img src="cid:product1" alt="Product 1" width="180" style="border-radius:6px; display:block;">
                    </td>
                    <td style="padding-left:20px; vertical-align:top;">
                    <h3 style="margin:0; font-size:18px; color:#222;">Smartwatch XT Pro</h3>
                    <p style="margin:8px 0; font-size:14px; color:#555;">
                        Sleek, fast, waterproof — your perfect companion for fitness and daily life.
                    </p>
                    <p style="margin:0; font-size:16px; color:#0d6efd; font-weight:bold;">$199.00</p>
                    <a href="https://example.com/products/smartwatch"
                        style="display:inline-block; margin-top:10px; background:#0d6efd; padding:8px 16px; color:#fff; text-decoration:none; border-radius:4px; font-size:14px;">
                        Shop Now
                    </a>
                    </td>
                </tr>
                </table>

                <!-- Product 2 -->
                <table width="100%" style="margin-bottom:25px;">
                <tr>
                    <td width="180">
                    <img src="cid:product2" alt="Product 2" width="180" style="border-radius:6px; display:block;">
                    </td>
                    <td style="padding-left:20px; vertical-align:top;">
                    <h3 style="margin:0; font-size:18px; color:#222;">Wireless Noise-Canceling Headphones</h3>
                    <p style="margin:8px 0; font-size:14px; color:#555;">
                        Crystal-clear sound and long-lasting comfort for work or travel.
                    </p>
                    <p style="margin:0; font-size:16px; color:#0d6efd; font-weight:bold;">$149.00</p>
                    <a href="https://example.com/products/headphones"
                        style="display:inline-block; margin-top:10px; background:#0d6efd; padding:8px 16px; color:#fff; text-decoration:none; border-radius:4px; font-size:14px;">
                        Shop Now
                    </a>
                    </td>
                </tr>
                </table>

                <!-- Product 3 -->
                <table width="100%" style="margin-bottom:25px;">
                <tr>
                    <td width="180">
                    <img src="cid:product3" alt="Product 3" width="180" style="border-radius:6px; display:block;">
                    </td>
                    <td style="padding-left:20px; vertical-align:top;">
                    <h3 style="margin:0; font-size:18px; color:#222;">Urban Style Backpack</h3>
                    <p style="margin:8px 0; font-size:14px; color:#555;">
                        Durable, lightweight, and perfect for work or weekend adventures.
                    </p>
                    <p style="margin:0; font-size:16px; color:#0d6efd; font-weight:bold;">$79.00</p>
                    <a href="https://example.com/products/backpack"
                        style="display:inline-block; margin-top:10px; background:#0d6efd; padding:8px 16px; color:#fff; text-decoration:none; border-radius:4px; font-size:14px;">
                        Shop Now
                    </a>
                    </td>
                </tr>
                </table>

            </td>
            </tr>

            <!-- Footer -->
            <tr>
            <td style="background:#f0f0f0; padding:20px 40px; text-align:center; font-size:13px; color:#777;">
                You are receiving this newsletter because you subscribed to product updates.<br>
                © 2025 Your Shop Name. All rights reserved.
            </td>
            </tr>

        </table>

        </td>`;

    await transporter.sendMail({
        from: '"Bob Miller" <bob.miller@example.com>',
        to: '"Alice Johnson" <alice.johnson@example.com>',
        subject: 'Check Out Our New Arrivals!',
        html: body,
        attachments: [
            {
                filename: "header.jpg",
                path: "./assets/lead.png",
                cid: "headerimg"
            },
            {
                filename: "product1.jpg",
                path: "./assets/watch.png",
                cid: "product1"
            },
            {
                filename: "product2.jpg",
                path: "./assets/headphones.png",
                cid: "product2"
            },
            {
                filename: "product3.jpg",
                path: "./assets/bag.png",
                cid: "product3"
            }
        ],
        auth: {
            user: 'bob.miller',
            pass: 'mysecretpassword123'
        }
    });
}

async function sendForgetPassword() {
    const body = `
        <!-- Wrapper -->
        <table width="100%" cellpadding="0" cellspacing="0" style="padding:20px 0;">
        <tr>
            <td align="center">

            <!-- Card -->
            <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff; color:#000; border-radius:8px; overflow:hidden;">

                <!-- Header -->
                <tr>
                <td style="background:#0d6efd; padding:20px 40px; text-align:center;">
                    <h1 style="color:#ffffff; margin:0; font-size:24px;">Reset Your Password</h1>
                </td>
                </tr>

                <!-- Message -->
                <tr>
                <td style="padding:30px 40px;">

                    <p style="font-size:15px; line-height:1.6;">
                    Hello John Doe,
                    </p>

                    <p style="font-size:15px; line-height:1.6;">
                    We received a request to reset the password for your Example account.
                    If you made this request, simply click the button below to reset your password:
                    </p>

                    <!-- Button -->
                    <div style="text-align:center; margin:30px 0;">
                    <a href="https://example.com/reset-password?token=YOUR_TOKEN"
                        style="background:#0d6efd; color:#fff; padding:14px 28px; text-decoration:none; font-size:16px; border-radius:6px; display:inline-block;">
                        Reset Password
                    </a>
                    </div>

                    <p style="font-size:15px; line-height:1.6;">
                    If you didn’t request a password reset, you can safely ignore this email.
                    Your password will remain unchanged.
                    </p>

                    <p style="font-size:15px; line-height:1.6; margin-top:30px;">
                    Best regards,<br>
                    <strong>Your Example Team</strong>
                    </p>

                </td>
                </tr>

                <!-- Footer -->
                <tr>
                <td style="background:#f0f0f0; padding:20px 40px; text-align:center; font-size:13px; color:#777;">
                    If you’re having trouble with the button above, copy and paste the link into your browser:<br>
                    <span style="color:#555; word-break:break-all;">https://example.com/reset-password?token=YOUR_TOKEN</span><br><br>
                    © 2025 Example. All rights reserved.
                </td>
                </tr>

            </table>

            </td>
        </tr>
        </table>`;

    await transporter.sendMail({
        from: 'zzz@example.com',
        to: '"Bob Miller" <bob.miller@example.com>',
        subject: 'Reset Your Password',
        html: body,
    });
}