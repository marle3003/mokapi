<template>
  <div v-if="workflows !== null">
    <p class="label">Workflows</p>
    <b-table small hover class="dataTable" :items="workflows" :fields="fields">
      <template v-slot:cell(show_details)="row">
        <div @click="toggleDetails(row)" v-if="row.item.logs !== null && row.item.logs.length > 0">
          <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
          <b-icon v-else icon="plus-square"></b-icon>
        </div>
      </template>
      <template v-slot:cell(duration)="data">
        {{ data.item.duration | duration }}
      </template>
      <template v-slot:row-details="row">
        <div class="logs">
          <div v-for="line in row.item.logs" :key="line" class="line">
            {{ line }}
          </div>
        </div>
      </template>
    </b-table>
  </div>
</template>

<script>

import moment from 'moment'

export default {
  name: 'Workflows',
  props: ['workflows'],
  data () {
    return {
      fields: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', 'duration'],
      detailsShown: []
    }
  },
  methods: {
    toggleDetails (row) {
      row.toggleDetails()
      const index = this.detailsShown.indexOf(row.item.key)

      if (row.item._showDetails) {
        this.detailsShown.push(row.item.key)
      } else {
        this.detailsShown.splice(index, 1)
      }
    },
    parseLog (logs) {
      let result = []
      let group = {}
      let inGroup = false
      let line = 0
      for (let log of logs) {
        line++
        if (log.startsWith('##[group]')) {
          inGroup = true
          group = { logs: [], text: log.substr(9), line: line }
          result.push(group)
        } else if (log === '##[endgroup]') {
          line--
          inGroup = false
        } else if (inGroup) {
          group.logs.push({text: log, line: line})
        } else {
          result.push({text: log, line: line})
        }
      }

      return result
    }
  },
  filters: {
    duration: function (time) {
      let ms = Math.round(time / 1000000)
      let d = moment.duration(ms)
      if (d.seconds() < 1) {
        return d.milliseconds() + ' [ms]'
      } else if (d.minutes() < 1) {
        return d.seconds() + ' [sec]'
      }
      return moment.duration(d).minutes()
    }
  }
}
</script>

<style scoped>
.step{
  line-height: 2rem;
  border-radius: 6px;
  padding-left: 8px;
}
.step.not-collapsed{
  background-color: var(--var-bg-color-secondary);
}
.step svg {
  -moz-transition: all .3s linear;
  -webkit-transition: all .3s linear;
  transition: all .3s linear;
}
.step.not-collapsed svg {
  -moz-transform:rotate(90deg);
  -webkit-transform:rotate(90deg);
  transform:rotate(90deg);
}
.logs .line {
  font-size: 0.65rem;
  display: flex;
  padding-bottom: 3px;
}
.logs .line .number {
  padding-right: 8px;
  color: var(--var-color-line-num);
}
.line svg {
  -moz-transition: all .3s linear;
  -webkit-transition: all .3s linear;
  transition: all .3s linear;
}
.line.not-collapsed svg {
  -moz-transform:rotate(90deg);
  -webkit-transform:rotate(90deg);
  transform:rotate(90deg);
}
</style>
