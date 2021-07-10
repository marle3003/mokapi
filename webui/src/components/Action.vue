<template>
  <div v-if="actions !== null">
    <p class="label">Actions</p>
    <b-table small hover class="dataTable" :items="actions" :fields="fields">
      <template v-slot:cell(show_details)="row">
        <div @click="toggleDetails(row)">
          <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
          <b-icon v-else icon="plus-square"></b-icon>
        </div>
      </template>
      <template v-slot:cell(duration)="data">
        {{ data.item.duration | duration }}
      </template>
      <template v-slot:row-details="row">
        <div v-for="step in row.item.steps" :key="step.id" >
          <div class="step" v-b-toggle="step.id">
            <b-icon icon="chevron-right"></b-icon>
            Run {{step.name}}
          </div>
          <b-collapse :id="step.id" class="logs">
            <div v-for="log in parseLog(step.log)" :key="log.line" class="pl-4 mt-1">
              <div>
                <div v-if="log.logs !== undefined">
                  <div class="line" v-b-toggle="'line'+log.line">
                    <span class="number">{{ log.line }}</span>
                    <b-icon icon="caret-right-fill" style="margin-right: 5px;" />
                    <span>{{ log.text }}</span>
                  </div>
                  <b-collapse :id="'line'+log.line" v-for="g in log.logs" :key="g.line">
                    <div class="line">
                      <span class="number">{{ g.line }}</span>
                      <span style="padding-left: 25px;white-space: pre-wrap;padding-bottom: 1px;">{{ g.text }}</span>
                    </div>
                  </b-collapse>
                </div>
                <div v-else>
                  <div class="line">
                    <span class="number">{{ log.line }}</span>
                    <span>{{ log.text }}</span>
                  </div>
                </div>
              </div>
            </div>
          </b-collapse>
        </div>
      </template>
    </b-table>
  </div>
</template>

<script>

import moment from 'moment'

export default {
  name: 'Action',
  props: ['actions'],
  data () {
    return {
      fields: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', 'duration', 'status'],
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
