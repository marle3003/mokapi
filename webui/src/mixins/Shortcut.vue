<script>
export default {
  created () {
    window.addEventListener('keyup', this.shortcut)
  },
  destroyed () {
    window.removeEventListener('keyup', this.shortcut)
  },
  methods: {
    shortcut (e) {}
  }
}
</script>
