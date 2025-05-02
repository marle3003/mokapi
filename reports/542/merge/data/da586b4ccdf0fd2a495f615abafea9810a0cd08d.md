# Test info

- Name: Visit Kafka topic mokapi.shop.products
- Location: /home/runner/work/mokapi/mokapi/webui/e2e/Dashboard/kafka/topic.order.spec.ts:10:5

# Error details

```
Error: Timed out 5000ms waiting for expect(locator).toHaveText(expected)

Locator: getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' }).getByRole('region', { name: 'Source' }).getByRole('region', { name: 'content' }).locator('.ace_content')
Expected pattern: /"features"/
Received string:  "{  \"category\": \"Sports\",  \"description\": \"Mob bowl what one example generosity whatever in philosophy today into every half for sharply lastly each usually week soften who e.g. be I.\",  \"id\": \"e629a4ad-8b8b-464d-b3ad-8f5324845b78\",  \"keywords\": \"hxvlQk\",  \"name\": \"Amity\",  \"price\": 537295.88,  \"subcategory\": \"wphgOtXftiVv\",  \"url\": \"https://www.directseize.info/facilitate\"}"
Call log:
  - expect.toHaveText with timeout 5000ms
  - waiting for getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' }).getByRole('region', { name: 'Source' }).getByRole('region', { name: 'content' }).locator('.ace_content')
    9 × locator resolved to <div class="ace_content">…</div>
      - unexpected value "{  "category": "Sports",  "description": "Mob bowl what one example generosity whatever in philosophy today into every half for sharply lastly each usually week soften who e.g. be I.",  "id": "e629a4ad-8b8b-464d-b3ad-8f5324845b78",  "keywords": "hxvlQk",  "name": "Amity",  "price": 537295.88,  "subcategory": "wphgOtXftiVv",  "url": "https://www.directseize.info/facilitate"}"

    at test (/home/runner/work/mokapi/mokapi/webui/e2e/components/source.ts:25:99)
    at /home/runner/work/mokapi/mokapi/webui/e2e/Dashboard/kafka/topic.order.spec.ts:98:13
    at /home/runner/work/mokapi/mokapi/webui/e2e/Dashboard/kafka/topic.order.spec.ts:93:9
    at /home/runner/work/mokapi/mokapi/webui/e2e/Dashboard/kafka/topic.order.spec.ts:55:5
```

# Page snapshot

```yaml
- banner:
  - navigation:
    - link "Mokapi home":
      - /url: ./
      - img "Mokapi home"
    - list:
      - listitem:
        - link "Dashboard":
          - /url: /dashboard?refresh=20
      - listitem:
        - button "Services"
      - listitem:
        - link "Guides":
          - /url: /docs/guides
      - listitem:
        - link "Configuration":
          - /url: /docs/configuration
      - listitem:
        - link "JavaScript API":
          - /url: /docs/javascript-api
      - listitem:
        - link "Resources":
          - /url: /docs/resources
      - listitem:
        - link "References":
          - /url: /docs/references
    - link "Version 0.11.0":
      - /url: https://github.com/marle3003/mokapi
    - text: 
- main:
  - heading "Dashboard" [level=1]
  - navigation "Services":
    - list:
      - listitem:
        - link "Overview":
          - /url: /dashboard
      - listitem:
        - link "HTTP":
          - /url: /dashboard/http
      - listitem:
        - link "Kafka":
          - /url: /dashboard/kafka
      - listitem:
        - link "SMTP":
          - /url: /dashboard/smtp
      - listitem:
        - link "LDAP":
          - /url: /dashboard/ldap
      - listitem:
        - link "Jobs":
          - /url: /dashboard/jobs
      - listitem:
        - link "Configs":
          - /url: /dashboard/configs
      - listitem:
        - link "Faker":
          - /url: /dashboard/tree
  - region "Info":
    - paragraph: Topic
    - paragraph: mokapi.shop.products
    - paragraph: Cluster
    - paragraph:
      - link "Cluster":
        - /url: /dashboard/kafka/service/Kafka%20World
        - text: Kafka World
    - text: Kafka
    - paragraph: Description
    - paragraph: Though literature second anywhere fortnightly am this either so me.
  - region "Topic Data":
    - tablist:
      - tab "Messages"
      - tab "Partitions"
      - tab "Groups"
      - tab "Configs" [selected]
    - tabpanel "Configs":
      - paragraph: Name
      - paragraph: shopOrder
      - paragraph: Title
      - paragraph: Shop New Order notification
      - paragraph: Summary
      - paragraph: A message containing details of a new order
      - paragraph: Description
      - paragraph:
        - text: More info about how the order notifications are
        - strong: created
        - text: and
        - strong: used
        - text: .
      - paragraph: Content Type
      - paragraph: application/json
      - tablist:
        - tab "Key"
        - tab "| Value" [selected]
      - tabpanel "| Value":
        - region "Schema":
          - region "Source":
            - text: 32 lines · 464 B
            - button "Copy raw content": 
            - button "Download raw content": 
            - region "Content":
              - textbox "Cursor at row 1"
              - text: "{ \"properties\": { \"category\": { \"type\": \"string\" }, \"description\": { \"type\": \"string\" }, \"features\": { \"type\": \"string\" }, \"id\": { \"type\": \"string\" }, \"keywords\": { \"type\": \"string\" }, \"name\": { \"type\": \"string\" }, \"price\": { \"type\": \"number\" }, \"subcategory\": { \"type\": \"string\" }, \"url\": { \"type\": \"string\""
          - button "Expand"
          - button "Example & Validate"
          - dialog "Value Validator - mokapi.shop.products":
            - heading "Value Validator - mokapi.shop.products" [level=6]
            - button "Close"
            - region "Source":
              - text: application/json · 10 lines · 385 B
              - button "Copy raw content": 
              - button "Download raw content": 
              - region "Content":
                - textbox "Cursor at row 10"
                - text: "{ \"category\": \"Sports\", \"description\": \"Mob bowl what one example generosity whatever in philosophy today into every half for sharply lastly each usually week soften who e.g. be I.\", \"id\": \"e629a4ad-8b8b-464d-b3ad-8f5324845b78\", \"keywords\": \"hxvlQk\", \"name\": \"Amity\", \"price\": 537295.88, \"subcategory\": \"wphgOtXftiVv\", \"url\": \"https://www.directseize.info/facilitate\" }"
            - status
            - button "Example"
            - button "Validate"
```

