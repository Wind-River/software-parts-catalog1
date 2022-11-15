<!-- To be implemented or deleted -->
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

<script lang="ts">
import { Vue, Options } from 'vue-class-component'

import Modal from '@/components/Modal.vue'
import csv from '@/components/CSV.vue'

import Uppy from '@uppy/core'
import XHRUpload from '@uppy/xhr-upload'
import DragDrop from '@uppy/drag-drop'
import { compareTwoStrings } from 'string-similarity'

import '@uppy/core/dist/style.css'
import '@uppy/drag-drop/dist/style.css'

type file = {
  id: string
  filename: string
  sha1: string
  uploadName: string
  contentType: string
  header: {
    Filename: string
    Header: Record<string, string[]>
  }
  match: boolean
  done: boolean
  archiveID?: string
}

type XHRUploadResponse = {
  status: number
  body: {
    Filename: string
    Uploadname: string
    Sha1: string
    isMeta: boolean
    'Content-Type': string
    Header: {
      Filename: string
      Header: {
        'Content-Disposition': string[]
        'Content-Type': string[]
      }
    }
    Extra?: string
  }
  uploadURL?: string
}

type XHRUploadProgress = {
  bytesUploaded: number
  bytesTotal: number
  uploadStarted: null | number // null or UNIX timestamp
  percentage: number // Integer [0, 100]
}

interface UppyFilePerc extends Uppy.UppyFile {
  perc?: string
}

type Row = {
  checksum_md5: string
  checksum_sha1: string
  checksum_sha256: string
  expression: string
  file_collection_id: number
  id: number
  insert_date: string
  license_rationale: string
  name: string
  path: string
  sha1: string
  size: number
  sortValue: number
}

@Options({
  components: {
    Modal,
    csv
  }
})
export default class Missing extends Vue {
  rows: Row[] = []
  loaded = false
  statusSort = 0
  searchText = ''
  showDetail = false
  detailData = { index: -1, data: 'data' }

  uploadList: UppyFilePerc[] = []
  uploadDict: Record<string, number> = {}
  files: file[] = []

  uppy: Uppy.Uppy<'strict'> | null = null

  get uploadElements (): UppyFilePerc[] {
    return this.uploadList.filter(function (value: UppyFilePerc): boolean {
      return value.progress?.percentage !== 1
    }).map(function (value: UppyFilePerc): UppyFilePerc {
      const perc = ((value.progress?.percentage || 0) * 100).toFixed(0)
      value.perc = perc
      return value
    })
  }

