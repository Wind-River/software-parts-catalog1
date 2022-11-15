<template>
  <table class="color">
    <tr class="header">
      <th v-for="(element, index) in header" :key="index">{{element}}</th>
    </tr>
    <tr v-for="(row, index) in body" :key="index">
      <td v-for="(element, index) in row" :key="index">{{element}}</td>
    </tr>
  </table>
</template>

<script>
export default {
  name: 'csv-table',
  props: ['text'],
  data () {
    return {
      test: 'A,B,C,D\n2,22,222,2222\n3,33,333,3333\n4,44,444,4444\n5,55,555,5555\n'
    }
  },
  computed: {
    rows () {
      return this.text.split('\n').reduce((array, element) => {
        if (element !== '') {
          array.push(element.split(','))
        }
        return array
      }, [])
    },
    header () {
      return this.rows[0]
    },
    body () {
      return this.rows.slice(1)
    }
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
