<template>
  <div>
    <modal v-if="showModal" @close="showModal = false" :payload="modalPayload">
      <div slot="header">
        <h2>Processing Failed</h2>
        <p>{{ modalMessage }}</p>
      </div>
      <csv v-if="modalCSV" :text="modalData"/>
      <div v-else>{{ modalData }}</div>
    </modal>
    <div id="drag-drop"/>
    <button :disabled="!processingAllowed" @click="processPressed">Download CSV</button>
    <button @click="resetPressed">Reset</button>
    <div v-if="false">
      <input id="autobox" type="checkbox" v-model="autoUpload"/>
      <label for="autobox">Auto-Process</label>
    </div>

    <div v-if="(csv.length > 0)">
      <h4 v-if="(csvError.length == 0)">CSV Processed: {{ csv }}</h4>
      <h4 v-else style="color: red">CSV Not Processed: {{ csv }}</h4>
    </div>
    <p style="color: red">{{ csvError }}</p>

    <table class="no-border">
      <tr class="no-border">
        <td class="no-border">
          <div v-if="(files.length > 0) || (uploadElements.length > 0)" class="file-list">
            <template v-if="processing">
              <h4 :style="{display:'inline-block'}">Processing Files</h4>
              <div :style="{display:'inline-block',height:'20px',width:'20px'}" class="spinner"/>
            </template>
            <h4 v-else>Uploaded Files {{ files.length }}</h4>
            <ul>
              <li v-for="(file, index) in files" :key="index">
                {{file.filename}}
              </li>
            </ul>
          </div>
        </td>

        <td class="no-border">
          <div v-if="(uploadElements.length > 0) || (files.length > 0)" class="upload-list">
            <h4>Upload Progress {{ uploadElements.length }} Left</h4>
            <ul>
              <li v-for="(element, index) in uploadElements" :key="index" v-if="!element.progress.uploadComplete">
                {{element.data.fullPath}}: {{element.perc}}%
              </li>
            </ul>
          </div>
        </td>
      </tr>
    </table>
  </div>
</template>

<script>
import Modal from '@/components/Modal.vue'
import csv from '@/components/CSV.vue'

import Uppy from 'uppy/lib/core/Core.js'
import XHRUpload from 'uppy/lib/plugins/XHRUpload'
import DragDrop from 'uppy/lib/plugins/DragDrop'
import * as download from 'downloadjs'

