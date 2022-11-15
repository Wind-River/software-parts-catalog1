<template>
  <div>
    <button id="show-modal" @click="showModal = true">Show Modal</button>
    <modal v-if="showModal" @close="showModal = false" :payload="payload">
      <h2 slot="header">Files with Errors</h2>
      <csv :text="text"/>
    </modal>
  </div>
</template>

<script>
import Modal from '@/components/Modal.vue'
import csv from '@/components/CSV.vue'

export default {
  name: 'test',
  data () {
    return {
      showModal: false,
      text: 'A,B,C,D\n2,22,222,2222\n3,33,333,3333\n4,44,444,4444\n5,55,555,5555\n',
      payload: {
        data: 'A,B,C,D\n2,22,222,2222\n3,33,333,3333\n4,44,444,4444\n5,55,555,5555\n',
        filename: 'test.csv',
        mime: 'text/csv'
      }
    }
  },
  components: {
    Modal,
    csv
  }
}
</script>