  sort (): void {
    console.log(`sort(${this.searchText}, ${this.statusSort})`)

    var top: Row[] = []
    var bottom: Row[] = []

    for (let i = 0; i < this.rows.length; i++) {
      const v = this.rows[i]

      if (this.searchText === '') {
        v.sortValue = 0
      } else {
        const sortValue = compareTwoStrings(v.name, this.searchText)
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

    const customCompare = function (a: Row, b: Row): number {
      let value = b.sortValue - a.sortValue
      if (value === 0) {
        value = a.id - b.id
      }

      return value
    }

    bottom.sort(customCompare)
    if (top.length > 0) {
      top.sort(customCompare)
      this.rows = top.concat(bottom)
    } else {
      this.rows = bottom
    }
  }

  radioSort (value: number): void {
    this.statusSort = value
    this.sort()
  }

  detail (index: number): void {
    if (this.detailData.index !== index) {
      const v = this.rows[index]
      this.detailData.data = 'Field, Value\n' +
      `Archive ID, ${v.id}\n` +
      `File Collection ID, ${v.file_collection_id}\n` +
      `Name, ${v.name}\n` +
      `Sha1, ${v.sha1}\n` +
      `Insert Date, ${v.insert_date}\n` +
      `License, ${v.expression}\n` +
      `Rationale, ${v.license_rationale}\n`
      this.detailData.index = index
    }

    this.showDetail = true
  }

  mounted (): void {
    this.loadCSV()

    if (this.uppy === null) {
      this.initUppy().run()
    } else {
      this.uppy.run()
    }
  }

  beforeUnmount (): void {
    if (this.uppy !== null) {
      this.uppy.close()
    }
  }

  loadCSV (): void {
    fetch(new Request('/api/missing/csv', { method: 'GET', mode: 'same-origin' })).then(response => {
      return response.json()
    }).then((object: Row[]) => {
      this.rows = object
      this.loaded = true
    }).catch(error => {
      alert(JSON.stringify(error))
    })
  }

  initUppy (): Uppy.Uppy<'strict'> {
    this.uppy = Uppy<Uppy.StrictTypes>({
      meta: {
        type: 'binary'
      },
      autoProceed: true
    })

    this.uppy.use(XHRUpload, {
      endpoint: '/api/upload',
      method: 'post',
      formData: true,
      fieldName: 'file',
      timeout: 0,
      limit: 3
    })

    this.uppy.use(
      DragDrop, {
        target: '#drag-drop',
        width: '100%',
        height: '100%',
        // note: 'Then Download CSV after upload completes',
        locale: {
          strings: {
            dropHereOr: 'Drop files here or %{browse}',
            browse: 'browse'
          }
        }
      }
    )

    this.uppy.on('file-added', file => {
      console.log(`file_added(${JSON.stringify(file)})`)
      const ind = this.uploadList.push(file) - 1
      this.uploadDict[file.id] = ind
    })

    this.uppy.on('complete', result => {
      if (this.files.length === this.uploadList.length) {
        if (this.uppy !== null) {
          this.uppy.reset()
        }
        console.log('this.uppy.reset()')
      }
    })

    this.uppy.on('upload-success', (file: Uppy.UppyFile, response: XHRUploadResponse) => {
      var archive: string | null = null
      if (response.body.Extra !== undefined && response.body.Extra !== '') {
        archive = JSON.parse(response.body.Extra)
        console.log(`archive: ${archive}`)
      }

      const f: file = {
        id: file.id,
        filename: response.body.Filename,
        sha1: response.body.Sha1,
        uploadName: response.body.Uploadname,
        contentType: response.body['Content-Type'],
        header: response.body.Header,
        match: archive !== null,
        done: false
      }

      if (archive !== null) {
        f.archiveID = archive
      }

      this.files.push(f)
      var ws = new WebSocket(`ws://${window.location.host}/api/missing/process`)
      ws.onopen = function (event) {
        console.log(`onopen: ${archive}`)
        ws.send(JSON.stringify(archive))
      }
      const onMessageFactory = function (self: Missing) {
        return function (ev: MessageEvent): void {
          const res = JSON.parse(ev.data)
          const archiveID = res.archiveID
          const sha1 = res.sha1
          console.log(`Searching for ${archiveID}`)

          for (let i = 0; i < self.files.length; i++) {
            const f = self.files[i]
            console.log(f)
            if (f.archiveID === archiveID) {
              f.done = true
              console.log(`Updated ${f}`)
            }
          }

          for (let i = 0; i < self.rows.length; i++) {
            const r = self.rows[i]
            if (r.id === archiveID) {
              r.sha1 = sha1
            }
          }

          ws.close()
        }
      }
      ws.onmessage = onMessageFactory(this)
    })

    this.uppy.on('upload-progress', (file: Uppy.UppyFile, progress: XHRUploadProgress) => {
      const f = this.uploadList[this.uploadDict[file.id]]
      if (f.progress === undefined) {
        f.progress = {
          bytesTotal: 0,
          bytesUploaded: 0,
          percentage: 0,
          uploadStarted: null,
          uploadComplete: false
        }
      }

      f.progress.bytesTotal = progress.bytesTotal
      f.progress.bytesUploaded = progress.bytesUploaded
      f.progress.percentage = progress.bytesUploaded / progress.bytesTotal

      if (f.progress.percentage === 1) {
        f.progress.uploadComplete = true
      }
    })

    return this.uppy
  }
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

  .uppy-DragDrop-arrow {
    width: 60px;
    height: 60px;
    fill: black;
    margin-bottom: 17px;
  }

  .uppy-DragDrop-note {
    font-size: 1em;
    color: black;
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
</style>
