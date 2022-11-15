<template>
  <div>
    <table>
      <thead>
        <tr>
          <th>Aliases</th>
          <th>Sha256</th>
          <th>Sha1</th>
          <th>Size</th>
          <th>Mime Type</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>{{fileNames}}</td>
          <td>{{fileSha25}}</td>
          <td>{{fileSha1}}</td>
          <td>{{fileSize}}</td>
          <td>{{fileMime}}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import Modal from '@/components/Modal.vue'
import csv from '@/components/CSV.vue'

// import * as download from 'downloadjs'

let bytesToHex = function (bytes) {
  let reducer = function(sum, elem) {
    return sum + ('0' + elem.toString(16)).slice(-2))
  }
      
  return bytes.reduce(reducer, '')
}

let hexToBytes = function (hex) {
  return hex.match(/[\s\S]{1,2}/g).map((elem) => {
    return parseInt(elem, 16)
  })
}

export default {
  data () {
    return {
      aliases: [],
      _fileSha256: [], // "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      _fileSha1: [], // "da39a3ee5e6b4b0d3255bfef95601890afd80709",
      fileSize: 0,
      fileMime: 'type/subtype;parameter=value'
    }
  },
  computed: {
    fileNames () {
      if (this.aliases.length === 0) {
        return bytesToHex(this.fileSha256)
      } else if (this.aliases.length === 1) {
        return this.aliases[0]
      } else {
        return this.aliases.join(', ')
      }
    },
    fileSha256: {
      get: function () {
        return bytesToHex(this._fileSha256)
      },
      set: function (value) {
        if (typeof (value) === 'string') {
          this._fileSha256 = hexToBytes(value)
        } else {
          this._fileSha256 = value
        }
      }
    },
    fileSha1: {
      get: function () {
        return bytesToHex(this._fileSha1)
      },
      set: function (value) {
        if (typeof (value) === 'string') {
          this._fileSha1 = hexToBytes(value)
        } else {
          this._fileSha1 = value
        }
      }
    }
  },
  methods: {
    fetchFile () {
      fetch(new Request(`/api/files/${this.fileSha256}`, { methad: 'GET', mode: 'same-origin' })).then((response) => {
        return response.json()
      }).then((json) => {
        this.fileSha256 = json.sha256
        this.fileSha1 = json.sha1
        this.fileSize = json.size
        this.fileMime = json.mime

        fetch(new Request(`/api/files/${this.fileSha256}/aliases`, { method: 'GET', mode: 'same-origin' })).then((response) => {
          return response.json()
        }).then((json) => {
          this.aliases = json.aliases
        }).catch((err) => {
          console.log(err)
        })
      }).catch((err) => {
        console.log(err)
      })
    }
  },
  mounted () {
    const vm = this
    vm.fileSha256 = vm.$route.params.sha256
    vm.fetchFile()
    // This is where uppy was initialized
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
