export interface Promotion {
  enabled: boolean
  discount: number
  validFrom: string
  validUntil: string
}

export const promotions: Promotion[] = [ 
  {
    enabled: true,
    discount: 30,
    validFrom: '2025-05-01',
    validUntil: '2026-05-06', // next date to be valid until midnight
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2025-05-06',
    validUntil: '2026-05-12',
  },
  {
    enabled: true,
    discount: 20,
    validFrom: '2026-05-20',
    validUntil: '2026-05-25',
  },
  {
    enabled: true,
    discount: 25,
    validFrom: '2026-05-29',
    validUntil: '2026-06-01',
  },
]