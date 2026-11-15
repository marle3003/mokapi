import { getRouteName } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute, useRouter } from '@/router';

// What the request and the response part of a ListOffsets request share: the topic rows are
// links, and an offset timestamp is a time or one of the special values Earliest and Latest.
export function useListOffsets() {
  const router = useRouter();
  const route = useRoute();
  const { format } = usePrettyDates();

  function goToTopic(topic: string, openInNewTab = false) {
    if (getSelection()?.toString()) {
      return
    }

    const to = {
      name: getRouteName('kafkaTopic').value,
      params: { service: route.params.service, topic: topic }
    }
    if (openInNewTab) {
      const routeData = router.resolve(to);
      window.open(routeData.href, '_blank')
    } else {
      router.push(to)
    }
  }

  function formatTimestamp(value: number): string {
    switch (value) {
      case -2: return 'Earliest';
      case -1: return 'Latest';
      default: return format(value);
    }
  }

  return { goToTopic, formatTimestamp }
}
