<!-- Package file upload page -->
<template>
  <v-container>
    <v-card class="d-flex flex-column pa-8 mt-12 bg-secondary">
      <h2 class="px-8">Update Part Details</h2>
      <Upload
        type="text/csv"
        message="Click to select CSV"
        :processing="processing"
        @sendFiles="handleUpload"
      />
    </v-card>
    <v-table v-if="uploadedCSV.length > 0" class="ma-12">
      <thead>
        <tr>
          <th>Name</th>
          <th>Verification Code</th>
          <th>License</th>
          <th>Rationale</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(csv, index) in uploadedCSV" :key="index">
          <td>{{ csv.name }}</td>
          <td>{{ csv.verification_code }}</td>
          <td>{{ csv.license }}</td>
          <td>{{ csv.license_rationale }}</td>
          <td><v-icon icon="mdi-check" color="primary"></v-icon></td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
  <v-dialog v-model="showDialog" transition="scale-transition">
    <v-card width="50%" class="align-self-center">
      <h3 class="mx-6 mt-6">{{ dialogMessage }}</h3>
      <v-btn @click="showDialog = false" color="primary" class="ma-6"
        >Done</v-btn
      >
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useMutation } from "@urql/vue"
import { Ref, ref } from "vue"
import Upload from "@/components/Upload.vue"

type UpdateCSV = {
  checksum: string
  copyright: string
  insert_date: string
  license: string
  license_notice: string
  license_rationale: string
  name: string
  verification_code: string
}

const uploadedCSV: Ref<UpdateCSV[]> = ref([])
const processing: Ref<boolean> = ref(false)
const showDialog: Ref<boolean> = ref(false)
const dialogMessage: Ref<string> = ref("Uploaded CSV has completed processing.")

const updateMutation = useMutation(`
  mutation($verificationCode: String!, $license: String, $licenseRationale: String){
    updateFileCollection(verificationCode: $verificationCode, license: $license, licenseRationale: $licenseRationale){
      verification_code_one
      verification_code_two
      license_expression
      license_rationale
    }
  }
`)

function parseCSV(file: File): Promise<UpdateCSV[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (event) => {
      if (typeof event.target?.result === "string") {
        const resultString: string = event.target!.result
        uploadedCSV.value.push(...csvToArray(resultString))
        resolve(uploadedCSV.value)
      } else {
        reject("error reading uploaded file")
      }
    }
    reader.onerror = (event) => {
      reject(event.target?.error)
    }
    reader.readAsText(file)
  })
}

function csvToArray(csvString: string): UpdateCSV[] {
  const rows = csvString.slice(csvString.indexOf("\n") + 1).split("\n")
  const arr = rows.reduce((arr: UpdateCSV[], row: string) => {
    const fields = row.split(",")
    const name = fields[0]
    const insert_date = fields[1]
    const checksum = fields[2]
    const verification_code = fields[3]
    const license = fields[4]
    const license_rationale = fields[5]
    const license_notice = fields[6]
    const copyright = fields[7]
    arr.push({
      name,
      insert_date,
      checksum,
      verification_code,
      license,
      license_rationale,
      license_notice,
      copyright,
    })
    return arr
  }, [])
  return arr.filter((value) => {
    return value.name !== ""
  })
}

async function handleUpload(files: File[]) {
  processing.value = true
  uploadedCSV.value = []
  for (const file of files) {
    await parseCSV(file).catch((error) => {
      console.log(error)
    })
  }
  console.log(uploadedCSV.value)
  for (const csv of uploadedCSV.value) {
    if (csv.verification_code === undefined || "") {
      dialogMessage.value = "Verification code required to update parts."
    } else {
      updateMutation
        .executeMutation({
          verificationCode: csv.verification_code,
          license: csv.license,
          licenseRationale: csv.license_rationale,
        })
        .then((result) => {
          if (
            result.error?.message === "[GraphQL] no data was given to update"
          ) {
            dialogMessage.value = "No data was given to update."
          } else if (result.error) {
            dialogMessage.value =
              "Error processing uploaded CSV, please check formatting."
          } else {
            dialogMessage.value = "Uploaded CSV has completed processing."
          }
        })
        .catch((error) => {
          console.log(error)
        })
    }
    processing.value = false
    showDialog.value = true
  }
}
</script>
