//Implements package search functionality with backend
export interface PackageRow {
  id: number
  name: string
  count: number
  sha1: string
  date: string
  packages: number

  loading: boolean
  depth: string
}

export type RowCount = number

export interface PackageCount {
  index: number
  id: number
  count: number
}

export interface PackageSearchService {
  loading: boolean
  // rows: PackageRow[]
  search (destination: PackageRow[], query: string, autofill: boolean): void
}

export function NewPackageSearchService (useWebSocket: boolean, host: string): PackageSearchService {
  if (!useWebSocket) {
    return new SynchronousSearchService()
  }

  console.info('Using WebSocket Search Service')
  return new WebSocketSearchService(host)
}

class SynchronousSearchService implements PackageSearchService {
  _loading: boolean = false

  get loading (): boolean {
    return this._loading
  }

  // _rows: PackageRow[] = []

  // get rows (): PackageRow[] {
  //   return this._rows
  // }

  search (destination: PackageRow[], query: string, autofill: boolean): void {
    this._loading = true
    // destination = []

    fetch(new Request(
      `/api/container/search?method=fast&query=${encodeURIComponent(query)}&depth=${encodeURIComponent('shallow')}`,
      { method: 'GET', mode: 'same-origin' }
    )).then((response) => {
      return response.json()
    }).then((object) => {
      destination = object.map((row: PackageRow) => {
        row.depth = 'shallow'
        return row
      })
      this._loading = false

      if (autofill) {
        this.updateCounts(destination).catch((err) => {
          console.error(err)
        }).finally(() => {
          this._loading = false
        })
      }
    }).catch((err) => {
      console.error(err)
      this._loading = false
    })
  }

  async updateCount (destination: PackageRow[], index: number): Promise<boolean> {
    destination[index].loading = true
    const row = destination[index]

    try {
      const response = await fetch(new Request(`/api/container/${row.id}`,
        { method: 'GET', mode: 'same-origin' })
      )
      const object = await response.json()
      destination[index].depth = 'deep'
      destination[index].count = object.count
      destination[index].loading = false
      return true
    } catch (err) {
      destination[index].loading = false
      console.error(err)
      return false
    }
  }

  async updateCounts (destination: PackageRow[]): Promise<void> {
    for (let index = 0; index < destination.length; index++) {
      const row = destination[index]
      if (row.loading || row.id <= 0 || row.depth === 'deep') {
        continue
      }

      try {
        if (!await this.updateCount(destination, index)) {
          console.error(`row[${index}] = ${row}; failed to update count`)
        }
      } catch (err) {
        console.error(`row[${index}] = ${row}; error updating count`)
      }
    }
  }
}

class WebSocketSearchService implements PackageSearchService {
  _loading: boolean = false

  get loading (): boolean {
    return this._loading
  }

  // _rows: PackageRow[] = []

  // get rows (): PackageRow[] {
  //   return this._rows
  // }

  host: string = ''
  ws?: WebSocket = undefined

  constructor (host: string) {
    this.host = host
  }

  search (destination: PackageRow[], query: string, autofill: boolean): void {
    console.debug(`WebSocketSearchService::search(${query}, ${autofill})`)
    this._loading = true
    // destination = []
    this.ws = new WebSocket(`ws://${window.location.host}/api/container/search?method=fast&query=${encodeURIComponent(query)}&depth=${encodeURIComponent('shallow')}&auto=${autofill}`)
    console.debug('created WebSocket')

    this.ws.onmessage = (message) => {
      // Receiving a RowCount signifies all rows have been sent
      const row: PackageRow | RowCount = JSON.parse(message.data)

      if (typeof row === 'number') {
        console.debug(`received RowCount: ${row}`)
        return this.finish(destination, query, autofill, row)
      }

      if (row.id > 0 && row.packages > 0) {
        row.depth = 'shallow'
        row.loading = false
      }
      console.debug(`received: ${JSON.stringify(row)}`)
      destination.push(row)
    }

    this.ws.onclose = () => {
      this._loading = false
    }
  }

  finish (destination: PackageRow[], query: string, autofill: boolean, rowCount: number): void {
    if (!autofill && this.ws !== undefined) {
      this.ws.close(1000)
    } else if (autofill) {
      if (this.ws === undefined) {
        console.error('autofill cannot proceed as ws is undefined')
        return
      }

      // change onmessage function to now accept updated package counts
      this.ws.onmessage = (message) => {
        const row: PackageCount = JSON.parse(message.data)
        destination[row.index].count = row.count
        destination[row.index].loading = false
        destination[row.index].depth = 'deep'

        if (this.ws !== undefined) {
          // look for next row to work on
          for (let i = row.index + 1; i < destination.length; i++) {
            const next = destination[i]
            if (next.loading || next.depth === 'deep') {
              continue
            }

            destination[i].loading = true

            // send row to count
            const payload = JSON.stringify({
              index: i,
              id: next.id
            })

            this.ws!.send(payload)
            return
          }

          // Reaching here means there is nothing else to work on
          this.ws.close(1000)
        }
      }

      // kickstart the process by sending the first message
      for (let i = 0; i < destination.length; i++) { // find first valid work
        const current = destination[i]
        if (current.loading || current.depth === 'deep') {
          continue
        }

        destination[i].loading = true

        const payload = JSON.stringify({
          index: i,
          id: current.id
        })

        this.ws!.send(payload)
        break
      }
    }
  }
}
