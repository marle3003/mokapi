---
title: "Mocking a Payment Service with Mokapi: A Real Checkout Flow for Playwright"
description: Learn how to mock a payment provider with Mokapi using real HTML pages and Kafka events, so Playwright tests can walk through an actual checkout flow.
subtitle: Most teams skip testing the checkout flow because they can't control a real payment service in their test environment. Here's how Mokapi removes that excuse.
image:
  url: "/mock-api-payment-flow.png"
  alt: Diagram showing a shop checkout flow where Mokapi mocks both the payment provider API and a hosted payment page, with a Kafka topic carrying the payment result back to the shop backend.
---

# Your Playwright Tests Deserve a Real Payment Flow (Without the Real Payment Service)

You know that moment when you're setting up e2e tests for a checkout flow and you just... stop? Because wiring up a
real payment service for tests feels like a rabbit hole nobody has time for. So you mock the API response, skip the
actual payment page, and tell yourself "we'll figure out the real integration later."

And "later" never really comes, does it.

So your tests cover everything around the payment. The cart, the product page, the order confirmation. But the actual
checkout? That's just... trusted. You ship it and hope.

And it's not only the tests. Think about that moment in a sprint review when a stakeholder wants to see the checkout
flow. Someone has to say "hang on, let me hack a successful payment somehow" and suddenly the demo is a developer
fumbling around in a sandbox account. Not a great look. Or a new developer joins the team and wants to understand
how the payment flow works. Same problem. There's no clean way to just... try it.

I've seen this pattern a lot. And honestly it makes sense. Integrating with a real payment provider in a test
environment means test accounts, API keys that expire, flaky sandbox environments, and data you can't fully control.
It's a mess. So people avoid it.

But here's what I kept thinking: what if the mock wasn't just a stubbed API response, but an actual page? A real form
that Playwright navigates, fills in, and submits. Just like a real user would. No shortcuts, no intercepted requests,
no pretending.

That's what we're going to build. A fully mocked payment service using Mokapi that serves real HTML, fires a Kafka
message, and lets your Playwright tests walk through the entire checkout flow end to end. Including the async part
where the backend processes the payment and the frontend finally shows "paid."

And because it's a real page, anyone can use it. A stakeholder, a designer, a developer on their first day. No setup,
no hacking, no "wait let me fix this first." Just click pay and watch the whole thing work.

Here's what the flow looks like in practice:

