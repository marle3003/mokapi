# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: dashboard-demo/ldap.spec.ts >> Visit LDAP Testserver
- Location: e2e/tests/dashboard-demo/ldap.spec.ts:8:5

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for getByRole('cell').getByText('HR Employee Directory')

```

# Page snapshot

```yaml
- generic [ref=e2]:
  - banner [ref=e3]:
    - navigation "Main" [ref=e4]:
      - generic [ref=e5]:
        - link "Mokapi home" [ref=e6] [cursor=pointer]:
          - /url: ./
          - img "Mokapi home" [ref=e7]
        - text:    
        - generic [ref=e8]:
          - list [ref=e10]:
            - listitem [ref=e11]:
              - link "Dashboard" [ref=e12] [cursor=pointer]:
                - /url: /dashboard?refresh=20
            - listitem [ref=e13]:
              - generic [ref=e14]:
                - generic [ref=e15]:
                  - link "Docs" [ref=e16] [cursor=pointer]:
                    - /url: /docs
                  - text:  
                - text:                                
            - listitem [ref=e17]:
              - generic [ref=e18]:
                - generic [ref=e19]:
                  - link "Resources" [ref=e20] [cursor=pointer]:
                    - /url: /resources
                  - text:  
                - text:      
          - generic [ref=e21]:
            - link "v0.11.0" [ref=e22] [cursor=pointer]:
              - /url: https://github.com/marle3003/mokapi
            - button "Search" [ref=e23] [cursor=pointer]:
              - generic "Search" [ref=e24]: 
            - button "" [ref=e25] [cursor=pointer]:
              - generic "Switch to dark mode" [ref=e26]: 
  - text: 
  - main [ref=e27]:
    - generic [ref=e31]:
      - generic [ref=e32]:
        - img [ref=e35]
        - generic [ref=e37]: Page not found
      - button "Back to Home" [ref=e40] [cursor=pointer]
