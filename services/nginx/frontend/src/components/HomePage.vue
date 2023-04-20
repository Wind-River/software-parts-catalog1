<!-- Component used for home page package search functionality -->
<template>
  <v-row class="d-flex justify-center">
    <div class="d-flex flex-column w-50 mt-12">
      <v-img
        src="catalog-logo2.png"
        height="150"
        width="auto"
        class="align-self-center"
      />
      <h4>Part Search</h4>
      <v-text-field
        density="compact"
        id="search-bar"
        label="Query:"
        v-model="searchQueryInput"
        :rules="[(v) => v.length > 1 || 'Minimum Query Length is 2 Characters']"
        @keyup.enter="search"
        append-inner-icon="mdi-magnify"
        @click:append-inner="search"
      ></v-text-field>
      <br />
      <v-progress-linear
        v-if="fetching"
        indeterminate
        striped
        color="primary"
      ></v-progress-linear>
      <h3 v-if="queryError">{{queryError.networkError? "Network Unavailable": "Too Many Results"}}</h3>
      <h3 v-if="data && data.find_archive.length < 1" class="my-4">
        No results found
      </h3>
    </div>
  </v-row>
  <v-row v-if="searchStore.results && searchStore.results.length > 0" class="justify-center">
    <v-table class="mx-8" density="compact">
      <thead>
        <tr>
          <th class="bg-primary">{{ searchStore.results.length }}</th>
          <th class="bg-primary">Name</th>
          <th class="bg-primary">License</th>
          <th class="bg-primary">Date</th>
          <th class="bg-primary">SHA256/SHA1</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in searchStore.results"
          :key="index"
          style="cursor: pointer"
          @click="redirect(row.archive.part_id)"
        >
          <td>{{ index + 1 }}</td>
          <td>{{ row.archive.name }}</td>
          <td>{{ row.archive.part.license? row.archive.part.license : "" }}</td>
          <td>
            {{ new Date(row.archive.insert_date).toLocaleDateString() }}
          </td>
          <td>
            {{
              row.archive.sha256
                ? "SHA256:" + row.archive.sha256.substring(0, 10) + "..."
                : "SHA1:" + row.archive.sha1.substring(0, 10) + "..."
            }}
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-row>
</template>

<script setup lang="ts">
import { Ref, ref } from "vue"
import { useRouter } from "vue-router"
import { useQuery } from "@urql/vue"
import { useSearchStore } from "@/stores/search"

const searchQueryInput: Ref<string> = ref("")
const searchQuery: Ref<string> = ref("")
const router = useRouter()
const searchStore = useSearchStore()

//Query to retrieve parts from catalog based on user search input
const result = useQuery({
  query: `
  query($searchQuery: String!){
  find_archive(query: $searchQuery, method: "fast"){
    archive{
      sha256
      Size
      md5
      sha1
      name
      insert_date
      part_id
      part{
        license
      }
    }
    distance
    }
  }
`,
  variables: { searchQuery },
  pause: true,
})
const data = result.data
const fetching = result.fetching
const queryError = result.error

//Applies rules to search function
async function search() {
  if (searchQueryInput.value.length > 1) {
    searchQuery.value = searchQueryInput.value
    await result.executeQuery()
    if(data.value.find_archive){
      searchStore.setResults(data.value.find_archive)
    }
  } else return
}

//Redirects browser to package detail page for selected part
function redirect(id: string): void {
  router.push({
    name: "Package Detail",
    params: { id: id },
  })
}
</script>
