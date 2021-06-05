<template>
  <div class="dashboard" v-if="request !== null">
    <div class="page-header">
      <h2>HTTP Request</h2>
    </div>
    <div class="page-body">
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">URL</p>
            <p>{{ request.url}}</p>
            <p class="label">Method</p>
            <p><b-badge pill class="operation" :class="request.method.toLowerCase()" >{{ request.method }}</b-badge></p>
            <p class="label">Status</p>
            <p><b-badge pill variant="success" >{{ request.httpStatus }}</b-badge></p>
          </b-col>
          <b-col>
            <p class="label">Content Type</p>
            <p>{{ request.contentType }}</p>
            <p class="label">Time</p>
            <p> {{ request.time | moment}}</p>
            <p class="label">Duration</p>
            <p> {{ request.responseTime | duration}}</p>
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <p class="label">Request Parameters</p>
            <b-table small hover class="dataTable" :items="request.parameters" :fields="parameters">
              <template v-slot:cell(show_details)="row">
                <div @click="toggleDetails(row)" v-if="showRawParameter(row.item)">
                  <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
                  <b-icon v-else icon="plus-square"></b-icon>
                </div>
              </template>
              <template v-slot:cell(value)="data">
                {{ data.item.value !== '' ? data.item.value : data.item.raw }}
              </template>
              <template v-slot:cell(openapi)="data">
                {{ data.item.value !== '' ? 'yes' : 'no' }}
              </template>
              <template v-slot:row-details="row">
                <b-card class="w-100">
                  <b-row class="mb-2">
                    <b-col sm="3" class="text-sm-right"><b>Raw:</b></b-col>
                    <b-col><vue-simple-markdown :source="row.item.raw" /></b-col>
                  </b-row>
                </b-card>
              </template>
            </b-table>

            <p class="label">Actions</p>
            <b-table small hover class="dataTable" :items="request.actions" :fields="actions">
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
                <div v-for="step in row.item.steps" :key="step.id" class="p-1" v-b-toggle="step.id">
                  Run {{step.name}}
                  <b-collapse :id="step.id">
                    <vue-simple-markdown :source="step.log" class="pl-4 mt-1" style="font-size: 0.6rem" />
                  </b-collapse>
                </div>
              </template>
            </b-table>

            <p class="label">Response Body</p>
            <vue-simple-markdown :source="request.responseBody" />
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from "moment";

export default {
  name: 'HttpRequest',
  mixins: [Api],
  data () {
    return {
      request: null,
      parameters: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', 'value', 'type', {key: 'openapi', label: 'OpenApi'}],
      actions: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', 'duration', 'status']
    }
  },
  created () {
    this.getData()
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
      this.request = await this.getHttpRequest(id)
    },
    showRawParameter (p) {
      return p.value !== '' && p.value !== p.raw
    },
    toggleDetails(row) {
      row.toggleDetails()
      const index = this.detailsShown.indexOf(row.item.key)

      if (row.item._showDetails) {
        this.detailsShown.push(row.item.key)
      } else {
        this.detailsShown.splice(index, 1)
      }
    }
  },
  filters: {
    moment: function (date) {
      return moment(date).local().format('YYYY-MM-DD HH:mm:ss')
    },
    fromNow: function (date) {
      return moment(date).fromNow(true)
    },
    duration: function (time) {
      let ms = Math.round(time / 1000000)
      let d = moment.duration(ms)
      if (d.seconds() < 1) {
        return d.milliseconds() + " [ms]"
      } else if (d.minutes() < 1) {
        return d.seconds() + " [sec]"
      }
      return moment.duration(d).minutes()
    }
  }
}
</script>

<style scoped>
  .dashboard{
    width: 90%;
    margin: auto;
    margin-top: 42px;
  }
  .page-header h2{
    font-weight: 400;
  }
  .menu-item svg {
    -moz-transition: all .3s linear;
    -webkit-transition: all .3s linear;
    transition: all .3s linear;
  }
  .not-collapsed svg {
    -moz-transform:rotate(90deg);
    -webkit-transform:rotate(90deg);
    transform:rotate(90eg);
  }
  .actions{
    border-width: 1px;
    border-style: solid;
    border-radius: 0.3rem;
    padding: 0.3rem;
    margin-bottom: 1rem;
  }
  .actions > .title{
    margin-bottom: 0.3rem
  }
  .actions > .info{
    font-size: 0.6rem;
    margin-bottom: 0;
  }
</style>
