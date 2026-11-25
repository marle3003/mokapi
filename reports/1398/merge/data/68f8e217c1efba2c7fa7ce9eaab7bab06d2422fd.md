# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: dashboard/job-log.spec.ts >> Job log: Pause >> a new run appears on top and the selection stays on the same run
- Location: e2e/tests/dashboard/job-log.spec.ts:736:5

# Error details

```
Error: expect(locator).toHaveCount(expected) failed

Locator:  getByRole('region', { name: 'Recent Jobs' }).getByRole('table', { name: 'Recent Jobs' }).locator('tbody tr')
Expected: 15
Received: 14
Timeout:  5000ms

Call log:
  - Expect "toHaveCount" getByRole('region', { name: 'Recent Jobs' }).getByRole('table', { name: 'Recent Jobs' }).locator('tbody tr') with timeout 5000ms
  - waiting for getByRole('region', { name: 'Recent Jobs' }).getByRole('table', { name: 'Recent Jobs' }).locator('tbody tr')
    13 × locator resolved to 14 elements
       - unexpected value "14"

```

# Page snapshot

```yaml
- generic [ref=e2]:
  - banner [ref=e3]:
    - navigation "Main" [ref=e4]:
      - generic [ref=e5]:
        - link [ref=e6] [cursor=pointer]:
          - /url: ./
          - img "Mokapi - Open-Source Multi-Protocol Mock API Tool" [ref=e7]
        - text:     
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
                - text:                                            
            - listitem [ref=e17]:
              - generic [ref=e18]:
                - generic [ref=e19]:
                  - link "Resources" [ref=e20] [cursor=pointer]:
                    - /url: https://mokapi.io/resources
                  - text:  
                - text:      
          - generic [ref=e21]:
            - button "Classic view" [ref=e22] [cursor=pointer]
            - link "v0.11.0" [ref=e23] [cursor=pointer]:
              - /url: https://github.com/marle3003/mokapi
            - button "Search" [ref=e24] [cursor=pointer]:
              - generic "Search" [ref=e25]: 
            - button "" [ref=e26] [cursor=pointer]:
              - generic "Switch to light mode" [ref=e27]: 
              - text: 
  - text: 
  - main [ref=e30]:
    - status
    - region "Dashboard" [ref=e31]:
      - heading "Dashboard" [level=1] [ref=e32]
      - generic [ref=e33]:
        - navigation "Activity bar" [ref=e34]:
          - button "Services" [pressed] [ref=e35] [cursor=pointer]:
            - generic [aria-hidden] [ref=e36]: 
          - button "Search" [ref=e37] [cursor=pointer]:
            - generic [aria-hidden] [ref=e38]: 
          - link "Jobs" [ref=e39] [cursor=pointer]:
            - /url: /dashboard/jobs
            - generic [aria-hidden] [ref=e40]: 
          - link "Configs" [ref=e41] [cursor=pointer]:
            - /url: /dashboard/configs
            - generic [aria-hidden] [ref=e42]: 
          - link "Faker" [ref=e43] [cursor=pointer]:
            - /url: /dashboard/faker
            - generic [aria-hidden] [ref=e44]: 
        - complementary "Sidebar" [ref=e45]:
          - searchbox "Filter services" [ref=e48]
          - list [ref=e49]:
            - listitem [ref=e50]:
              - link "Overview" [ref=e51] [cursor=pointer]:
                - /url: /dashboard
          - generic [ref=e52]:
            - heading "Pinned" [level=2] [ref=e53]
            - paragraph [ref=e54]: Pin a tab with its star.
          - generic [ref=e55]:
            - generic [ref=e56]:
              - button "Collapse HTTP" [expanded] [ref=e57] [cursor=pointer]:
                - generic [aria-hidden] [ref=e58]: 
              - link "HTTP" [ref=e59] [cursor=pointer]:
                - /url: /dashboard/http
            - list "HTTP services" [ref=e60]:
              - listitem [ref=e61]:
                - link "Books API 0" [ref=e62] [cursor=pointer]:
                  - /url: /dashboard/http/services/Books%20API
                  - generic [ref=e63]: Books API
                  - generic [ref=e64]: "0"
              - listitem [ref=e65]:
                - link "Swagger Petstore Service has errors 4" [ref=e66] [cursor=pointer]:
                  - /url: /dashboard/http/services/Swagger%20Petstore
                  - generic [ref=e67]: Swagger Petstore
                  - img "Service has errors" [ref=e68]: 
                  - generic [ref=e69]: "4"
          - generic [ref=e70]:
            - generic [ref=e71]:
              - button "Collapse Kafka" [expanded] [ref=e72] [cursor=pointer]:
                - generic [aria-hidden] [ref=e73]: 
              - link "Kafka" [ref=e74] [cursor=pointer]:
                - /url: /dashboard/kafka
            - list "Kafka services" [ref=e75]:
              - listitem [ref=e76]:
                - link "Kafka World 12" [ref=e77] [cursor=pointer]:
                  - /url: /dashboard/kafka/service/Kafka%20World
                  - generic [ref=e78]: Kafka World
                  - generic [ref=e79]: "12"
          - generic [ref=e80]:
            - generic [ref=e81]:
              - button "Collapse MQTT" [expanded] [ref=e82] [cursor=pointer]:
                - generic [aria-hidden] [ref=e83]: 
              - link "MQTT" [ref=e84] [cursor=pointer]:
                - /url: /dashboard/mqtt
            - list "MQTT services" [ref=e85]:
              - listitem [ref=e86]:
                - link "MQTT Temperature Sensor API 2" [ref=e87] [cursor=pointer]:
                  - /url: /dashboard/mqtt/service/MQTT%20Temperature%20Sensor%20API
                  - generic [ref=e88]: MQTT Temperature Sensor API
                  - generic [ref=e89]: "2"
          - generic [ref=e90]:
            - generic [ref=e91]:
              - button "Collapse Websocket" [expanded] [ref=e92] [cursor=pointer]:
                - generic [aria-hidden] [ref=e93]: 
              - link "Websocket" [ref=e94] [cursor=pointer]:
                - /url: /dashboard/websocket
            - list "Websocket services" [ref=e95]:
              - listitem [ref=e96]:
                - link "WebSocket Chat API 12" [ref=e97] [cursor=pointer]:
                  - /url: /dashboard/websocket/service/WebSocket%20Chat%20API
                  - generic [ref=e98]: WebSocket Chat API
                  - generic [ref=e99]: "12"
          - generic [ref=e100]:
            - generic [ref=e101]:
              - button "Collapse Mail" [expanded] [ref=e102] [cursor=pointer]:
                - generic [aria-hidden] [ref=e103]: 
              - link "Mail" [ref=e104] [cursor=pointer]:
                - /url: /dashboard/mail
            - list "Mail services" [ref=e105]:
              - listitem [ref=e106]:
                - link "Mail Second Server 4" [ref=e107] [cursor=pointer]:
                  - /url: /dashboard/mail/service/Mail%20Second%20Server
                  - generic [ref=e108]: Mail Second Server
                  - generic [ref=e109]: "4"
              - listitem [ref=e110]:
                - link "Mail Testserver 8" [ref=e111] [cursor=pointer]:
                  - /url: /dashboard/mail/service/Mail%20Testserver
                  - generic [ref=e112]: Mail Testserver
                  - generic [ref=e113]: "8"
          - generic [ref=e114]:
            - generic [ref=e115]:
              - button "Collapse LDAP" [expanded] [ref=e116] [cursor=pointer]:
                - generic [aria-hidden] [ref=e117]: 
              - link "LDAP" [ref=e118] [cursor=pointer]:
                - /url: /dashboard/ldap
            - list "LDAP services" [ref=e119]:
              - listitem [ref=e120]:
                - link "LDAP Second Server 3" [ref=e121] [cursor=pointer]:
                  - /url: /dashboard/ldap/service/LDAP%20Second%20Server
                  - generic [ref=e122]: LDAP Second Server
                  - generic [ref=e123]: "3"
              - listitem [ref=e124]:
                - link "LDAP Testserver 12" [ref=e125] [cursor=pointer]:
                  - /url: /dashboard/ldap/service/LDAP%20Testserver
                  - generic [ref=e126]: LDAP Testserver
                  - generic [ref=e127]: "12"
        - generic [ref=e128]:
          - navigation "Tabs" [ref=e129]:
            - list "Open tabs" [ref=e130]:
              - listitem [ref=e131]:
                - link "Jobs" [ref=e132] [cursor=pointer]:
                  - /url: /dashboard/jobs
                - button "Pin Jobs" [ref=e134] [cursor=pointer]:
                  - generic [aria-hidden] [ref=e135]: 
                - button "Close Jobs" [ref=e136] [cursor=pointer]:
                  - generic [aria-hidden] [ref=e137]: 
          - region "Jobs" [ref=e141]:
            - generic [ref=e142]:
              - tablist [ref=e143]:
                - tab "Jobs" [ref=e144] [cursor=pointer]
                - tab "Runs" [selected] [ref=e145] [cursor=pointer]
              - tabpanel "Runs" [ref=e147]:
                - generic [ref=e149]:
                  - region "Recent Jobs" [ref=e150]:
                    - generic [ref=e151]:
                      - heading "Recent Jobs" [level=2] [ref=e152]
                      - generic [ref=e153]:
                        - button "Add filter" [ref=e156] [cursor=pointer]:
                          - generic [aria-hidden] [ref=e157]: 
                          - text: Add filter
                        - button "Pause" [ref=e159] [cursor=pointer]:
                          - generic [aria-hidden] [ref=e160]: 
                          - text: Pause
                      - table "Recent Jobs" [ref=e162]:
                        - rowgroup [ref=e163]:
                          - row [ref=e164]:
                            - columnheader "Status" [ref=e165]
                            - columnheader "Name" [ref=e166]
                            - columnheader "Schedule" [ref=e167]
                            - columnheader "Time" [ref=e168]
                            - columnheader "Duration" [ref=e169]
                        - rowgroup [ref=e170]:
                          - row [ref=e171]:
                            - cell "Success" [ref=e172]
                            - cell [ref=e174]:
                              - link "cron" [ref=e175] [cursor=pointer]:
                                - /url: /dashboard/jobs/8234
                            - cell "3s" [ref=e176]
                            - cell "2026-11-25 09:01:48" [ref=e177]
                            - cell "2 [sec]" [ref=e178]
                          - row [selected] [ref=e179]:
                            - cell "Error" [ref=e180]
                            - cell [ref=e182]:
                              - link "cron" [ref=e183] [cursor=pointer]:
                                - /url: /dashboard/jobs/8134
                            - cell "3s" [ref=e184]
                            - cell "2026-11-25 08:01:48" [ref=e185]
                            - cell "2 [sec]" [ref=e186]
                          - row [ref=e187]:
                            - cell "Warning" [ref=e188]
                            - cell [ref=e190]:
                              - link "cron" [ref=e191] [cursor=pointer]:
                                - /url: /dashboard/jobs/81344
                            - cell "3s" [ref=e192]
                            - cell "2026-11-25 08:01:48" [ref=e193]
                            - cell "2 [sec]" [ref=e194]
                          - row [ref=e195]:
                            - cell "Success" [ref=e196]
                            - cell [ref=e198]:
                              - link "every" [ref=e199] [cursor=pointer]:
                                - /url: /dashboard/jobs/7001
                            - cell "1m" [ref=e200]
                            - cell "2025-05-01 09:00:00" [ref=e201]
                            - cell "1 [sec]" [ref=e202]
                          - row [ref=e203]:
                            - cell "Error" [ref=e204]
                            - cell [ref=e206]:
                              - link "every" [ref=e207] [cursor=pointer]:
                                - /url: /dashboard/jobs/7002
                            - cell "1m" [ref=e208]
                            - cell "2025-05-01 08:59:00" [ref=e209]
                            - cell "1 [sec]" [ref=e210]
                          - row [ref=e211]:
                            - cell "Success" [ref=e212]
                            - cell [ref=e214]:
                              - link "cron" [ref=e215] [cursor=pointer]:
                                - /url: /dashboard/jobs/7003
                            - cell "0 * * * *" [ref=e216]
                            - cell "2025-05-01 08:58:00" [ref=e217]
                            - cell "1 [sec]" [ref=e218]
                          - row [ref=e219]:
                            - cell "Success" [ref=e220]
                            - cell [ref=e222]:
                              - link "cleanup" [ref=e223] [cursor=pointer]:
                                - /url: /dashboard/jobs/7004
                            - cell "24h" [ref=e224]
                            - cell "2025-05-01 08:57:00" [ref=e225]
                            - cell "1 [sec]" [ref=e226]
                          - row [ref=e227]:
                            - cell "Success" [ref=e228]
                            - cell [ref=e230]:
                              - link "cron" [ref=e231] [cursor=pointer]:
                                - /url: /dashboard/jobs/7005
                            - cell "3s" [ref=e232]
                            - cell "2025-05-01 08:56:00" [ref=e233]
                            - cell "1 [sec]" [ref=e234]
                          - row [ref=e235]:
                            - cell "Success" [ref=e236]
                            - cell [ref=e238]:
                              - link "every" [ref=e239] [cursor=pointer]:
                                - /url: /dashboard/jobs/7006
                            - cell "1m" [ref=e240]
                            - cell "2025-05-01 08:55:00" [ref=e241]
                            - cell "1 [sec]" [ref=e242]
                          - row [ref=e243]:
                            - cell "Success" [ref=e244]
                            - cell [ref=e246]:
                              - link "cron" [ref=e247] [cursor=pointer]:
                                - /url: /dashboard/jobs/7007
                            - cell "3s" [ref=e248]
                            - cell "2025-05-01 08:54:00" [ref=e249]
                            - cell "1 [sec]" [ref=e250]
                          - row [ref=e251]:
                            - cell "Success" [ref=e252]
                            - cell [ref=e254]:
                              - link "every" [ref=e255] [cursor=pointer]:
                                - /url: /dashboard/jobs/7008
                            - cell "1m" [ref=e256]
                            - cell "2025-05-01 08:53:00" [ref=e257]
                            - cell "1 [sec]" [ref=e258]
                          - row [ref=e259]:
                            - cell "Success" [ref=e260]
                            - cell [ref=e262]:
                              - link "cron" [ref=e263] [cursor=pointer]:
                                - /url: /dashboard/jobs/7009
                            - cell "3s" [ref=e264]
                            - cell "2025-05-01 08:52:00" [ref=e265]
                            - cell "1 [sec]" [ref=e266]
                          - row [ref=e267]:
                            - cell "Success" [ref=e268]
                            - cell [ref=e270]:
                              - link "every" [ref=e271] [cursor=pointer]:
                                - /url: /dashboard/jobs/7010
                            - cell "1m" [ref=e272]
                            - cell "2025-05-01 08:51:00" [ref=e273]
                            - cell "1 [sec]" [ref=e274]
                          - row [ref=e275]:
                            - cell "Success" [ref=e276]
                            - cell [ref=e278]:
                              - link "cron" [ref=e279] [cursor=pointer]:
                                - /url: /dashboard/jobs/7011
                            - cell "3s" [ref=e280]
                            - cell "2025-05-01 08:50:00" [ref=e281]
                            - cell "1 [sec]" [ref=e282]
                      - generic [ref=e284]:
                        - generic [ref=e285]: Showing 14 of 14 requests
                        - combobox [ref=e286]:
                          - option "10"
                          - option "25" [selected]
                          - option "50"
                          - option "100"
                  - region "Job details" [ref=e287]:
                    - generic [ref=e288]:
                      - link "Open as tab" [ref=e289] [cursor=pointer]:
                        - /url: /dashboard/jobs/8134
                      - button "Close" [ref=e290] [cursor=pointer]
                    - generic [ref=e293]:
                      - generic [ref=e294]:
                        - paragraph [ref=e295]: Status
                        - paragraph [ref=e296]:
                          - generic [ref=e297]: Error
                      - generic [ref=e298]:
                        - paragraph [ref=e299]: Name
                        - paragraph [ref=e300]: cron
                      - generic [ref=e301]:
                        - paragraph [ref=e302]: Time
                        - paragraph [ref=e303]: 2026-11-25 08:01:48
                    - tablist [ref=e304]:
                      - tab "Logs" [selected] [ref=e305] [cursor=pointer]
                      - tab "Details" [ref=e306] [cursor=pointer]
                    - tabpanel [ref=e307]:
                      - alert [ref=e308]:
                        - heading "Error" [level=5] [ref=e309]
                        - paragraph [ref=e310]: There was an error
                      - list [ref=e311]:
                        - listitem [ref=e312]:
                          - generic [ref=e313]:
                            - generic [ref=e314]: 
                            - generic [ref=e315]:
                              - text: an error message, see
                              - link "https://mokapi.io" [ref=e316] [cursor=pointer]:
                                - /url: https://mokapi.io
```

