// The auto refresh runs only with a refresh interval in the route (?refresh=<seconds>).
// 0, an empty or a negative value turns it off, like a missing or non-numeric one.
export function refreshIntervalSeconds(route: { query: Record<string, unknown> }): number | undefined {
    const seconds = Number(route.query.refresh)
    return Number.isFinite(seconds) && seconds > 0 ? seconds : undefined
}
