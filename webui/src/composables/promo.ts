import { computed } from 'vue'
import { promotions, type Promotion } from '@/config/promo'

export function usePromo() {
  const now = new Date()

  const activePromotion = computed<Promotion | undefined>(() =>
    promotions.find(p =>
      p.enabled &&
      now >= new Date(p.validFrom) &&
      now <= new Date(p.validUntil)
    )
  )
  return {
    activePromotion,
  }
}