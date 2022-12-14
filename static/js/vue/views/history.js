import API from '../apicaller.js'

export default {
  data() {
    return {
      hisloading: false,
      selected_item: {},
      dialogPlaySelected: false,
      pagesize: 30,
      pageix: 0,
      transition: 'scale-transition',
    }
  },
  created() {
    this.pageix = 0
    let req = { pageix: this.pageix, pagesize: this.pagesize }
    API.FetchHistory(this, req)
  },
  computed: {
    ...Vuex.mapState({
      history: state => {
        return state.hs.history
      },
      last_fetch: state => {
        return state.hs.last_fetch
      }
    })
  },
  methods: {
    askForPlayItem(item) {
      console.log('ask to play history item: ', item)
      this.selected_item = item
      this.selected_item.itemquestion = item.title
      if (item.title === ''){
        this.selected_item.itemquestion = item.uri
      }
      this.dialogPlaySelected = true
    },
    playSelectedItem() {
      console.log('playSelectedItem is: ', this.selected_item)
      this.dialogPlaySelected = false

      let req = { uri: this.selected_item.uri, force_type: this.selected_item.type }
      API.PlayUri(this, req)

      this.$router.push('/')
    },
    loadMore() {
      console.log('Load more')
      this.pageix += 1
      let req = { pageix: this.pageix, pagesize: this.pagesize }
      API.FetchHistory(this, req)
    }
  },
  template: `
  <v-container pa-1>
    <v-skeleton-loader
      :loading="hisloading"
      :transition="transition"
      height="94"
      type="list-item-three-line"
    >
      <v-card color="grey lighten-4" flat tile>
        <v-card-title> History</v-card-title>
        <v-container>
          <v-list dense nav>
            <v-list-item
              v-for="plitem in history"
              :key="plitem.id"
              @click="askForPlayItem(plitem)"
            >
              <v-list-item-icon>
                <v-icon>{{ plitem.icon }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ plitem.title }}</v-list-item-title>
                <v-list-item-title>{{ plitem.uri }}</v-list-item-title>
                <v-list-item-title>{{ plitem.playedAt }}</v-list-item-title>
                <v-list-item-title  v-if="plitem.duration">
                  Duration: {{ plitem.duration }}</v-list-item-title
                >
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-divider></v-divider>
          <v-row justify="center">
            <v-btn icon text @click="loadMore" :disabled="last_fetch"
              >More<v-icon>more_horiz</v-icon>
            </v-btn>
          </v-row>
        </v-container>
      </v-card>
    </v-skeleton-loader>
    <v-container>
      <v-dialog v-model="dialogPlaySelected" persistent max-width="290">
        <v-card>
          <v-card-title class="headline">Question</v-card-title>
          <v-card-text
            >Do you want to play "{{ selected_item.itemquestion }}" again?</v-card-text
          >
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="green darken-1" text @click="playSelectedItem"
              >OK</v-btn
            >
            <v-btn
              color="green darken-1"
              text
              @click="dialogPlaySelected = false"
              >Cancel</v-btn
            >
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-container>
  </v-container>`
}