export default {
  data () {
    return {
      files: [],
      uploadList: [],
      uploadDict: {},
      autoUpload: false,
      processing: false,
      newFile: false,

      modalData: '',
      modalPayload: {},
      modalCSV: false,
      showModal: false,

      csv: '',
      csvError: ''
    }
  },
  computed: {
    processingAllowed () {
      return (this.files.length > 0) && !this.processing
    },
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
  methods: {
    process () {
      if (!this.processing && this.files.length > 0) {
        this.processing = true
        var ok = true
        // var status = 200
        fetch(new Request('/api/upload/process', { method: 'POST', mode: 'same-origin', body: JSON.stringify(this.files) })).then((response) => {
          ok = response.ok
          // status = response.status
          return response.text()
        }).then((text) => {
          if (ok) {
            download(text, 'tk-prefilled.csv', 'text/csv')
            this.reset()
          } else {
            this.modalData = text
            this.modalPayload = {
              data: this.modalData,
              filename: 'tk-error.csv',
              mime: 'text/csv'
            }
            this.modalMessage = 'Errors encounterd processing the following files'
            this.modalCSV = true
            this.showModal = true
          }

          this.processing = false
        }).catch((result) => {
          alert(JSON.stringify(result))
          this.processing = false
        })
      }
    },
    reset () {
      console.log('this.reset()')
      this.files = []
      this.uploadList = []
      this.uploadDict = {}
      // leave this.autoUpload unchanged
      this.processing = false
      this.newFile = false
      this.csv = ''
      this.csvError = ''
    },
    processPressed (event) {
      this.process()
    },
    resetPressed (event) {
      this.reset()
    },
    debug () {
      // alert(JSON.stringify(this.uppy.getState()))
      console.log(this.uppy.getState())
    }
  },
  mounted () {
    const vm = this

    this.uppy = new Uppy({
      meta: { type: 'binary' },
      autoProceed: true
      // onBeforeUpload (files, done) {
      //   vm.uploadList = []
      //   vm.uploadDict = {}

      //   console.log('upload ' + Object.keys(files).length)
      //   for (var key in files) {
      //     let ind = vm.uploadList.push(files[key]) - 1
      //     vm.uploadDict[key] = ind
      //   }
      // }
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
        note: 'Then Download CSV after upload completes',
        locale: {
          strings: {
            dropHereOr: 'Drop files here or',
            browse: 'browse'
          }
        }
      }
    )
    this.uppy.on('file-added', (file) => {
      let ind = vm.uploadList.push(file) - 1
      vm.uploadDict[file.id] = ind
    })
    this.uppy.on('complete', (result) => {
      if (vm.files.length === vm.uploadList.length) {
        vm.uppy.reset()
        console.log('vm.uppy.reset()')
        if (vm.newFile && vm.autoUpload) {
          vm.newFile = false
          vm.process()
        }
      }
    })
    this.uppy.on('upload-error', (file, error) => {
      if (file.extension === 'csv') {
        vm.csv = file.name
        vm.csvError = error === '' ? 'upload-error' : error
      }
    })
    this.uppy.on('upload-success', (file, payload, uploadURL) => {
      if (payload.isMeta === false) {
        let f = {
          id: file.id,
          filename: payload.Filename,
          sha1: payload.Sha1,
          uploadname: payload.Uploadname,
          contentType: payload['Content-Type'],
          header: payload.Header
        }

        vm.files.push(f)
        vm.newFile = true
      } else {
        if (payload.hasOwnProperty('Extra')) {
          this.modalData = null // this.modalData = 'test'
          this.modalPayload = {
            data: payload.Extra,
            filename: 'tk-csv-error.csv',
            mime: 'text/csv'
          }
          this.modalMessage = 'Errors were encountered while updating. Details can be found by Downloading the attached CSV file.'
          this.modalCSV = false
          this.showModal = true
        }

        vm.csv = payload.Filename
        console.log('Set csv to ' + vm.csv)
      }
    })
    this.uppy.on('upload-progress', (file, progress) => {
      let f = vm.uploadList[vm.uploadDict[file.id]]
      f.progress.bytesTotal = progress.bytesTotal
      f.progress.bytesUploaded = progress.bytesUploaded
      f.progress.percentage = progress.bytesUploaded / progress.bytesTotal

      if (f.progress.percentage === 1) {
        f.progress.uploadComplete = true
      }
    })

    this.uppy.run()
  },
  components: {
    Modal,
    csv
  }
}
</script>

<style lang="scss">
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

  .spinner {
    -webkit-animation: spin 1.5s linear infinite;
            animation: spin 1.5s linear infinite;
    border: 3px solid #DDD;
    border-top-width: 3px;
    border-top-style: solid;
    border-radius: 50%;
  }

  @-webkit-keyframes spin {
    0% {
      border-top-color: #42A5F5;
    }
    50% {
      border-top-color: #EC407A;
    }
    100% {
      border-top-color: #42A5F5;
      -webkit-transform: rotate(360deg);
              transform: rotate(360deg);
    }
  }

  @keyframes spin {
    0% {
      border-top-color: #42A5F5;
    }
    50% {
      border-top-color: #EC407A;
    }
    100% {
      border-top-color: #42A5F5;
      -webkit-transform: rotate(360deg);
              transform: raotate(360deg);
    }
  }

  // .file-list {
  //   float: left;
  // }
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
