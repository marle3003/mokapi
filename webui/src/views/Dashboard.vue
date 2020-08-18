<template>
  <div class="dashboard">
      <div class="page-header">
          <h2>Dashboard</h2>
      </div>
      <div class="page-body">

        <b-card-group deck>
        <b-card body-class="info-body">
          <b-card-title class="info">Uptime Since</b-card-title>
          <b-card-text class="text-center value">{{ dashboard.serverUptime | fromNow }}</b-card-text>
          <b-card-text class="text-right additional">{{ dashboard.serverUptime | moment }}</b-card-text>  
        </b-card>
        <b-card body-class="info-body">
          <b-card-title class="info">Total Requests</b-card-title>
          <b-card-text class="text-center value">{{ dashboard.totalRequests }}</b-card-text>
        </b-card>
        <b-card body-class="info-body">
          <b-card-title class="info">Request Errors</b-card-title>
          <b-card-text class="text-center value text-danger">{{ dashboard.requestsWithError }}</b-card-text>
        </b-card>
        </b-card-group>

        <b-card-group deck>
          <b-card>
            <b-card-title class="info">Services</b-card-title>        
            <b-card-text v-if="loaded">
              <div class="row">
              <div class="col-5 p-0">
            <doughnut-chart :height="250" :chartdata="serviceStatus" />
              </div>
              <div class="col-7 p-0 align-self-center">
                <b-list-group>
                  <b-list-group-item class="legend-item p-0 pb-2">
                    <span class="mr-auto">Success</span>
                    <b-badge>{{ dashboard.serviceStatus.total - dashboard.serviceStatus.errors }}</b-badge>
                  </b-list-group-item>
                  <b-list-group-item class="legend-item p-0 pb-2">
                    <span class="mr-auto">Errors</span>
                    <b-badge>{{ dashboard.serviceStatus.errors }}</b-badge>
                  </b-list-group-item>
                </b-list-group>
              </div>
              </div>
            </b-card-text>
          </b-card>
        </b-card-group>

        <b-card-group deck>
          <b-card class="w-100">
            <b-card-title class="info">Recent Request Errors</b-card-title>
            <b-table :items="dashboard.lastErrors" :fields="lastErrorsField">
              <template v-slot:cell(method)="data">
                <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
              </template>
              <template v-slot:cell(httpStatus)="data">
                <b-icon icon="circle-fill" class="response icon mr-1" variant="success" v-if="data.item.httpStatus >= 200 && data.item.httpStatus < 300"></b-icon>
                <b-icon icon="circle-fill" class="response icon mr-1" variant="warning" v-if="data.item.httpStatus >= 300 && data.item.httpStatus < 400"></b-icon>
                <b-icon icon="circle-fill" class="response icon mr-1 client-error" v-if="data.item.httpStatus >= 400 && data.item.httpStatus < 500"></b-icon>
                <b-icon icon="circle-fill" class="response icon mr-1" variant="danger" v-if="data.item.httpStatus >= 500 && data.item.httpStatus < 600"></b-icon>
                {{ data.item.httpStatus }}
              </template>
            </b-table>
          </b-card>
        </b-card-group>
      </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from 'moment'
import DoughnutChart from '@/components/DoughnutChart'

export default {
    components: {
      'doughnut-chart': DoughnutChart
    },
    mixins: [Api],
    data(){
      return {
        dashboard: {},
        loaded: false,
        timer: null,
        lastErrorsField: ['method', 'url', 'httpStatus', 'error', 'time'],
      }
    },
    created() {
      this.getData();
      this.timer = setInterval(this.getData, 20000)
    },
    computed: {
      serviceStatus: function(){
        let serviceStatus = this.dashboard.serviceStatus
        let success = serviceStatus.total - serviceStatus.errors
         return {datasets: [{ 
            data: [success, serviceStatus.errors],
            backgroundColor: ['rgb(110, 181, 110)', 'rgb(186, 86, 86)']
           }],
           labels: ["Success", "Errors"]};
      }
    },
    filters: {
      moment: function (date){
        return moment(date).format()
      },
      fromNow: function (date){
        return moment(date).fromNow(true);
      }
    },
    methods: {
      async getData () {
        this.dashboard = await this.getDashboard()
        this.loaded = true
      }
    },
    beforeDestroy () {
      clearInterval(this.timer)
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
    font-weight: 700;
}
  .card{
    margin: 15px;
  }
.card p{
    margin-bottom: 0;
}
.info{
    color: #a0a1a7;
    font-size: 1.2rem;
  }
.info-body{
  padding: 0.8rem;
}
.value {
  font-size: 1.8rem;
    font-weight: 600;
    color: #007bff;
}
.additional{
  color: #a0a1a7;
  font-size: 0.8rem;
}
.legend-item {
  border: 0 none;
  font-weight: 600;
}
.response.icon{
    vertical-align: middle;
    font-size: 0.5rem;
}
</style>