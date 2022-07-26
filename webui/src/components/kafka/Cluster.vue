<template>
  <div v-if="cluster !== null">
    <b-card-group deck>
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p style="font-size: 1.25rem; font-weight: 500;">{{ cluster.name }}</p>
          </b-col>
          <b-col>
            <p class="label">Version</p>
            <p>{{ cluster.version }}</p>
          </b-col>
                    <b-col>
            <p class="label">Contact</p>
            <p v-if="cluster.contact">
              <a :href="cluster.contact.url">{{ cluster.contact.name }}</a>
              <a :href="'mailto:' + cluster.contact.email" v-if="cluster.contact.email">
                <b-icon-envelope></b-icon-envelope>
              </a>
            </p>
          </b-col>
        </b-row>
        <b-row>

        </b-row>
        <b-row v-if="cluster.description !== ''">
          <b-col>
            <p class="label">Description</p>
            <vue-simple-markdown :source="cluster.description" />
          </b-col>
        </b-row>
      </b-card>
    </b-card-group>
    <b-card-group deck>
      <b-card
        body-class="info-body"
        class="text-center"
      >
        <b-card-title class="info">Messages</b-card-title>
        <b-card-text class="text-center value">{{ metric(cluster.metrics, 'kafka_messages_total') }}</b-card-text>
      </b-card>
      <b-card
        body-class="info-body"
        class="text-center"
      >
        <b-card-title class="info">Size</b-card-title>
        <b-card-text class="text-center value">{{ this.size | prettyBytes }}</b-card-text>
      </b-card>
    </b-card-group>

    <b-card-group deck>
      <b-card class="w-100">
        <b-card-title class="info text-center">Topics</b-card-title>
        <b-table
          small
          hover
          class="dataTable selectable"
          :items="topics"
          :fields="topicFields"
          style="table-layout: fixed"
          @row-clicked="topicClickHandler"
        >
        <template v-slot:cell(partitions)="data">
            <span>{{ data.item.partitions.map(x => x.id).join(", ") }}</span>
          </template>
          <template v-slot:cell(messages)="data">
            {{ metric(cluster.metrics, 'kafka_messages_total', {name: 'topic', value: data.item.name}) }}
          </template>
          <template v-slot:cell(time)="data">
            <span>{{ metric(cluster.metrics, 'kafka_message_timestamp', {name: 'topic', value: data.item.name}) | moment }}</span>
          </template>
        </b-table>

        <b-card-title class="info text-center">Groups</b-card-title>
          <b-table
            :items="groups"
            :fields="groupFields"
            table-class="dataTable"
          >
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
          </b-table>
      </b-card>
    </b-card-group>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import Shortcut from '@/mixins/Shortcut'

import Header from '@/components/dashboard/Header'

export default {
  name: 'Topic',
  mixins: [Api, Filters, Refresh, Metrics, Shortcut],
  components: {
    'dashboard-header': Header
  },
  data () {
    return {
      cluster: null,
      topicFields: [
        'name',
        'partitions',
        { key: 'messages', class: 'kafka-message' },
        'time'
      ],
      groupFields: [
        { key: 'name', class: 'text-left' },
        { key: 'state', thStyle: 'width:10%' },
        { key: 'protocol', class: 'text-left', label: 'Strategy' },
        { key: 'coordinator', class: 'text-left' },
        { key: 'leader', class: 'text-left' },
        { key: 'members', class: 'text-left' }
      ]
    }
  },
  computed: {
    topics: function () {
      let result = this.cluster.topics.sort(this.compareByName)
      return result
    },
    groups: function () {
      let result = this.cluster.groups.sort(this.compareByName)
      return result
    }
  },
  methods: {
    async getData () {
      const name = this.$route.params.cluster
      if (!name) {
        return
      }

      this.$http.get(this.baseUrl + '/api/services/kafka/' + name).then(
        r => {
          this.cluster = r.data
        },
        r => {
          this.cluster = null
        }
      )
    },
    topicClickHandler (item) {
      this.$router.push({
        name: 'kafkaTopic',
        params: { cluster: this.$route.params.cluster, topic: item.name },
        query: { refresh: '5' }
      })
    },
    compareByName (x, y) {
      return x.name.localeCompare(y.name)
    }
  }
}
</script>

<style scoped>
.dashboard {
  width: 90%;
  margin: 42px auto auto;
}
.page-header h2 {
  font-weight: 400;
}
.card {
  border-color: var(--var-border-color);
  margin: 7px;
}
.info {
  font-size: 0.7rem;
  font-weight: 300;
}
.info-body {
  padding: 0.8rem;
}
.value {
  font-size: 2.25rem;
  font-weight: 300;
}
</style>
<style>
.kafka-message {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50%;
  width: 50%;
}
</style>
