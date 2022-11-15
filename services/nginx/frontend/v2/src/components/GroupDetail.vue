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

<script>
export default {
    data() {
        return {
            gid: 0,
            loaded: false,
            group: {},
            subs: [],
            packages: [],
            showModal: false,
            err: 'Uninitialized Error'
        }
    },
    computed: {
        hasSub() {
            return this.subs.length > 0
        },
        hasPackage() {
            return this.packages.length > 0
        }
    },
    methods: {
    },
    mounted() {
        this.gid = this.$route.params.id

        var ok = true
        fetch(new Request('/api/group/'+this.gid, { method: 'GET', mode: 'same-origin' })).then((response) => {
            ok = response.ok
            return response.json()
        }).then((obj) => {
            if (obj.group != null) {
                this.group = obj.group
            }
            if (obj.subs != null) {
                this.subs = obj.subs
            }
            if (obj.packages != null) {
                this.packages = obj.packages
            }
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
