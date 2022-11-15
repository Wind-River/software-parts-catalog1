<template>
    <div>
        <h2>Group Search</h2>
        <label for="search-bar">Query: </label>
        <input id="search-bar" v-model="searchQuery" @keyup.enter="search">
        <br>
        <br>
        <h3 v-if="loading">Loading</h3>
        <table v-if="ready" clas="color">
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

<script>
export default {
    data() {
        return {
            searchQuery: '',
            rows: [],
            loading: false,
            ready: false
        }
    },
    methods: {
        search() {
            this.loading = true
            this.ready = false
            var ok = true
            fetch(new Request('/api/group/search?query='+encodeURIComponent(this.searchQuery), { method: 'GET', mode: 'same-origin' })).then((response) => {
                ok = response.ok
                return response.json()
            }).then((obj) => {
                this.rows = obj
                this.loading = false
                this.ready = true
            }).catch((err) => {
                this.loading = false
                alert(JSON.stringify(err))
            })
        },
        redirect(id) {
            this.$router.push({ path: 'group/'+id })
        }
    }
}
</script>

<style>
    #search-bar {
        width: 100%;
    }
    table, tr {
        width: 100%;
        padding: 10px;
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
        background-color: #F0F0F0;
    }

    tr:nth-child(even) {
        background-color: #F0F0F0;
    }
</style>