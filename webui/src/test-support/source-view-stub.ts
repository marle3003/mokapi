import { defineComponent, h, type PropType } from 'vue'

// Stands in for SourceView.vue, which pulls in ace. It renders the preview content as text, so a
// test compares the value. Use it as
//   vi.mock('@/components/dashboard/SourceView.vue', () => import('@/test-support/source-view-stub'))
export default defineComponent({
  props: {
    source: { type: Object as PropType<{ preview?: { content: string } }>, required: true },
  },
  setup(props) {
    return () => h('pre', props.source.preview?.content)
  },
})