# Test source

```ts
  644 |         expect(names).toEqual(['cron'])
  645 |     })
  646 | 
  647 |     test('a click on a chip edits the filter', async ({ page }) => {
  648 |         const { log } = useJobLog(page)
  649 |         await seedStorage(page, 'job-jobs-filter', { name: 'every' })
  650 |         await page.goto(runsPath)
  651 |         await expect(log.rows).toHaveCount(5)
  652 | 
  653 |         await log.chip('Name: every').edit.click()
  654 |         await page.getByRole('dialog', { name: 'Name', exact: true }).locator('input').fill('cleanup')
  655 |         await page.keyboard.press('Escape')
  656 | 
  657 |         await expect(log.chips).toHaveText(['Name: cleanup'])
  658 |         await expect(log.rows).toHaveCount(1)
  659 |     })
  660 | 
  661 |     test('the x on a chip removes the filter', async ({ page }) => {
  662 |         const { log } = useJobLog(page)
  663 |         await seedStorage(page, 'job-jobs-filter', { name: 'every' })
  664 |         await page.goto(runsPath)
  665 |         await expect(log.rows).toHaveCount(5)
  666 | 
  667 |         await log.chip('Name: every').remove.click()
  668 | 
  669 |         await expect(log.chips).toHaveCount(0)
  670 |         await expect(log.rows).toHaveCount(14)
  671 |     })
  672 | 
  673 |     test('the filter is saved and survives a reload', async ({ page }) => {
  674 |         const { log } = useJobLog(page)
  675 |         await page.goto(runsPath)
  676 |         const popover = await log.openAddFilter('Name')
  677 |         await popover.locator('input').fill('every')
  678 |         await page.keyboard.press('Escape')
  679 |         await expect(log.rows).toHaveCount(5)
  680 | 
  681 |         await page.reload()
  682 | 
  683 |         await expect(log.chips).toHaveText(['Name: every'])
  684 |         await expect(log.rows).toHaveCount(5)
  685 |         expect(await page.evaluate(() => localStorage.getItem('job-jobs-filter'))).toBe(JSON.stringify({ name: 'every' }))
  686 |     })
  687 | 
  688 |     test('the Name chip is the only filter on offer', async ({ page }) => {
  689 |         const { log } = useJobLog(page)
  690 |         await page.goto(runsPath)
  691 | 
  692 |         await log.addFilter.click()
  693 | 
  694 |         await expect(log.addFilterMenu.getByRole('menuitem')).toHaveText(['Name'])
  695 |     })
  696 | })
  697 | 
  698 | const arrivingJob = {
  699 |     id: '9001',
  700 |     traits: { namespace: 'job', name: 'cron' },
  701 |     time: new Date().toISOString(),
  702 |     data: { schedule: '3s', maxRuns: -1, runs: 11, nextRun: '2025-05-01T09:00:00Z', duration: 1000, logs: [], ...cron },
  703 | }
  704 | 
  705 | async function serveArrivingJob(page: Page) {
  706 |     let arrived = false
  707 |     let loads = 0
  708 |     await page.route('**/api/events?*', async route => {
  709 |         loads++
  710 |         const response = await route.fetch()
  711 |         if (!arrived) {
  712 |             return route.fulfill({ response })
  713 |         }
  714 |         const body = await response.json()
  715 |         return route.fulfill({ response, json: { ...body, items: [arrivingJob, ...body.items], total: body.total + 1 } })
  716 |     })
  717 |     return {
  718 |         arrive: () => { arrived = true },
  719 |         loads: () => loads,
  720 |     }
  721 | }
  722 | 
  723 | test.describe('Job log: Pause', () => {
  724 |     // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
  725 |     // so only runFor() lets a reload happen.
  726 |     async function openRefreshingLog(page: Page) {
  727 |         const { log } = useJobLog(page)
  728 |         const events = await serveArrivingJob(page)
  729 |         await page.clock.install()
  730 |         await page.goto(`${jobsPath}?refresh=1#tab-runs`)
  731 |         await expect(log.rows).toHaveCount(14)
  732 |         await pauseClock(page)
  733 |         return { log, events }
  734 |     }
  735 | 
  736 |     test('a new run appears on top and the selection stays on the same run', async ({ page }) => {
  737 |         const { log, events } = await openRefreshingLog(page)
  738 |         await log.time(log.row(errorRun)).click()
  739 |         await expect(log.selectedRow).toHaveCount(1)
  740 | 
  741 |         events.arrive()
  742 |         await page.clock.runFor(1000)
  743 | 
> 744 |         await expect(log.rows).toHaveCount(15)
      |                                ^ Error: expect(locator).toHaveCount(expected) failed
  745 |         await expect(log.rows.first().locator(`a[href$="/jobs/${arrivingJob.id}"]`)).toHaveCount(1)
  746 |         await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)
  747 |         await expect(page).toHaveURL(new RegExp(`job=${errorRun}`))
  748 |     })
  749 | 
  750 |     test('Pause stops reloading and shows Paused', async ({ page }) => {
  751 |         const { log, events } = await openRefreshingLog(page)
  752 |         await expect(log.paused).toHaveCount(0)
  753 | 
  754 |         await log.pause.click()
  755 | 
  756 |         await expect(log.paused).toBeVisible()
  757 |         await expect(log.resume).toBeVisible()
  758 |         await expect(log.pause).toHaveCount(0)
  759 |         const before = events.loads()
  760 |         events.arrive()
  761 |         await page.clock.runFor(3000)
  762 |         expect(events.loads()).toBe(before)
  763 |         await expect(log.rows).toHaveCount(14)
  764 |     })
  765 | 
  766 |     test('Resume loads at once and keeps the selection', async ({ page }) => {
  767 |         const { log, events } = await openRefreshingLog(page)
  768 |         await log.time(log.row(errorRun)).click()
  769 |         await log.pause.click()
  770 |         await expect(log.paused).toBeVisible()
  771 |         events.arrive()
  772 |         await page.clock.runFor(3000)
  773 |         await expect(log.rows).toHaveCount(14)
  774 | 
  775 |         await log.resume.click()
  776 | 
  777 |         await expect(log.rows).toHaveCount(15)
  778 |         await expect(log.paused).toHaveCount(0)
  779 |         await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)
  780 |     })
  781 | 
  782 |     test('the countdown keeps running while the log is paused', async ({ page }) => {
  783 |         const { log } = await openRefreshingLog(page)
  784 |         await log.time(log.row(plain)).click()
  785 |         await expect(log.panel.tab('Details')).toBeVisible()
  786 |         await moveClockPastNewHandlers(page)
  787 |         await log.panel.tab('Details').click()
  788 |         const nextRun = log.panel.element.locator('p.label', { hasText: /^Next Run$/ }).locator('xpath=following-sibling::*[1]')
  789 |         await expect(nextRun).toContainText('(in ')
  790 |         await log.pause.click()
  791 |         await expect(log.paused).toBeVisible()
  792 |         const before = countdownSeconds(await nextRun.textContent())
  793 | 
  794 |         await page.clock.runFor(3000)
  795 | 
  796 |         await expect.poll(async () => countdownSeconds(await nextRun.textContent())).toBe(before - 3)
  797 |     })
  798 | 
  799 |     test('no Pause without ?refresh', async ({ page }) => {
  800 |         const { log } = useJobLog(page)
  801 |         await page.goto(runsPath)
  802 |         await expect(log.rows).toHaveCount(14)
  803 |         await expect(log.pause).toHaveCount(0)
  804 |         await expect(log.paused).toHaveCount(0)
  805 |     })
  806 | 
  807 |     test('the pause is not kept after a reload', async ({ page }) => {
  808 |         const { log } = useJobLog(page)
  809 |         await page.goto(`${jobsPath}?refresh=1#tab-runs`)
  810 |         await log.pause.click()
  811 |         await expect(log.paused).toBeVisible()
  812 | 
  813 |         await page.reload()
  814 | 
  815 |         await expect(log.rows).toHaveCount(14)
  816 |         await expect(log.paused).toHaveCount(0)
  817 |         await expect(log.pause).toBeVisible()
  818 |     })
  819 | })
  820 | 
  821 | // The Classic layout is the default of the fixture.
  822 | test.describe('Job log: Classic layout', () => {
  823 |     test.use({ layout: 'classic' })
  824 | 
  825 |     test('a run can be selected', async ({ page }) => {
  826 |         const { log, view } = useJobLog(page)
  827 |         await page.goto(runsPath)
  828 |         await expect(log.rows).toHaveCount(14)
  829 | 
  830 |         await log.time(log.row(errorRun)).click()
  831 | 
  832 |         await expect(log.panel.element).toBeVisible()
  833 |         await expect(view.field(log.panel.element, 'Name')).toHaveText('cron')
  834 |         await expect(log.selectedRow).toHaveCount(1)
  835 |     })
  836 | 
  837 |     test('the log can be paused and resumed', async ({ page }) => {
  838 |         const { log } = useJobLog(page)
  839 |         await page.goto(`${jobsPath}?refresh=1#tab-runs`)
  840 |         await expect(log.rows).toHaveCount(14)
  841 | 
  842 |         await log.pause.click()
  843 |         await expect(log.paused).toBeVisible()
  844 |         await log.resume.click()
```