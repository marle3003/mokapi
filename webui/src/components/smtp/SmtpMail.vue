<template>
  <div v-if="mail !== null">
    <b-card class="w-100 mb-3">
      <b-row>
        <b-col>
          <div v-if="mail.sender !== null">
            <p class="label">Sender</p>
            <div>
              <span v-if="mail.sender.Name !== ''">{{ mail.sender.Name }} &lt;</span><span>{{ mail.sender.Address }}</span><span v-if="mail.sender.Name !== ''">&gt;</span>
            </div>
          </div>
          <div v-if="mail.from !== null && mail.from.length > 0">
            <p class="label">From</p>
            <div>
              <div v-for="from in mail.from" :key="from.Address">
                <span v-if="from.Name !== ''">{{ from.Name }} &lt;</span><span>{{ from.Address }}</span><span v-if="from.Name !== ''">&gt;</span>
              </div>
            </div>
          </div>
          <div v-if="mail.replyTo !== null && mail.replyTo.length > 0">
            <p class="label">Reply to</p>
            <div>
              <div v-for="replyTo in mail.replyTo" :key="replyTo.Address">
                <span v-if="replyTo.Name !== ''">{{ replyTo.Name }} &lt;</span><span>{{ replyTo.Address }}</span><span v-if="replyTo.Name !== ''">&gt;</span>
              </div>
            </div>
          </div>
          <div v-if="mail.contentType !== ''">
            <p class="label">Content type</p>
            <div>{{ mail.contentType}}</div>
          </div>
          <div v-if="mail.encoding !== ''">
            <p class="label">Encoding</p>
            <div>{{ mail.encoding}}</div>
          </div>
        </b-col>
        <b-col>
          <div v-if="mail.to !== null && mail.to.length > 0">
            <p class="label">To</p>
            <div>
              <div v-for="to in mail.to" :key="to.Address">
                <span v-if="to.Name !== ''">{{ to.Name }} &lt;</span><span>{{ to.Address }}</span><span v-if="to.Name !== ''">&gt;</span>
              </div>
            </div>
          </div>
          <div v-if="mail.cc !== null && mail.cc.length > 0">
            <p class="label">CC</p>
            <div>
              <div v-for="cc in mail.cc" :key="cc.Address">
                <span v-if="cc.Name !== ''">{{ cc.Name }} &lt;</span><span>{{ cc.Address }}</span><span v-if="cc.Name !== ''">&gt;</span>
              </div>
            </div>
          </div>
          <div v-if="mail.bcc !== null && mail.bcc.length > 0">
            <p class="label">BCC</p>
            <div>
              <div v-for="bcc in mail.bcc" :key="bcc.Address">
                <span v-if="bcc.Name !== ''">{{ bcc.Name }} &lt;</span><span>{{ bcc.Address }}</span><span v-if="bcc.Name !== ''">&gt;</span>
              </div>
            </div>
          </div>
          <p class="label">Time</p>
          <div>{{ mail.time | moment }}</div>
        </b-col>
      </b-row>
    </b-card>
    <b-card class="w-100" >
      <b-row>
        <b-col>
          <p class="label" v-if="mail.subject !== ''">Subject</p>
          <p>{{ mail.subject }}</p>
          <p class="label">Body</p>
          <p v-html="this.$options.filters.body(mail.htmlBody)" v-if="mail.htmlBody !== ''"></p>
          <p v-if="mail.textBody !== ''">{{ mail.textBody}}</p>
        </b-col>
      </b-row>
    </b-card>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Workflows from '@/components/Workflows'
import Shortcut from '@/mixins/Shortcut'

export default {
  name: 'SmtpMail',
  components: {
    'workflows': Workflows
  },
  mixins: [Api, Filters, Shortcut],
  data () {
    return {
      mail: null,
      event: null
    }
  },
  created () {
    this.getData()
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
       if (!id){
        return
      }
      this.event = await this.getEvent(id)
      this.mail = this.event.data
    },
  }
}
</script>

<style scoped>
</style>
<style>
.mailBody{
  margin-bottom: 5px;
}
</style>
