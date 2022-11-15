<!-- To be implemented or deleted -->
<template>
  <div>
      <h2>Group Search</h2>
      <label for="search-bar">Query: </label>
      <input id="search-bar" v-model="searchQuery" @keyup.enter="search">
      <br>
      <br>
      <h3 v-if="loading">Loading</h3>
      <table v-if="ready" class="color">
          <tr class="header">
              <th>Name</th>
              <th>Type</th>
              <th>#</th>
          </tr>
          <tr v-for="(row, index) in rows" :key="index" @click="redirect(row.id)">
              <td>{{ row.path }}</td>
              <td>{{ row.sub == 0 ? 'Child' : 'Parent' }}</td>
              <td>{{ row.sub == 0 ? row.packages : row.sub }}</td>
          </tr>
      </table>
  </div>
</template>

<script lang="ts">
import { Vue } from 'vue-class-component'

export default class GroupSearch extends Vue {
    searchQuery = ''
    rows: string[][] = []
    loading = false
    ready = false

    search () {
      this.ready = false

      fetch(new Request(
        `/api/group/search?query=${encodeURIComponent(this.searchQuery)}`,
        { method: 'GET', mode: 'same-origin' })).then((response) => {
        if (!response.ok) {
          throw response.statusText
        }

        return response.json()
      }).then((object) => {
        this.rows = object
        this.loading = false
        this.ready = true
      }).catch((err) => {
        this.loading = false
        alert(JSON.stringify(err))
      })
    }

    redirect (id: number): void {
      this.$router.push({
        name: 'Group Detail',
        params: { id: id }
      })
    }
}
</script>

<style scoped>
#search-bar {
    width: 100%;
}
</style>
