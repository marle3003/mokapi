export interface Promotion {
  enabled: boolean
  discount: number
  validFrom: string
  validUntil: string // next date to be valid until midnight
}

export const promotions: Promotion[] = [ 
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-11-11',
    validUntil: '2026-11-16', 
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-11-18',
    validUntil: '2026-11-26',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-11-27',
    validUntil: '2026-11-31',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-12-01',
    validUntil: '2026-12-07',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-12-09',
    validUntil: '2026-12-17',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-12-26',
    validUntil: '2027-01-01',
  },
]