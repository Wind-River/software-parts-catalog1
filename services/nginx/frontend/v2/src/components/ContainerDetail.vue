<template>
  <div v-if="!loaded">
    <h2>Loading</h2>
  </div>
  <div v-else>
    <modal v-if="showModal" @close="showModal = false" :payload="modalPayload">
      <div slot="header">
        <h2>Download Failed</h2>
      </div>
      <p>{{err}}</p>
    </modal>
    <table class="color">
      <tr class="header">
        <th>Field</th>
        <th>Value</th>
      </tr>
      <tr>
        <td>File Count</td>
        <td>{{data.count}}</td>
      </tr>
      <tr>
        <td>Date</td>
        <td>{{dateString}}</td>
      </tr>
      <tr>
        <td>License</td>
        <td>{{data.license}}</td>
      </tr>
      <tr>
        <td>Rationale</td>
        <td>{{data.rationale}}</td>
      </tr>
    </table>

    <br>

    <table class="color">
      <tr class="header">
        <th>Name</th>
        <th>Sha1</th>
        <th></th>
      </tr>
      <tr v-for="(archive, index) in data.archives" :key="index">
        <td>{{archive.name}}</td>
        <td>{{archive.sha1}}</td>
        <td v-if="archive.path != null"><button @click="downloadArchive(archive.id, archive.name)">Download</button></td>
        <td v-else></td>
      </tr>
    </table>
  </div>
</template>

<script>
import * as download from 'downloadjs'

export default {
  data () {
    return {
      cid: 0,
      loaded: false,
      data: {},
      showModal: false,
      err: 'Uninitialized Error'
    }
  },
  computed: {
    dateString () {
      var d = new Date(this.data.date)
      return d.getUTCFullYear() + '-' + d.getUTCMonth() + '-' + d.getUTCDate() + ' UTC'
    }
  },
  methods: {
    downloadArchive (id, name, depth = 0) {
      // var ok = true
      var ctype = ''
      fetch(new Request('/api/container/download/' + id, { method: 'GET', mode: 'same-origin' })).then((response) => {
        ctype = response.headers.get('Content-Type')
        // ok = response.ok
        return response.blob()
      }).catch((err) => {
        if (depth < 3) {
          downloadArchive(id, name, depth + 1)
        } else {
          this.err = JSON.stringify(err)
          this.showModal = true
        }
      }).then((blob) => {
        download(blob, name, ctype)
      })
    }
  },
  mounted () {
    this.cid = this.$route.params.id

    // var ok = true
    fetch(new Request('/api/container/' + this.cid, { method: 'GET', mode: 'same-origin' })).then((response) => {
      // ok = response.ok
      return response.json()
    }).then((obj) => {
      this.data = obj
      this.loaded = true
    }).catch((err) => {
      alert(JSON.stringify(err))
    })
  }
}
</script>

<style scoped>
  table, tr {
  border: 1px solid black;
  border-collapse: collapse;
  }

  tr.header {
  /* background-color: #B63A3A; */
  background-color: red;
  text-shadow: 1px 1px 0 #771C1C;
  color: #FFFFFF;
  }

  table.color {
  background-color: #EAEAEA;
  margin: 5px;
  }

  td:not(.full), th:not(.full) {
  border-bottom: 1px solid black;
  border-right: 1px solid black;
  }

  td.full, th.full {
  border: 1px solid black;
  text-align: center;
  }

  .color {
  background-color: #FAFAFA;
  }

  tr:nth-child(even) {
  background-color: #FAFAFA;
  }
</style>