```

# Test source

```ts
  1   | import { test, expect } from '../models/fixture-website'
  2   | import { getCellByColumnName } from '../helpers/table'
  3   | 
  4   | test.use({ colorScheme: 'light' })
  5   | // reset storage state
  6   | test.use({ storageState: { cookies: [], origins: [] } });
  7   | 
  8   | test('Visit LDAP Testserver', async ({ page }) => {
  9   | 
  10  |     await page.goto('/dashboard-demo');
> 11  |     await page.getByRole('cell').getByText('HR Employee Directory').click();
      |                                                                     ^ Error: locator.click: Test timeout of 30000ms exceeded.
  12  | 
  13  |     await test.step('Verify service info', async () => {
  14  | 
  15  |         const info = page.getByRole('region', { name: 'Info' })
  16  |         await expect(info.getByLabel('Name')).toHaveText('HR Employee Directory');
  17  |         await expect(info.getByLabel('Version')).toHaveText('1.0.0');
  18  |         await expect(info.getByLabel('Contact')).not.toBeVisible();
  19  |         await expect(info.getByLabel('Type of API')).toHaveText('LDAP');
  20  |         await expect(info.getByLabel('Description')).toHaveText('LDAP server for internal employee contact information.');
  21  | 
  22  |     });
  23  | 
  24  |     await test.step('Verify Servers', async () => {
  25  | 
  26  |         const region = page.getByRole('region', { name: 'Servers' });
  27  |         const table = region.getByRole('table', { name: 'Servers' });
  28  |         const rows = table.locator('tbody tr');
  29  |         await expect(rows).toHaveCount(1);
  30  | 
  31  |         await expect(await getCellByColumnName(table, 'Address', rows.nth(0))).toHaveText(':8389');
  32  | 
  33  |     });
  34  | 
  35  |     await test.step('Verify Configs', async () => {
  36  | 
  37  |         const table = page.getByRole('table', { name: 'Configs' });
  38  |         await expect(await getCellByColumnName(table, 'URL')).toContainText('/ldap.yaml');
  39  |         await expect(await getCellByColumnName(table, 'Provider')).toHaveText('File');
  40  | 
  41  |     });
  42  | 
  43  |     await test.step('Verify Recent Requests', async () => {
  44  |         const table = page.getByRole('table', { name: 'Recent Requests' });
  45  |         let rows = table.locator('tbody tr');
  46  | 
  47  |         // Unbind
  48  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(0))).toHaveText('Unbind');
  49  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(0))).toBeEmpty();
  50  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(0))).toBeEmpty();
  51  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(0))).toHaveText('-');
  52  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(0))).not.toBeEmpty();
  53  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(0))).not.toBeEmpty();
  54  | 
  55  |         // Search useAccountControl
  56  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(1))).toHaveText('Search');
  57  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(1))).toHaveText('dc=hr,dc=example,dc=com');
  58  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(1))).toHaveText('(userAccountControl:1.2.840.113556.1.4.803:=512)');
  59  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(1))).toHaveText('Success');
  60  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(1))).not.toBeEmpty();
  61  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(1))).not.toBeEmpty();
  62  | 
  63  |         // Delete
  64  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(2))).toHaveText('Delete');
  65  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(2))).toHaveText('uid=ctaylor,ou=people,dc=hr,dc=example,dc=com');
  66  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(2))).toHaveText('');
  67  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(2))).toHaveText('Success');
  68  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(2))).not.toBeEmpty();
  69  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(2))).not.toBeEmpty();
  70  | 
  71  |         // ModifyDN
  72  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(3))).toHaveText('ModifyDN');
  73  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(3))).toHaveText('uid=cbrown,ou=people,dc=hr,dc=example,dc=com');
  74  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(3))).toHaveText('uid=ctaylor');
  75  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(3))).toHaveText('Success');
  76  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(3))).not.toBeEmpty();
  77  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(3))).not.toBeEmpty();
  78  | 
  79  |         // Compare
  80  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(4))).toHaveText('Compare');
  81  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(4))).toHaveText('uid=bmiller,ou=people,dc=hr,dc=example,dc=com');
  82  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(4))).toHaveText('telephoneNumber == +1 555 123 9876');
  83  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(4))).toHaveText('CompareTrue');
  84  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(4))).not.toBeEmpty();
  85  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(4))).not.toBeEmpty();
  86  | 
  87  |         // Modify
  88  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(5))).toHaveText('Modify');
  89  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(5))).toHaveText('uid=bmiller,ou=people,dc=hr,dc=example,dc=com');
  90  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(5))).toHaveText('add telephoneNumber');
  91  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(5))).toHaveText('Success');
  92  |         await expect(await getCellByColumnName(table, 'Time', rows.nth(5))).not.toBeEmpty();
  93  |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(5))).not.toBeEmpty();
  94  | 
  95  |         // Add
  96  |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(6))).toHaveText('Add');
  97  |         await expect(await getCellByColumnName(table, 'DN', rows.nth(6))).toHaveText('uid=cbrown,ou=people,dc=hr,dc=example,dc=com');
  98  |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(6))).toBeEmpty();
  99  |         await expect(await getCellByColumnName(table, 'Status', rows.nth(6))).toHaveText('Success');
  100 |         await expect(await getCellByColumnName(table, 'Time', rows.nth(6))).not.toBeEmpty();
  101 |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(6))).not.toBeEmpty();
  102 | 
  103 |         // Search memberOf
  104 |         await expect(await getCellByColumnName(table, 'Operation', rows.nth(7))).toHaveText('Search');
  105 |         await expect(await getCellByColumnName(table, 'DN', rows.nth(7))).toHaveText('dc=hr,dc=example,dc=com');
  106 |         await expect(await getCellByColumnName(table, 'Criteria', rows.nth(7))).toHaveText('(memberOf=cn=Sales,ou=departments,dc=hr,dc=example,dc=com)');
  107 |         await expect(await getCellByColumnName(table, 'Status', rows.nth(7))).toHaveText('Success');
  108 |         await expect(await getCellByColumnName(table, 'Time', rows.nth(7))).not.toBeEmpty();
  109 |         await expect(await getCellByColumnName(table, 'Duration', rows.nth(7))).not.toBeEmpty();
  110 | 
  111 |         // Search uid
```