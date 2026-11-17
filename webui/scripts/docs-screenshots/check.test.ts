// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { compareShots, findImages, findPlaceholders } from './check.ts'

describe('findPlaceholders', () => {
  it('returns each placeholder with file and line', () => {
    const markdown = [
      '# Dashboard',
      '<!-- screenshot: request log with chips -->',
      'text',
      '',
      '<!--screenshot:   panel on the Value sub tab-->',
    ].join('\n')

    expect(findPlaceholders(markdown, 'http/dashboard.md')).toEqual([
      { file: 'http/dashboard.md', line: 2, description: 'request log with chips' },
      { file: 'http/dashboard.md', line: 5, description: 'panel on the Value sub tab' },
    ])
  })

  it('returns nothing without placeholders and ignores other comments', () => {
    expect(findPlaceholders('# Title\n<!-- TODO -->\ntext', 'a.md')).toEqual([])
  })
})

describe('findImages', () => {
  it('returns docs images with the /docs/ prefix stripped', () => {
    const markdown = [
      '<img src="/docs/http/dashboard.png" alt="x" />',
      'text <img width="700" src="/docs/get-started/petstore.jpg" title="" /> more',
    ].join('\n')

    expect(findImages(markdown, 'get-started/running.md')).toEqual([
      { file: 'get-started/running.md', line: 1, src: 'http/dashboard.png' },
      { file: 'get-started/running.md', line: 2, src: 'get-started/petstore.jpg' },
    ])
  })

  it('ignores images that are not below /docs/', () => {
    const markdown = '<img src="https://example.com/a.png" />\n<img src="/website/b.png" />'

    expect(findImages(markdown, 'a.md')).toEqual([])
  })
})

describe('compareShots', () => {
  const page = 'http/dashboard.md'
  const shot = { name: 'http-log', file: 'http/dashboard-request-log.png' }

  it('reports nothing for a consistent set', () => {
    const images = [{ file: page, line: 3, src: shot.file }]

    expect(compareShots([shot], [], images)).toEqual({
      placeholders: [],
      imagesWithoutShot: [],
      shotsWithoutReference: [],
    })
  })

  it('reports remaining placeholders', () => {
    const placeholder = { file: page, line: 4, description: 'log' }

    expect(compareShots([], [placeholder], []).placeholders).toEqual([placeholder])
  })

  it('reports an image of a dashboard page that has no shot', () => {
    const old = { file: page, line: 26, src: 'http/dashboard-mock-rest-api.jpg' }

    expect(compareShots([], [], [old]).imagesWithoutShot).toEqual([old])
  })

  it('ignores images of other pages', () => {
    const logo = { file: 'http/overview.md', line: 1, src: 'http/logo.png' }

    expect(compareShots([], [], [logo]).imagesWithoutShot).toEqual([])
  })

  it('reports a shot that no page references', () => {
    expect(compareShots([shot], [], []).shotsWithoutReference).toEqual([shot])
  })

  it('does not report a website shot without reference in docs', () => {
    const home = { name: 'faker-home', file: '../webui/public/website/home.png', website: true }

    expect(compareShots([home], [], []).shotsWithoutReference).toEqual([])
  })
})
