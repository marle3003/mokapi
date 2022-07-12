<script>
export default {
  methods: {
    maxMetric (metrics, name, ...labels) {
      if (!metrics) {
        return
      }

      var value = 0
      for (var metric of metrics) {
        if (!metric.name.startsWith(name)) {
          continue
        }
        if (labels.length === 0) {
          if (metric.value > value) {
            value = metric.value
          }
        } else if (labels.length > 0 && this.matchLabels(metric, labels)) {
          if (metric.value > value) {
            value = metric.value
          }
        }
      }
      return value
    },
    metric (metrics, name, ...labels) {
      if (!metrics) {
        return
      }

      var value = 0
      for (var metric of metrics) {
        if (!metric.name.startsWith(name)) {
          continue
        }
        if (labels.length === 0) {
          value += metric.value
        } else if (labels.length > 0 && this.matchLabels(metric, labels)) {
          value += metric.value
        }
      }
      return value
    },
    matchLabels (metric, labels) {
      for (var label of labels) {
        const s = `${label.name}="${label.value}"`
        if (!metric.name.includes(s)) {
          return false
        }
      }
      return true
    }
  }
}
</script>
