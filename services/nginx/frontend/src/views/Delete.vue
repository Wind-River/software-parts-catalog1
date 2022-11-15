<!-- To be implemented or deleted -->
<template>
    <div>
        <h2>Archive Management</h2>
        <p v-if="deleted !== null" class="green">Deleted {{deleted}}</p>
        <label for="search-bar">Archive ID: </label>
        <input id="search-bar" v-model="archiveIDQuery" @keyup.enter="search">
        <button @click="search">GET</button>
        <h4 v-if="loading">Loading</h4>
        <div v-if="archiveID != NaN && archiveID > 0 && !loading">
            <p>Archive: {{archiveName}}</p>
            <label for="license">License Expression: </label>
            <input id="license" v-model="licenseInput" @keyup.enter="update">
            <label for="rationale">Rationale: </label>
            <input id="rationale" v-model="rationaleInput" @keyup.enter="update">
            <button @click="update">Update</button>
            <button @click="del">Delete</button>
        </div>
        <div v-if="error !== null">
            <p class="error">{{error}}</p>
        </div>
    </div>
</template>

<script>
export default {
  data () {
    return {
      archiveIDQuery: '',
      archiveID: 0,
      loading: false,
      archiveName: '',
      archiveLicense: '',
      archiveRationale: '',
      licenseInput: '',
      rationaleInput: '',
      deleted: null,
      error: null
    }
  },
  methods: {
    reset () {
      this.archiveIDQuery = ''
      this.archiveID = 0
      this.loading = false
      this.archiveName = ''
      this.archiveLicense = ''
      this.archiveRationale = ''
      this.licenseInput = ''
      this.rationaleInput = ''
      this.deleted = null
      this.error = null
    },
    search () {
      this.error = null
      this.archiveID = parseInt(this.archiveIDQuery)
      if (!isNaN(this.archiveID) && this.archiveID > 0) {
        // Do Search
        this.loading = true
        var ok = true
        fetch(new Request(`/api/archive?query=${encodeURIComponent(this.archiveID)}`,
          { method: 'GET', mode: 'same-origin' })).then((response) => {
          ok = response.ok
          this.loading = false
          if (ok) {
            return response.json()
          } else {
            return response.text()
          }
        }).then((obj) => {
          if (ok) {
            this.archiveName = obj.Name

            this.archiveLicense = obj.LicenseExpression
            this.licenseInput = obj.LicenseExpression

            if (obj.LicenseRationale != null) {
              this.archiveRationale = obj.LicenseRationale
              this.rationaleInput = obj.LicenseRationale
            } else {
              this.archiveRationale = null
              this.rationaleInput = ''
            }
          } else {
            this.reset()
            this.error = obj
          }
        }).catch((err) => {
          this.loading = false
          alert(JSON.stringify(err))
        })
      } else {
        this.error = this.archiveIDQuery + ' is Not a Number'
      }
    },
    update () {
      var payload = {
        ArchiveID: this.archiveID,
        LicenseExpression: null,
        LicenseRationale: null
      }
      if (this.archiveLicense !== this.licenseInput) {
        payload.LicenseExpression = this.licenseInput
      }
      if (this.archiveRationale !== this.rationaleInput) {
        payload.LicenseRationale = this.rationaleInput
      }

      if (payload.LicenseExpression != null || payload.LicenseRationale != null) { // make sure there is ANYTHING to be updated
        var ok = false
        fetch(new Request('/api/archive'),
          { method: 'POST', body: JSON.stringify(payload), mode: 'same-origin', headers: { 'Content-Type': 'application/json' } }).then(response => {
          ok = response.ok
          if (ok) {
            this.search()
          } else {
            return response.text()
          }
        }).then((obj) => {
          this.error = obj
        }).catch((err) => {
          alert(JSON.stringify(err))
        })
      }
    },
    del () {
      var ok = true
      fetch(new Request(`/api/archive/${this.archiveID}`),
        { method: 'DELETE', mode: 'same-origin' }).then(response => {
        ok = response.ok

        if (ok) {
          const tmp = this.archiveID
          this.reset()
          this.deleted = tmp
        } else {
          return response.text()
        }
      }).then((obj) => {
        this.error = obj
      }).catch((err) => {
        alert(JSON.stringify(err))
      })
    }
  }
}
</script>

<style scoped>
  .error {
      color: red;
  }
  .green {
      color: green;
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
</style>
