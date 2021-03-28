<script>
import { Line, mixins } from 'vue-chartjs'
export default {
  extends: Line,
  mixins: [mixins.reactiveProp],
  computed: {
    options() {
      return {
        scales: {
          yAxes: [{
            ticks: {
              maxTicksLimit: 5,
              callback: (value) => {
                return this.prettyBytes(value)
              },
            },
            gridLines: {
              color: '#d3d4d5',
              drawBorder: false,
            }
          }],
          xAxes:[{
            type: 'time',
            time: {
              unit: 'minute',
              unitStepSize: 5,
            },
            gridLines: {
              color: '#d3d4d5',
            }
          }]
        },
        responsive: true,
        maintainAspectRatio: false,
        legend: {
          position: 'bottom'
        },
        legendCallback: function(chart) {
            return ''
        }
      }
    }
  },
  mounted () {
    this.renderChart(this.chartData, this.options)
    $('#legend').prepend(mybarChart.generateLegend());
  },
  methods: {
    prettyBytes: function (num){
      // jacked from: https://github.com/sindresorhus/pretty-bytes
      if (typeof num !== 'number' || isNaN(num)) {
        return 0
      }

      var exponent;
      var unit;
      var neg = num < 0;
      var units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

      if (neg) {
        num = -num;
      }

      if (num < 1) {
        return (neg ? '-' : '') + num + ' B';
      }

      exponent = Math.min(Math.floor(Math.log(num) / Math.log(1000)), units.length - 1);
      num = (num / Math.pow(1000, exponent)).toFixed(2) * 1;
      unit = units[exponent];

      return (neg ? '-' : '') + num + ' ' + unit;
    }
  }
}
</script>