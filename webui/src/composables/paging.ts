import { ref, watch } from 'vue'

export function usePageSize(defaultSize = 25) {
  const STORAGE_KEY = 'page_size'

  // Read initial value from local storage, or fall back to default
  const savedSize = localStorage.getItem(STORAGE_KEY)
  const pageSize = ref(savedSize ? parseInt(savedSize, 10) : defaultSize)

  // Watch for changes and save them automatically to local storage
  watch(pageSize, (newSize) => {
    localStorage.setItem(STORAGE_KEY, newSize.toString())
  })

  return pageSize
}
