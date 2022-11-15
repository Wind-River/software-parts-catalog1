<template>
  <div class="part">
    <table class="color">
        <thead>
            <tr class="header">
                <th v-if="aliases.length === 1">Aliases</th>
                <th>Sha256</th>
                <th>Sha1</th>
                <th>Size</th>
                <th>Mime Type</th>
                <th></th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td v-if="aliases.length === 1">{{fileNames}}</td>
                <td>{{fileSha256}}</td>
                <td>{{fileSha1}}</td>
                <td>{{fileSize}}</td>
                <td>{{fileMime}}</td>
                <td>
                  <button @click="servePart(fileSha256Bytes)">Download</button>
                  <button @click="viewPart(fileSha256Bytes)">View</button>
                </td>
            </tr>
        </tbody>
    </table>
    <p v-if="error">{{error}}</p>

    <table v-if="aliases.length > 1" class="color">
      <thead>
        <tr class="header">
          <th>#</th>
          <th>Alias</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(alias, index) in aliases" :key="index">
          <td>{{index + 1}}</td>
          <td>{{alias}}</td>
        </tr>
      </tbody>
    </table>

    <p class="view" v-if="view">
      {{viewData}}
    </p>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component'
import download from 'downloadjs'

interface apiPartsResponse {
    sha256: number[]
    sha1: number[]
    size: number
    mime: string
}

@Options({
  props: {
    identifier: String
  }
})
export default class Home extends Vue {
    identifier!: string

    aliases: string[] = []
    fileSha256Bytes: number[] = []
    fileSha1Bytes: number[] = []
    fileSize = 0
    fileMime = 'type/subtype;parameter=value'
    error = ''

    cache: Blob | null = null
    cacheSha256: number[] = []
    cacheMime = ''
    cacheName = ''

    viewData = ''

    bytesToHex (bytes: number[]): string {
      return bytes.reduce(
        function (previousValue: string, currentValue: number): string {
          const hex = currentValue.toString(16)
          if (hex.length === 1) {
            return `${previousValue}0${hex}`
          }

          return `${previousValue}${hex}`
        },
        ''
      )
    }

    hexToBytes (hex: string): number[] {
      return hex.match(/[\s\S]{1,2}/g)?.map(function (value: string): number {
        return parseInt(value, 16)
      }) || []
    }

    get fileSha256 (): string {
      return this.bytesToHex(this.fileSha256Bytes)
    }

    set fileSha256 (value: string) {
      this.fileSha256Bytes = this.hexToBytes(value)
    }

    get fileSha1 (): string {
      return this.bytesToHex(this.fileSha1Bytes)
    }

    set fileSha1 (value: string) {
      this.fileSha1Bytes = this.hexToBytes(value)
    }

    get fileNames (): string {
      if (this.aliases.length === 0) {
        return this.fileSha256
      } else if (this.aliases.length === 1) {
        return this.aliases[0]
      } else {
        return this.aliases.join(', ')
      }
    }

    fetchFile (): void {
      fetch(new Request(`/api/parts/${this.fileSha256}`, { method: 'GET', mode: 'same-origin' })).then((response: Response) => {
        return response.json()
      }).then((info: apiPartsResponse) => {
        this.fileSha256Bytes = info.sha256
        this.fileSha1Bytes = info.sha1
        this.fileSize = info.size
        this.fileMime = info.mime

        fetch(new Request(`/api/parts/${this.fileSha256}/aliases`, { method: 'GET', mode: 'same-origin' })).then((response: Response) => {
          return response.json()
        }).then((object) => {
          if (object.aliases instanceof Array) {
            this.aliases = object.aliases
          } else {
            console.error(`aliases returned ${JSON.stringify(object)}`)
          }
        }).catch((err) => {
          this.error = err
        })
      }).catch((err) => {
        this.error = err
      })
    }

    downloadPart (sha256: number[], name = ''): Promise<Response | void> {
      if (this.cache !== null) {
        return Promise.resolve()
      }

      if (name === '') {
        name = `part_sha256_${this.bytesToHex(sha256)}`
      }

      this.cacheName = name

      return fetch(new Request(`/api/parts/${this.bytesToHex(sha256)}/download`,
        { method: 'GET', mode: 'same-origin' }))
    }

    servePart (sha256: number[], name = ''): void {
      this.downloadPart(sha256, name).catch(error => {
        this.error = JSON.stringify(error)
      }).then(response => {
        if (response instanceof Response) {
          return response.blob()
        }

        if (this.cache instanceof Blob) {
          return Promise.resolve(this.cache)
        }

        throw new Error('cache is void')
      }).catch(error => {
        this.error = JSON.stringify(error)
      }).then(blob => {
        if (blob instanceof Blob) {
          download(blob, this.cacheName, this.cacheMime)
        }
      })
    }

    viewPart (sha256: number[], name = ''): void {
      this.downloadPart(sha256, name).catch(error => {
        this.error = JSON.stringify(error)
      }).then(response => {
        if (response instanceof Response) {
          return response.blob()
        }

        if (this.cache instanceof Blob) {
          return Promise.resolve(this.cache)
        }

        throw new Error('cache is void')
      }).catch(error => {
        this.viewData = JSON.stringify(error)
      }).then(blob => {
        if (blob instanceof Blob) {
          blob.text().catch(error => {
            this.viewData = JSON.stringify(error)
          }).then(text => {
            if (typeof text === 'string') {
              this.viewData = text
            } else {
              this.viewData = 'text was void'
            }
          })
        }
      })
    }

    get view (): boolean {
      return this.viewData !== ''
    }

    created (): void {
      // process identifier property
      if (this.identifier.startsWith('sha256:') || this.identifier.length === 64) {
        // identifier is a sha256
        let start = 0
        if (this.identifier.startsWith('sha256:')) {
          start += 7
        }

        this.fileSha256 = this.identifier.substring(start)
      }

      // initialize null values
      if (this.fileSha256Bytes === []) {
        this.fileSha256 = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
      }
      if (this.fileSha1Bytes === []) {
        this.fileSha1 = 'da39a3ee5e6b4b0d3255bfef95601890afd80709'
      }
    }

    mounted (): void {
      this.fetchFile()
    }
}
</script>

<style scoped>
  table, tr {
      border: 1px solid black;
      border-collapse: collapse;
  }

  td:not(.full), th:not(.full) {
    border-bottom: 1px solid black;
    border-right: 1px solid black;
  }

  td.full, th.full {
    border: 1px solid black;
    text-align: center;
  }

  .view {
    border: 1px solid black;
    background-color: #FAFAFA;
    margin: 1em;
    padding: 0.25em;
    border-radius: 0.5em;
  }
</style>