# Test source

```ts
   1 | import { test, Locator, expect } from "playwright/test"
   2 |
   3 | export interface Source {
   4 |     lines?: ExpectedString
   5 |     size?: ExpectedString
   6 |     content: ExpectedString
   7 |     filename: ExpectedString
   8 |     clipboard: ExpectedString 
   9 | }
  10 |
  11 | export function useSourceView(locator: Locator) {
  12 |     return {
  13 |         async test(expected: Source) {
  14 |             const source = locator.getByRole('region', { name: 'Source' })
  15 |             if (expected.lines) {
  16 |                 await expect(source.getByLabel('Lines of Code')).toHaveText(expected.lines)
  17 |             } else {
  18 |                 await expect(source.getByLabel('Lines of Code')).not.toBeVisible()
  19 |             }
  20 |             if (expected.size) {
  21 |                 await expect(source.getByLabel('Size of Code')).toHaveText(expected.size)
  22 |             } else {
  23 |                 await expect(source.getByLabel('Size of Code')).not.toBeVisible()
  24 |             }
> 25 |             await expect(source.getByRole('region', { name: 'content' }).locator('.ace_content')).toHaveText(expected.content)
     |                                                                                                   ^ Error: Timed out 5000ms waiting for expect(locator).toHaveText(expected)
  26 |
  27 |             await source.getByRole('button', { name: 'Copy raw content' }).click()
  28 |             let clipboardText = await locator.page().evaluate('navigator.clipboard.readText()')
  29 |             if (typeof expected.clipboard === 'string') {
  30 |                 await expect(clipboardText).toContain(expected.clipboard)
  31 |             }else if (expected.clipboard instanceof RegExp) {
  32 |                 await expect(clipboardText).toMatch(expected.clipboard)
  33 |             }
  34 |
  35 |             await test.step('Check download', async () => {
  36 |                 const [ download ] = await Promise.all([
  37 |                     locator.page().waitForEvent('download'),
  38 |                     source.getByRole('button', { name: 'Download raw content' }).click()
  39 |                 ])
  40 |                 await expect(download.suggestedFilename()).toBe(expected.filename)
  41 |             })
  42 |         }
  43 |     }
  44 | }
```

# Local changes

```diff
diff --git a/webui/package-lock.json b/webui/package-lock.json
index 788ec5e4..1f5bed7d 100644
--- a/webui/package-lock.json
+++ b/webui/package-lock.json
@@ -38,7 +38,7 @@
         "@vue/eslint-config-typescript": "^14.5.0",
         "@vue/tsconfig": "^0.7.0",
         "eslint": "^9.25.1",
-        "eslint-plugin-vue": "^10.0.0",
+        "eslint-plugin-vue": "^10.1.0",
         "npm-run-all": "^4.1.5",
         "prettier": "^3.5.3",
         "typescript": "~5.8.3",
@@ -2657,9 +2657,9 @@
       }
     },
     "node_modules/eslint-plugin-vue": {
-      "version": "10.0.0",
-      "resolved": "https://registry.npmjs.org/eslint-plugin-vue/-/eslint-plugin-vue-10.0.0.tgz",
-      "integrity": "sha512-XKckedtajqwmaX6u1VnECmZ6xJt+YvlmMzBPZd+/sI3ub2lpYZyFnsyWo7c3nMOQKJQudeyk1lw/JxdgeKT64w==",
+      "version": "10.1.0",
+      "resolved": "https://registry.npmjs.org/eslint-plugin-vue/-/eslint-plugin-vue-10.1.0.tgz",
+      "integrity": "sha512-/VTiJ1eSfNLw6lvG9ENySbGmcVvz6wZ9nA7ZqXlLBY2RkaF15iViYKxglWiIch12KiLAj0j1iXPYU6W4wTROFA==",
       "dev": true,
       "license": "MIT",
       "dependencies": {
diff --git a/webui/package.json b/webui/package.json
index ec8ef123..68fc36a8 100644
--- a/webui/package.json
+++ b/webui/package.json
@@ -46,7 +46,7 @@
     "@vue/eslint-config-typescript": "^14.5.0",
     "@vue/tsconfig": "^0.7.0",
     "eslint": "^9.25.1",
-    "eslint-plugin-vue": "^10.0.0",
+    "eslint-plugin-vue": "^10.1.0",
     "npm-run-all": "^4.1.5",
     "prettier": "^3.5.3",
     "typescript": "~5.8.3",
```