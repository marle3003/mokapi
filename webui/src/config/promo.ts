export interface Promotion {
  enabled: boolean
  discount: number
  validFrom: string
  validUntil: string
}

export const promotions: Promotion[] = [ 
  {
    enabled: true,
    discount: 25,
    validFrom: '2025-12-29',
    validUntil: '2026-01-07', // next date to be valid until midnight
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2026-01-13',
    validUntil: '2026-01-19',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2026-01-20',
    validUntil: '2026-01-26',
  },
  {
    enabled: true,
    discount: 30,
    validFrom: '2026-01-30',
    validUntil: '2026-02-02',
  },
]