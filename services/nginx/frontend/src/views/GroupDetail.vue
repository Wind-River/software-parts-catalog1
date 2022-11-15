<!-- To be implemented or deleted -->
<template>
    <div v-if="!loaded">
        <h2>Loading</h2>
    </div>
    <div v-else>
        <modal v-if="showModal" @close="showModal = false" :payload="modalPayload">
            <template v-slot:header>
                <h2>Download Failed</h2>
            </template>
            <p>{{err}}</p>
        </modal>
        <table class="color">
            <tr class="header">
                <th>Field</th>
                <th>Value</th>
            </tr>
            <tr>
                <td>Name</td>
                <td>{{group.path}}</td>
            </tr>
            <tr>
                <td>Sub-Groups</td>
                <td>{{group.sub}}</td>
            </tr>
            <tr>
                <td>Containers</td>
                <td>{{group.packages}}</td>
            </tr>
        </table>

        <br>

        <div v-if="hasSub">
            <h4>Sub-Groups</h4>
            <table class="color">
                <tr class="header">
                    <th>Name</th>
                    <th>Sub-Groups</th>
                    <th>Packages</th>
                </tr>
                <tr v-for="(sub, index) in subs" :key="index">
                    <td>{{sub.path}}</td>
                    <td>{{sub.sub}}</td>
                    <td>{{sub.packages}}</td>
                </tr>
            </table>

            <br>
        </div>

        <div v-if="hasPackage">
            <h4>Packages</h4>
            <table class="color">
                <tr class="header">
                    <th>Name</th>
                    <th>File Count</th>
                    <th>Date</th>
                    <th>Checksum</th>
                </tr>
                <tr v-for="(p, index) in packages" :key="index">
                    <td>{{p.name}}</td>
                    <td>{{p.count}}</td>
                    <td>{{p.date}}</td>
                    <td>{{p.sha1}}</td>
                </tr>
            </table>

            <br>
        </div>
    </div>
</template>

<script lang="ts">
import { Vue } from 'vue-class-component'

export default class GroupDetail extends Vue {
    gid = 0
    loaded = false
    group: any = {}
    subs = []
    packages = []
    showModal = false
    err = 'Uninitialized Error'

    get hasSub () {
      return this.subs.length > 0
    }

    get hasPackage () {
      return this.packages.length > 0
    }

    created () {
      var id: string
      if (typeof this.$route.params.id === 'string') {
        id = this.$route.params.id
      } else {
        id = this.$route.params.id[0]
      }

      this.gid = parseInt(id)
    }

    mounted () {
      fetch(new Request(`/api/group/${this.gid}`,
        { method: 'GET', mode: 'same-origin' })).then((response) => {
        if (!response.ok) {
          throw response.statusText
        }

        return response.json()
      }).then((object) => {
        if (object.group != null) {
          this.group = object.group
        }
        if (object.subs != null) {
          this.subs = object.subs
        }
        if (object.packages != null) {
          this.packages = object.packages
        }
        this.loaded = true
      }).catch((err) => {
        alert(JSON.stringify(err))
      })
    }
}
</script>

<style>

</style>
