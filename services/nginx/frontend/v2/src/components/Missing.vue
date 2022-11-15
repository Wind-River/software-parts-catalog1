<template>
  <div>
    <modal v-if="showDetail" @close='showDetail = false' :payload=false>
      <csv :text="detailData.data"/>
    </modal>
    <div id="drag-drop"/>

    <div v-if="loaded">

      <br>
      <input type='radio' id='status-sort-none' value=0 :checked="statusSort==0" @click="radioSort(0)">
      <label for='status-sort-none'>#</label>
      <input type='radio' id='status-sort-x' value=1 :checked="statusSort==1" @click="radioSort(1)">
      <label for='status-sort-x' class='missing'>{{ '\u2717' }}</label>
      <input type='radio' id='status-sort-o' value=2 :checked="stausSort==2" @click="radioSort(2)">
      <label for='status-sort-o' class='loaded'>{{ '\u2713' }}</label>
      <input v-model='searchText' placeholder='Name Search' @keyup.enter='sort()'>

      <ol>
        <li v-for="(element, index) in uploadElements" :key="index">
          {{ element.data.fullPath }}: {{ element.perc }}%
        </li>
      </ol>
      <ul>
        <li v-for="(element, index) in files" :key="index">
          <span v-if="element.done">{{ element.filename}}: <span class='loaded'>{{ '\u2713' }}</span></span>
          <span v-else-if="element.match"> {{ element.filename }}: <span class="loading">processing</span></span>
          <span v-else>{{ element.filename }}: No Match</span>
        </li>
      </ul>

      <table class="color">
        <tr class="header">
          <th>Status</th>
          <th>Name</th>
          <th>Date</th>
          <th>License</th>
        </tr>
        <tr v-for="(row, index) in rows" :key="index" @click="detail(index)">
          <td v-if='row.sha1==""' class='missing'> {{ '\u2717' }} </td>
          <td v-else class='loaded'> {{ '\u2713' }} </td>
          <td>{{ row.name }}</td>
          <td>{{ row.insert_date.substr(0,10) }}</td>
          <td>{{ row.expression }}</td>
        </tr>
      </table>
    </div>
    <h1 v-else class="loading">Loading</h1>
  </div>
</template>

<script>

import Modal from '@/components/Modal.vue'
import csv from '@/components/CSV.vue'

import Uppy from 'uppy/lib/core/Core.js'
import XHRUpload from 'uppy/lib/plugins/XHRUpload'
import DragDrop from 'uppy/lib/plugins/DragDrop'
import Str from 'string-similarity'

export default {
  data () {
    return {
      rows: [],
      loaded: false,
      statusSort: 0,
      searchText: '',
      showDetail: false,
      detailData: { 'index': -1, 'data': 'data' },

      uploadList: [],
      uploadDict: {},
      files: []
    }
  },
  computed: {
    uploadElements () {
      return this.uploadList.reduce((array, element) => {
        if (!element.progress.uploadComplete) {
          let perc = (element.progress.percentage * 100).toFixed(0)
          element.perc = perc
          array.push(element)
        }

        return array
      }, [])
    }
  },
  // watch: {
  //   statusSort (val) {
  //     this.sort()
  //   }
  // },
  methods: {
    radioSort (value) {
      this.statusSort = value
      this.sort()
    },
    sort () {
      console.log('sort(' + this.searchText + ', ' + this.statusSort + ')')
      var top = []
      var bottom = []

      for (let i = 0; i < this.rows.length; i++) {
        let v = this.rows[i]

        if (this.searchText === '') {
          v.sortValue = 0
        } else {
          let sortValue = Str.compareTwoStrings(v.name, this.searchText)
          v.sortValue = sortValue
        }

        if (this.statusSort === 1 && v.sha1 === '') {
          // sort missing top
          top.push(v)
        } else if (this.statusSort === 2 && v.sha1 !== '') {
          // sort loaded top
          top.push(v)
        } else {
          bottom.push(v)
        }
      }

      function custom_compare (a, b) {
        let value = b.sortValue - a.sortValue
        if (value === 0) {
          value = a.id - b.id
        }

        return value
      }

      bottom.sort(custom_compare)
      if (top.length > 0) {
        top.sort(custom_compare)
        this.rows = top.concat(bottom)
      } else {
        this.rows = bottom
      }
    },
    detail (index) {
      if (this.detailData.index !== index) {
        let v = this.rows[index]
        this.detailData.data =
          ['Field', 'Value'].join(',') + '\n' +
          ['Archive ID', v.id].join(',') + '\n' +
          ['File Collection ID', v.file_collection_id] + '\n' +
          ['Name', v.name] + '\n' +
          ['Sha1', v.sha1] + '\n' +
          ['Insert Date', v.insert_date] + '\n' +
          ['License', v.expression] + '\n' +
          ['Rationale', v.license_rationale] + '\n'
        this.detailData.index = index
      }
      this.showDetail = true
    }
  },
  mounted () {
    var vm = this
    initUppy(vm)
    loadCSV(vm)
  },
  components: {
    Modal,
    csv
  }
}