![Demo of a payment workflow using mock API](https://raw.githubusercontent.com/marle3003/mokapi-payment-workflow/main/mokapi-payment-workflow.gif "Payment workflow visualized via mock API")

If you've used Mokapi before for basic REST or Kafka mocking, this might change how you think about what a mock can actually be.

## Mokapi can serve HTML. Here's why that changes everything.

Most developers, when they hear "mock," think of something like this: intercept the request, return a fake response,
move on. And that works fine for a lot of cases. Your code calls an endpoint, gets back a JSON blob, renders something.
Done.

But a payment flow isn't really like that. It's not just an API call. It's a whole experience. The user leaves your
app, lands on a payment page, fills in their card details, hits pay, and gets redirected back. There's a form. There's
validation. There's a redirect. There's an async event that fires on the backend after the payment processes.

If you stub just the API response you're not really testing that flow. You're testing a shortcut version of it. And
shortcuts have a way of hiding bugs.

Here's where Mokapi does something most people don't realize it can do. It can serve HTML pages. Not just respond
to API calls, but actually act as a web server. So instead of intercepting the redirect to the payment provider and
returning a fake success response, you let the redirect happen. Playwright lands on a real page, served by Mokapi.
It fills in a form. It clicks pay. Mokapi handles the POST, produces a Kafka message, and redirects back to your app.

From Playwright's perspective, nothing is mocked. It just navigated a checkout flow.

Think about what that means. Your tests are now exercising the actual redirect behavior. The actual form submission.
The actual async processing. The only thing that's fake is the payment provider itself, and honestly that's the only
part you want to be fake.

And if the mock behaves like the real service, your tests don't need to know they're talking to a mock. That's kind
of the whole point.

## Meet the example

Alright, let's get concrete. Abstract ideas are nice, but I'd rather show you something you can actually run.

The example we're building is a simple online shop. Nothing fancy. A frontend, a backend, a shopping cart and a
checkout page. The kind of thing you've probably built or worked on a dozen times. The interesting part isn't the
shop itself, it's what happens when you hit pay.

In a real setup your backend would call the payment provider's API to create a checkout session, Stripe for example,
passing the order amount and a URL to redirect to once payment completes. The provider returns a payment URL, and
your frontend redirects the customer there. After payment, the provider fires an event to your backend, typically
via a webhook or in our case a Kafka topic, to confirm the payment status. Your backend consumes that event, updates
the order, and your frontend shows "paid."

In our setup, Mokapi plays the role of that payment provider. Completely. There are actually two separate mocked
services here, and that separation matters. One mocks the payment provider's API, the endpoint your backend calls
to create a session. The other mocks the hosted payment page itself, the one the customer's browser actually lands
on. In production these would be two different surfaces of the same provider. In our mock they're two different Mokapi
configs, which keeps the example honest about who owns what.

Here's the full picture of what's running:

- The SUT: a frontend and a backend for the online shop
- Mokapi: serving the payment provider API, the payment page, and mocking the Kafka topic
- Playwright: running the e2e test, navigating the whole flow

![Diagram showing a shop checkout flow where Mokapi mocks both the payment provider API and a hosted payment page, with a Kafka topic carrying the payment result back to the shop backend](/mock-api-payment-flow.png "How Mokapi mocks a payment service: API, HTML page, and Kafka event")

The complete setup is in the [GitHub repository](https://github.com/marle3003/mokapi-payment-workflow). I'll show the relevant pieces here so you understand what's going on,
but if you want to dig into the full config and scripts, that's where to go.

Let's start walking through it.

## The simulation API: your test's remote control

Sometimes you want to test something specific about the payment flow. Not just the happy path, but what happens during
those few seconds while the payment is processing. Does the frontend show a spinner? Does it say "processing"? Or does
it just sit there with no feedback at all?

That's what the simulation API is for. One endpoint. `POST /simulate/payment-delay` lets you configure a delay before
Mokapi produces the Kafka message for a specific order. So you can deliberately slow things down and assert what the
frontend shows in that window.

```javascript
await request.post('/simulate/payment-delay', {
  data: {
    orderId: orderId,
    delayMs: 3000
  }
})
```

Notice it's scoped to a specific `orderId`, not a global setting. Playwright tests often run in parallel, so a delay
configured for one test shouldn't leak into another. Each test reads its own `orderId` from the checkout page and
configures the delay just for that order.

That "processing" state is really easy to forget about when you're building a feature. And even easier to forget to
test. But it's what real users see, sometimes for several seconds. With the delay simulation you can test it
deliberately, every time, without any timing luck involved.

## The payment page

So what does Playwright actually land on when it gets redirected to the mock payment service? A real HTML form. Credit
card number, expiry date, CVV. In a real project your team would decide how closely this mirrors the actual payment
provider. If you're mocking Stripe you might want to replicate the full Stripe checkout experience, every field, every
step, so the test and the demo feel identical to production. That's completely possible with this approach. For this
article we're keeping the form simple so we can focus on the mechanics without getting lost in details.

Mokapi serves this page via a GET endpoint defined in an OpenAPI spec. If you haven't done this before it might feel
a bit unusual because we're used to OpenAPI describing JSON APIs. But it works just the same for HTML responses. You
define the path, the method, and tell it to return `text/html`. Mokapi takes care of the rest.

```yaml
paths:
  /payment:
    get:
      summary: Show the payment form
      parameters:
        - name: sessionId
          in: query
          required: true
          schema:
            type: string
          description: The session ID created by the payment provider API
      responses:
        '200':
          description: Payment form
          content:
            text/html:
              schema:
                type: string
    post:
      summary: Process the payment form submission
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                sessionId:
                  type: string
                cardNumber:
                  type: string
                expiry:
                  type: string
                cvv:
                  type: string
      responses:
        '302':
          description: Redirect to successUrl after payment
          headers:
            location:
              schema:
                type: string
```

The actual HTML is returned by a Mokapi script. And this is where it gets interesting because the script isn't just
returning a static string. It's a real handler that receives the request, builds the response, and has full control
over what comes back. The form shows the order total and order ID, both of which Mokapi looks up from the session that
was created when your backend called the payment provider API.

Simple. Clean. Playwright lands here, sees a form, fills it in. No difference from landing on a real payment page.
The full version with styling and error states is in the [GitHub repo](https://github.com/marle3003/mokapi-payment-workflow)
but this gives you the shape of it.

## What happens when you hit Pay

This is where it gets interesting. When Playwright submits the form, Mokapi's script takes over. And it's doing more
than just returning a success response. It's behaving like a real payment service would.

It looks up the session, takes the order ID, calls `produce()` to fire a Kafka message onto the `order-payment-state`
topic, and redirects back to the shop using the success URL that was stored when the session was created.

```javascript
export default function() {
    on('http', (req, res) => {
        if (req.api !== 'Payment App') {
            return
        }

        if (req.method === 'GET') {
            const sessionId = req.query.sessionId;
            const session = sessions[sessionId]

            if (!session) {
                res.statusCode = 400
                res.data = 'Session not found'
                return
            }

            const html = render(paymentPageTemplate, { sessionId, ...session, error: undefined })
            res.data = html
        } else if (req.method === 'POST') {
            let { sessionId, delay } = req.body
            const session = sessions[sessionId]

            if (!session) {
                res.statusCode = 200
                const html = render(paymentPageTemplate, { sessionId, error: 'Session not found' })
                res.data = html
                return
            }

            const kafkaProduceRequest = {
                topic: 'order-payment-state',
                messages: [
                    {
                        key: session.orderId,
                        data: {
                            orderId: session.orderId,
                            paymentStatus: 'success'
                        }
                    }
                ]
            }

            if (!delay && simulations.payment[session.orderId]) {
                delay = simulations.payment[session.orderId].delayMs
            }

            if (delay > 0) {
                delete simulations.payment[session.orderId]

                setTimeout(() => {
                    produce(kafkaProduceRequest)
                }, delay)
            } else {
                try {
                    produce(kafkaProduceRequest)
                } catch (err) {
                    const html = render(paymentPageTemplate, { sessionId, error: `${err}` })
                    res.data = html
                    return
                }
            }

            res.statusCode = 302
            res.headers = { Location: session.successUrl }
        }
    })
}
```

One small but useful detail. If `produce()` fails because the message doesn't match the topic schema, that error
surfaces directly on the payment page. So if something is wrong a tester sees it immediately right there, rather than
staring at an order that never updates and wondering what happened.

And that Kafka message is real. Your SUT backend is listening to that topic just like it would in production. It
consumes the message, updates the order status, and the frontend reflects that. Mokapi didn't fake the outcome. It
triggered the actual processing chain.

That's the part worth sitting with for a second. The mock isn't returning a pre-cooked response that bypasses your
backend logic. It's producing an event that your backend has to actually handle. If there's a bug in how your backend
processes that Kafka message, this test will catch it.

## The async part: waiting like a real user

So Playwright just got redirected back to the shop's confirmation page. And here's the thing: the order status might
not be "paid" yet. The backend still needs to consume the Kafka message and update the order. That takes a moment. Not
long, but a moment.

And you know what? That's fine. That's exactly how it works in production too.

This is something I think is worth being explicit about in your tests. Don't hide the async nature of the flow.
Playwright has built in waiting mechanisms that handle this really naturally. Instead of a hardcoded sleep you just
wait for the element that shows the payment status to appear with the right value.

```javascript
await expect(page.getByRole('heading', { name: 'Payment confirmed' })).toBeVisible({ timeout: 30000 });
```

Playwright will keep checking until it sees "paid" or the timeout hits. Clean, readable, and it mirrors what a real
user experiences. They land on the confirmation page and wait a second for it to update.

Now here's where the delay simulation becomes really useful. Remember `POST /simulate/payment-delay`? If you configure
a three-second delay before Mokapi produces the Kafka message for a specific order, you can test what the frontend
shows during that window.

```javascript
await request.post(`${simulationUrl}/simulate/payment-delay`, {
    headers: { 'Content-Type': 'application/json' },
    data: { 
        orderId,
        delayMs: 3000
    },
});

// Now submit the payment form
// Assert the pending state appears first
await expect(page.getByRole('heading', { name: 'Processing your payment' })).toBeVisible();

// Then wait for it to resolve
await expect(page.getByRole('heading', { name: 'Payment confirmed' })).toBeVisible({ timeout: 30000 });
```

That "processing" state is really easy to forget about when you're building a feature. And even easier to forget to
test. With the delay simulation you can test it deliberately, every time, without any timing luck involved.

## What Playwright actually sees

Let's zoom out and look at the full test. Because after all the setup we've talked about, the actual Playwright code
tells a complete story.

```javascript
test(name, async ({ page, request, simulationUrl }) => {
    await page.goto('');

    await test.step('Add product to cart', async () => {
        const article = page.getByRole('article', { name: 'Wireless Headphones' })
        await article.getByRole('button', { name: 'Add to cart' }).click();
    });

    await test.step('Verify order details on checkout page', async () => {
        const nav = page.getByRole('navigation');
        await nav.getByRole('button', { name: 'Cart' }).click();

        const shoppingCart = page.getByRole('region', { name: 'Shopping cart items' });
        const cartItems = shoppingCart.getByRole('listitem');
        await expect(cartItems).toHaveCount(1);
        await expect(cartItems.first()).toContainText('Wireless Headphones');
        await expect(cartItems.first()).toContainText('$79.99');
    });

    await test.step('Proceed to checkout', async () => {
        await page.getByRole('button', { name: 'Proceed to checkout' }).click();

        const orderId = await page.getByLabel('Order ID').textContent()

        await request.post(`${simulationUrl}/simulate/payment-delay`, {
            headers: { 'Content-Type': 'application/json' },
            data: { orderId, delayMs: delay },
        });
    })

    await test.step('Pay and verify order completion', async () => {
        await page.getByRole('button', { name: 'Pay' }).click();

        await page.getByRole('textbox', { name: 'Card number' }).fill('4242424242424242');
        await page.getByRole('textbox', { name: 'Expiry' }).fill('12/34');
        await page.getByRole('textbox', { name: 'CVC' }).fill('123');
        await page.getByRole('button', { name: 'Pay' }).click();

        if (delay > 0) {
            await expect(page.getByRole('heading', { name: 'Processing your payment' })).toBeVisible();
        }

        await expect(page.getByRole('heading', { name: 'Payment confirmed' })).toBeVisible({ timeout: 30000 });
    })
});
```

Read it top to bottom and it tells a story. Open the shop, add a product, go to the cart, proceed to checkout,
click pay, get redirected to the payment page, fill in the card details, submit, land back on the shop, wait for
confirmation. A new developer on the team can read this and immediately understand the full user journey.

And that redirect in the middle is important. The test isn't jumping directly to the payment form. It clicks the pay
button on the shop's checkout page, the backend creates a session with the payment provider and returns a URL, and
the redirect happens naturally. That means you're also testing that your backend correctly creates the session and
constructs the right redirect URL. That's real behavior being exercised, not skipped.

Notice what the test doesn't know though. It doesn't know Mokapi is serving the payment page. It doesn't know a Kafka
message is being produced behind the scenes. It doesn't know the backend is consuming that message asynchronously.
It just navigates, interacts, and asserts. Exactly like a real user would.

That separation is what makes this approach so valuable. Your tests describe behavior, not implementation. If you
eventually point the tests at a real payment provider, the test itself doesn't change. The flow is the flow.

## Why this matters more than you think

Look, there are a lot of ways to fake a payment flow in tests. You can stub the API response, you can skip the checkout
entirely and just seed a "paid" order directly in the database, you can flag the user as a test user and bypass the
whole thing in code. These approaches work.

But they all have the same blind spot. They're not testing the flow. They're testing around it.

And the flow is where bugs live. The redirect that sends the wrong order ID. The frontend that doesn't handle the
"processing" state before the Kafka message arrives. The backend that has a subtle bug in how it parses the payment
event. None of these show up when you stub a response or seed a database row. They show up in production, in front
of real users, with real money involved.

That's what makes this approach different. You're not testing around the payment flow. You're walking through it,
every single time, in a way that's repeatable and automated and available to anyone on the team.

And honestly that last part is underrated. Because it's not just about the tests. A stakeholder can open the app in
a staging environment and click through a complete checkout without anyone having to set anything up. A designer can
click through the checkout and verify the confirmation page looks right once the payment status comes back. A new
developer can run the Playwright suite on their first day and watch the whole flow execute. No sandbox accounts, no
test credit card numbers buried in a wiki, no "let me just quickly hack this so you can see it."

The mock is the service. It's always there, it always works, and anyone can use it.

That's a different relationship with your test environment than most teams have. And once you've worked this way it's
hard to go back.

## What's next

We covered a lot of ground here. A mocked payment provider API and a mocked payment page that Playwright navigates
like a real user. A Kafka topic that gets a real message produced onto it. A simulation API that gives your tests a
remote control for the scenarios they need. And a test that reads like a user story, not a list of API stubs.

But payment flows are just one example of this pattern. Any external service that takes your user somewhere else and
brings them back is a candidate for this approach. And the one that comes up almost as often as payments is identity.
Login flows, single sign-on, OAuth redirects. The same problem: your user leaves your app, authenticates somewhere
else, and comes back with a token.

That's what the next article is about. Mocking an IDP with Mokapi. Serving a real login form, handling the redirect
back to your app, and getting into the slightly thornier territory of tokens and how your backend accepts them in a
test environment. There are a few different approaches there and they're worth exploring properly.

For now, grab the [GitHub repository](https://github.com/marle3003/mokapi-payment-workflow) and have a look at the
full example. The Mokapi config, the scripts, the Playwright tests, all of it is there. Run it, break it, adapt it to
your own setup. And if you've been avoiding testing your payment flow because it felt too hard to mock properly, well.
Maybe it's not as hard as you thought.

## Further reading

If this pattern resonates, a few other articles go deeper on related pieces:

- [Testing Kafka Workflows with Playwright and Mokapi](testing-kafka-workflows-playwright.md) for more on the Kafka side of this example
- [Testing Email Workflows with Playwright and Mokapi](testing-email-workflows-playwright.md) for the same pattern applied to email
- [Acceptance Testing with Mokapi](acceptance-testing.md) for the bigger picture on why testing the real flow matters