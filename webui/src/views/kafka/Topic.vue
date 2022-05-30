<template>
  <div class="dashboard" v-if="topic !== null">
    <div class="page-header">
      <b-row class="mb-2 ml-0 mr-0">
        <b-col cols="auto" class="mr-auto pl-0">
          <h2>Kafka Topic</h2>
        </b-col>
        <b-col cols="auto" class="pr-0">
          <div class="close" @click="$router.go(-1)">
            <b-icon icon="x" class="border rounded p-1"></b-icon>
          </div>
        </b-col>
      </b-row>
    </div>
    <div class="page-body">
      <b-card-group deck>
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">Name</p>
            <p>{{ topic.name }}</p>
          </b-col>
        </b-row>
        <b-row v-if="topic.description !== ''">
          <b-col>
            <p class="label">Description</p>
            <vue-simple-markdown :source="topic.description" />
          </b-col>
        </b-row>
      </b-card>
      </b-card-group>
      <b-card-group deck>
        <b-card body-class="info-body" class="text-center">
          <b-card-title class="info">Records</b-card-title>
          <b-card-text class="text-center value">{{ metric(cluster.metrics, 'kafka_messages_total', {name: 'topic', value: topic.name}) }}</b-card-text>
        </b-card>
        <b-card body-class="info-body" class="text-center">
          <b-card-title class="info">Size</b-card-title>
          <b-card-text class="text-center value">{{ this.size | prettyBytes }}</b-card-text>
        </b-card>
      </b-card-group>

      <b-card-group deck>
        <b-card class="w-100">
          <b-card-title class="info text-center">Records</b-card-title>
          <b-table small hover class="dataTable" :items="messages" :fields="messageFields" style="table-layout: fixed" @row-clicked="toggleDetails">
            <template v-slot:cell(show_details)="row">
              <div @click="toggleDetails(row)">
                <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
                <b-icon v-else icon="plus-square"></b-icon>
              </div>
            </template>
            <template v-slot:cell(time)="data">
              {{ data.item.data.time | moment }}
            </template>
            <template v-slot:row-details="row">
              <b-card class="w-100">
                <b-row class="mb-2">
                  {{ row.item.message }}
                </b-row>
              </b-card>
            </template>
          </b-table>
        </b-card>
      </b-card-group>

      <b-card-group deck>
        <b-card class="w-100">
          <b-card-title class="info text-center">Partitions</b-card-title>
          <b-table small hover class="dataTable" :items="partitions" style="table-layout: fixed">
            <template v-slot:cell(leader)="data">
              <span v-if="data.item.leader !== null">
                <span v-if="data.item.leader.name.length > 0">{{ data.item.leader.name }}</span>
                {{ data.item.leader.addr }}
              </span>
            </template>
          </b-table>
        </b-card>
      </b-card-group>

      <b-card-group deck>
        <b-card class="w-100">
          <b-card-title class="info text-center">Groups</b-card-title>
          <b-table small hover class="dataTable" :items="groups" :fields="groupFields" style="table-layout: fixed">
            <template v-slot:cell(members)="data">
              <div v-for="member in data.item.members" :key="member.name">
                  <span :id="member.name">{{ member.name }}</span>
                  <b-popover :target="member.name" triggers="hover" placement="top">
                    <template #title>{{ member.name }}</template>
                    <p class="label">Address</p>
                     <p>{{ member.addr }}</p>
                     <p class="label">Client Software</p>
                     <p>{{ member.clientSoftwareName }} {{ member.clientSoftwareVersion }}</p>
                     <p class="label">Last Heartbeat</p>
                     <p>{{ member.heartbeat | fromNow }}</p>
                  </b-popover>
                </div>
            </template>
            <template v-slot:cell(lag)="data">
              <span>{{ metric(cluster.metrics, 'consumer_group_lag', {name: 'service', value: cluster.name}, {name: 'topic', value: topic.name}, {name: 'group', value: data.item.name}) }}</span>
            </template>
          </b-table>
        </b-card>
      </b-card-group>

    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import Shortcut from '@/mixins/Shortcut'

export default {
  name: 'Topic',
  mixins: [Api, Filters, Refresh, Metrics, Shortcut],
  data () {
    return {
      cluster: null,
      messages: null,
      messageFields: [{key: 'show_details', label: '', thStyle: 'width: 3%'}, {key: 'data.offset', label: 'Offset'}, {key: 'data.key', label: 'Key'}, {key: 'data.message', label: 'Message', class: 'kafka-message'}, 'partition', 'time'],
      groupFields: ['name', {key: 'lag', thStyle: 'width:5%'}, {key: 'state', thStyle: 'width:10%'}, {label: 'Strategy', key: 'protocol'}, 'coordinator', 'leader', 'members'],
      detailsShown: []
    }
  },
  computed: {
    topic: function () {
      if (!this.cluster) {
        return null
      }
      const name = this.$route.params.topic
      for (let topic of this.cluster.topics) {
        if (topic.name === name) {
          return topic
        }
      }
      return null
    },
    partitions: function () {
      var t = this.topic
      if (!t) {
        return null
      }
      return t.partitions
    },
    groups: function () {
      if (!this.cluster) {
        return null
      }
      const name = this.$route.params.topic
      const groups = []
      if (!this.cluster.groups) {
        return groups
      }

      for (let group of this.cluster.groups) {
        if (group.topics && group.topics.includes(name)) {
          groups.push(group)
        }
      }
      return groups
    }
  },
  methods: {
    async getData () {
      const name = this.$route.params.cluster
      this.$http.get(this.baseUrl + '/api/services/kafka/' + name).then(
        r => {
          this.cluster = r.data
        },
        r => {
          this.cluster = null
        }
      )
      this.$http.get(this.baseUrl + '/api/events?namespace=kafka&topic=' + this.$route.params.topic).then(
        r => {
          this.messages = r.data
        },
        r => {
          this.messages = []
        }
      )
    },
    toggleDetails (row) {
      row.toggleDetails()
      const index = this.detailsShown.indexOf(row.item.key)

      if (row.item._showDetails) {
        this.detailsShown.push(row.item.key)
      } else {
        this.detailsShown.splice(index, 1)
      }
    },
    shortcut (e) {
      console.log(e.key.toLowerCase())
      let cmd = e.key.toLowerCase()
      if (cmd === 'escape') {
        this.$router.go(-1)
      }
    }
  }
}
</script>

<style scoped>
.dashboard{
  width: 90%;
  margin: 42px auto auto;
}
.page-header h2{
  font-weight: 400;
}
.card{
  border-color: var(--var-border-color);
  margin: 7px;
}
.info{
  font-size: 0.7rem;
  font-weight: 300;
}
.info-body{
  padding: 0.8rem;
}
.value {
  font-size: 2.25rem;
  font-weight: 300;
}
</style>
<style>
.kafka-message{
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50%;
  width: 50%;
}
</style>
