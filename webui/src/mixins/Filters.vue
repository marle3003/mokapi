<script>
import moment from 'moment'
import http from 'http-status-codes'

export default {
  filters: {
    moment: function (value) {
      if (value === 0 || value === '') {
        return '-'
      }

      if (typeof value === 'string') {
        return moment(value).format('YYYY-MM-DD HH:mm:ss')
      }

      return moment
        .unix(value)
        .local()
        .format('YYYY-MM-DD HH:mm:ss')
    },
    fromNow: function (value) {
      if (value === 0) {
        return '-'
      }
      return moment.unix(value).fromNow(true)
    },
    duration: function (time) {
      let d = moment.duration(time)
      if (d.seconds() < 1) {
        return d.milliseconds() + ' [ms]'
      } else if (d.minutes() < 1) {
        return d.seconds() + ' [sec]'
      }
      return moment.duration(d).minutes()
    },
    prettyBytes: function (num) {
      // jacked from: https://github.com/sindresorhus/pretty-bytes
      if (typeof num !== 'number' || isNaN(num)) {
        return 0
      }

      let exponent
      let unit
      let neg = num < 0
      let units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

      if (neg) {
        num = -num
      }

      if (num < 1) {
        return (neg ? '-' : '') + num + ' B'
      }

      exponent = Math.min(
        Math.floor(Math.log(num) / Math.log(1000)),
        units.length - 1
      )
      num = (num / Math.pow(1000, exponent)).toFixed(2) * 1
      unit = units[exponent]

      return (neg ? '-' : '') + num + ' ' + unit
    },
    httpStatusText: function (status) {
      return http.getStatusText(status)
    }
  }
}
</script>
