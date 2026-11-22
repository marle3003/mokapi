import { expect } from '@playwright/test'

// `2026-11-16 12:00:00 (in 1h 2min 3s)` in seconds
export function countdownSeconds(text: string | null): number {
    const match = /\(in (?:(\d+)h )?(?:(\d+)min )?(\d+)s\)/.exec(text ?? '')
    expect(match, `a countdown in "${text}"`).not.toBeNull()
    return Number(match![1] ?? 0) * 3600 + Number(match![2] ?? 0) * 60 + Number(match![3])
}
