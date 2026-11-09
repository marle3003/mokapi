import { computed, onMounted, ref } from 'vue'
import { loadMarkdown } from '@/composables/docs-content'
import { useMarkdown } from '@/composables/markdown'

// The rendered release notes, empty until docs/release.md is loaded (or if it does not exist).
export function useReleaseNotes() {
  const markdown = ref<string>()
  onMounted(async () => {
    markdown.value = await loadMarkdown('release.md')
  })
  return computed(() => useMarkdown(markdown.value).content)
}
