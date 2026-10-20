// composables/breadcrumb.ts
import { type InjectionKey, type Ref, inject, provide, ref, watchEffect, onUnmounted } from 'vue'

const key: InjectionKey<Ref<BreadcrumbItem[]>> = Symbol('breadcrumb-tail')

/** Im Layout: stellt den Platz für den variablen Teil bereit */
export function provideBreadcrumb() {
  const tail = ref<BreadcrumbItem[]>([])
  provide(key, tail)
  return tail
}

/** In der Unterseite: setzt den variablen Teil */
export function useBreadcrumb(items: () => BreadcrumbItem[]) {
  const tail = inject(key)
  if (!tail) return
  watchEffect(() => { tail.value = items() })
  onUnmounted(() => { tail.value = [] })
}