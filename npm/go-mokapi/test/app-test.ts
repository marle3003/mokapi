import { app } from "mokapi";

app.http().get('/foo', (req, res) => {})
app.http().route('').get(() =>{}).post(() => {})
app.api('title').http().get('', (req, res) => {})
app.http().api('title').use((req, res) => {})
app.http().custom('/foo', () => {})
app.http().get('', () => {}, { track: true })
// @ts-ignore
app.http().get('', () => {}, { track: 123 })
app.http().get('', () => {}, { priority: 123 })
// @ts-ignore
app.http().get('', () => {}, { priority: true })
app.http().get('', () => {}, { tags: { foo: 'bar' } })
// @ts-ignore
app.http().get('', () => {}, { tags: true })
// @ts-ignore
app.http().route(123)
// @ts-ignore
app.http().put(123, () => {})
// @ts-ignore
app.http().route(('')).put('', () => {})
app.http().use((req, res) => {
    const operationId = req.operationId
    // @ts-ignore
    const id: number = req.operationId
    res.data = {}
    res.body = ''
    // @ts-ignore
    res.body = {}
    res.stopPropagation()
    res.context.foo = 'bar'
    res.context = {}
})