<template>
  <div
    class="dashboard"
    v-if="cluster !== null"
  >
    <div class="page-header">
      <b-row class="mb-2 ml-0 mr-0">
        <b-col
          cols="auto"
          class="mr-auto pl-0"
        >
          <h2>Kafka Cluster - {{ cluster.name }}</h2>
        </b-col>
        <b-col
          cols="auto"
          class="pr-0"
        >
          <div
            class="close"
            @click="$router.go(-1)"
          >
            <b-icon
              icon="x"
              class="border rounded p-1"
            ></b-icon>
          </div>
        </b-col>
      </b-row>
    </div>
    <div class="page-body">
      <b-card-group deck>
        <b-card class="w-100">
          <b-row>
            <b-col>
              <p class="label">Version</p>
              <p>{{ cluster.version }}</p>
            </b-col>
            <b-col>
              <p class="label">Contact</p>
              <p v-if="cluster.contact !== null">
                <a :href="cluster.contact.url">{{ cluster.contact.name }}</a>
                <a :href="'mailto:' + cluster.contact.email" v-if="cluster.contact.email">
                  <b-icon-envelope></b-icon-envelope>
                </a>
              </p>
            </b-col>
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
            class="dataTable"
            :items="cluster.topics"
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
              :items="cluster.groups"
              :fields="groupFields"
              table-class="dataTable"
            >
              <template v-slot:cell(members)="data">
                {{ data.item.members.map(x => x.name).join(', ') }}
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
      topicFields: [
        'name',
        'partitions',
        { key: 'messages', class: 'kafka-message' },
        'time'
      ],
      groupFields: [
        { key: 'name', class: 'text-left' },
        { key: 'state', thStyle: 'width:10%' },
        { key: 'assignmentStrategy', class: 'text-left' },
        { key: 'coordinator', class: 'text-left' },
        { key: 'leader', class: 'text-left' },
        { key: 'members', class: 'text-left' }
      ]
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
    },
    topicClickHandler (item) {
      this.$router.push({
        name: 'kafkaTopic',
        params: { cluster: this.$route.params.cluster, topic: item.name },
        query: { refresh: '5' }
      })
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