function loadCSV (self) {
  // var ok = true
  fetch(new Request('/api/missing/csv', { method: 'GET', mode: 'same-origin' })).then((response) => {
    // ok = response.ok
    return response.json()
  }).then((obj) => {
    self.rows = obj // new Heap(obj, (a, b) => {return a.id == b.id;}, (a, b) => {return b.id - a.id;})
    self.loaded = true
  }).catch((err) => {
    alert(JSON.stringify(err))
  })
}

function initUppy (self) {
  self.uppy = new Uppy({
    meta: { type: 'binary' },
    autoProceed: true
  })
  self.uppy.use(XHRUpload, {
    endpoint: '/api/missing/upload',
    method: 'post',
    formData: true,
    fieldName: 'file',
    timeout: 0,
    limit: 3
  })
  self.uppy.use(
    DragDrop, {
      target: '#drag-drop',
      width: '100%',
      height: '100%',
      note: 'Then Download CSV after upload completes',
      locale: {
        strings: {
          dropHereOr: 'Drop files here or',
          browse: 'browse'
        }
      }
    }
  )
  self.uppy.on('file-added', (file) => {
    console.log('file-added(' + JSON.stringify(file) + ')')
    let ind = self.uploadList.push(file) - 1
    self.uploadDict[file.id] = ind
  })
  self.uppy.on('complete', (result) => {
    if (self.files.length === self.uploadList.length) {
      self.uppy.reset()
      console.log('self.uppy.reset()')
      if (self.newFile && self.autoUpload) {
        self.newFile = false
        self.process()
      }
    }
  })
  self.uppy.on('upload-error', (file, error) => {
    if (file.extension === 'csv') {
      self.csv = file.name
      self.csvError = error === '' ? 'upload-error' : error
    }
  })
  self.uppy.on('upload-success', (file, payload, uploadURL) => {
    // console.log('upload-success('+JSON.stringify(file)+', '+JSON.stringify(payload)+', '+uploadURL+')')
    var archive
    if (payload.Extra === '') {
      archive = false
    } else {
      archive = JSON.parse(payload.Extra)
      console.log('archive: ' + archive)
    }
    let f = {
      id: file.id,
      filename: payload.Filename,
      sha1: payload.Sha1,
      uploadname: payload.Uploadname,
      contentType: payload['Content-Type'],
      header: payload.Header,
      match: archive !== false,
      done: false
    }
    if (archive !== false) {
      f.archiveID = archive
    }

    let _index = self.files.push(f) - 1
    var ws = new WebSocket('ws://' + window.location.host + '/api/missing/process')
    ws.onopen = (event) => {
      console.log('onopen: ' + archive)
      ws.send(JSON.stringify(archive))
    }
    ws.onmessage = (event) => {
      let res = JSON.parse(event.data)
      let archiveID = res.archiveID
      let sha1 = res.sha1
      console.log('Searching for ' + archiveID)

      for (let i = 0; i < self.files.length; i++) {
        var f = self.files[i]
        console.log(f)
        if (f.archiveID === archiveID) {
          f.done = true
          console.log('Updated ' + f)
        }
      }
      for (let i = 0; i < self.rows; i++) {
        let r = self.rows[i]
        if (r.id === archiveID) {
          r.Sha1 = sha1
        }
      }

      ws.close()
    }
  })
  self.uppy.on('upload-progress', (file, progress) => {
    let f = self.uploadList[self.uploadDict[file.id]]
    f.progress.bytesTotal = progress.bytesTotal
    f.progress.bytesUploaded = progress.bytesUploaded
    f.progress.percentage = progress.bytesUploaded / progress.bytesTotal

    if (f.progress.percentage === 1) {
      f.progress.uploadComplete = true
    }
  })

  self.uppy.run()
}
</script>

