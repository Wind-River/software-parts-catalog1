<template>
  <div class="part">
  <p v-if="error">{{error}}</p>

  <table>
      <thead>
          <tr>
              <!-- <th>Aliases</th> -->
              <th>Sha256</th>
              <th>Sha1</th>
              <th>Size</th>
              <th>Mime Type</th>
          </tr>
      </thead>
      <tbody>
        <template v-for="(part, index) in parts" :key="index">
          <tr @click="view(part.sha256)">
            <!-- <td>{{part.aliases}}</td> -->
            <td>{{bytesToHex(part.sha256)}}</td>
            <td>{{bytesToHex(part.sha1)}}</td>
            <td>{{part.size}}</td>
            <td>{{part.mime}}</td>
          </tr>
        </template>
      </tbody>
  </table>
  </div>
</template>

<script lang="ts">
import { Vue } from 'vue-class-component'

interface apiPartsResponse {
    sha256: number[]
    sha1: number[]
    size: number
    mime: string
}

export default class Home extends Vue {
    error = ''
    parts: apiPartsResponse[] = []

    view (sha256: number[]): void {
      const hex = this.bytesToHex(sha256)

      this.$router.push({
        name: 'View Part',
        params: { identifier: `sha256:${hex}` }
      })
    }

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

    fetchFiles (): void {
      fetch(new Request('/api/parts', { method: 'GET', mode: 'same-origin' })).then((response: Response) => {
        return response.json()
      }).then((info: apiPartsResponse[]) => {
        this.parts = info
      }).catch((err) => {
        this.error = err
      })
    }

    mounted (): void {
      this.fetchFiles()
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
</style>