<style scoped lang="scss">
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

  .loaded {
    color: green;
  }
  .missing {
    color: red;
  }

  .loading:after {
    overflow: hidden;
    display: inline-block;
    vertical-align: bottom;
    -webkit-animation: ellipsis steps(4, end) 900 ms infinite;
    animation: ellipsis steps(4, end) 900ms infinite;
    content: '\2026';
    width: 0px;
  }
  @keyframes ellipsis {
    to {
      width: 20px;
    }
  }
  @-webkit-keyframes ellipsis {
    to {
      width: 20px;
    }
  }

  /**/
  .no-border {
    border: none !important;
  }
  td.no-border {
    vertical-align: top;
    width: 50%;
  }

  // uppy DragDrop
  .uppy-DragDrop-container {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 7px;
    background-color: white;
    // cursor: pointer;
  }

  .uppy-DragDrop-inner {
    margin: 0;
    text-align: center;
    padding: 80px 20px;
    line-height: 1.4;
  }

  .uppy-DragDrop-arrow {
    width: 60px;
    height: 60px;
    fill: lighten(gray, 30%);
    margin-bottom: 17px;
  }

  .uppy-DragDrop-container.is-dragdrop-supported {
    border: 2px dashed;
    border-color: lighten(gray, 10%);
  }

  // .uppy-DragDrop-container.is-dragdrop-supported .uppy-DragDrop-dragText {
  //   display: inline;
  // }

  .uppy-DragDrop-container.drag {
    border-color: gray;
    background-color: darken(white, 10%);
  }

  .uppy-DragDrop-container.drag .uppy-DragDrop-arrow {
    fill: gray;
  }

  /* http://tympanus.net/codrops/2015/09/15/styling-customizing-file-inputs-smart-way/ */
  .uppy-DragDrop-input {
    width: 0.1px;
    height: 0.1px;
    opacity: 0;
    overflow: hidden;
    position: absolute;
    z-index: -1;
  }

  .uppy-DragDrop-label {
    display: block;
    cursor: pointer;
    font-size: 1.15em;
    margin-bottom: 5px;
  }

  .uppy-DragDrop-note {
    font-size: 1em;
    //color: lighten(gray, 10%);
    color: black;
  }

  .uppy-DragDrop-dragText {
    color: cornflowerblue;
  }

  // uppy utils

  @mixin reset-button() {
    background: none;
    -webkit-appearance: none;
    font-family: inherit;
    font-size: inherit;
    line-height: 1;
    padding: 0;
    margin: 0;
    border: 0;
    color: inherit;
  }

  // uppy common
    .uppy {
    // all: initial;
    box-sizing: border-box;
    font-family: -apple-system, BlinkMacSystemFont,
      'avenir next', avenir,
      helvetica, 'helvetica neue',
      ubuntu, roboto, noto,
      'segoe ui', arial, sans-serif;
    line-height: 1;
    // -webkit-font-smoothing: antialiased;
  }

  .uppy *, .uppy *:before, .uppy *:after {
    box-sizing: inherit;
  }

  // https://blog.prototypr.io/align-svg-icons-to-text-and-say-goodbye-to-font-icons-d44b3d7b26b4

  .UppyIcon {
    max-width: 100%;
    max-height: 100%;
    fill: currentColor;
    display: inline-block;
    vertical-align: text-top;
    overflow: hidden;
    // width: 1em;
    // height: 1em;
  }

  .UppyIcon--svg-baseline {
    bottom: -0.125em;
    position: relative;
  }

  // Buttons

  .UppyButton--circular {
    @include reset-button;
    box-shadow: 1px 2px 4px 0px rgba(black, 0.2);
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.3s;
  }

  .UppyButton--blue {
    color: white;
    background-color: cornflowerblue;

    &:hover,
    &:focus {
      background-color: darken(cornflowerblue, 10%);
    }
  }

  .UppyButton--white {
    color: asphalt-gray;
    background-color: white;

    &:hover,
    &:focus {
      color: white;
      background-color: darken(cornflowerblue, 10%);
    }
  }

  .UppyButton--yellow {
    color: white;
    background-color: yellow;

    &:hover,
    &:focus {
      background-color: darken(yellow, 5%);
    }
  }

  .UppyButton--green {
    color: white;
    background-color: green;

    &:hover,
    &:focus {
      background-color: darken(green, 10%);
    }
  }

  .UppyButton--red {
    color: white;
    background-color: red;

    &:hover,
    &:focus {
      background-color: darken(red, 10%);
    }
  }

  .UppyButton--sizeM {
    width: 60px;
    height: 60px;
  }

  .UppyButton--sizeS {
    width: 45px;
    height: 45px;
  }
</style>
