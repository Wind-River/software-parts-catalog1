<!-- Component used for home page package search functionality -->
<template>
  <v-row class="d-flex justify-center">
    <div class="d-flex flex-column w-50 mt-12">
      <v-img
        src="catalog-logo2.png"
        height="200"
        width="auto"
        class="align-self-center"
      />
      <h2>Part Search</h2>
      <v-text-field
        id="search-bar"
        label="Query:"
        class="d-block"
        v-model="searchQuery"
        :rules="[(v) => v.length > 1 || 'Minimum Query Length is 2 Characters']"
        @keyup.enter="search"
        append-inner-icon="mdi-magnify"
        @click:append-inner="search"
      ></v-text-field>
      <br />
      <h3 v-if="fetching" class="my-4">Loading</h3>
      <h3 v-if="queryError" class="my-4">{{ queryError }}</h3>
      <h3 v-if="data && data.find_archive.length === 0">No results found</h3>
    </div>
  </v-row>
  <v-row v-if="data && data.find_archive.length > 0" class="justify-center">
    <v-table class="mx-8 w-75">
      <thead class="bg-primary">
        <tr>
          <th>{{ data.find_archive.length }}</th>
          <th>Name</th>
          <th>Date</th>
          <th>SHA256/SHA1</th>
          <th>Extraction Status</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in data.find_archive"
          :key="index"
          style="cursor: pointer"
          @click="redirect(row.archive.file_collection_id)"
        >
          <td>{{ index + 1 }}</td>
          <td>{{ row.archive.name }}</td>
          <td>{{ new Date(row.archive.insert_date).toLocaleDateString() }}</td>
          <td>
            {{
              row.archive.sha256
                ? "SHA256:" + row.archive.sha256.substring(0, 10) + "..."
                : "SHA1:" + row.archive.sha1.substring(0, 10) + "..."
            }}
          </td>
          <td>
            {{ row.archive.extract_status === 0 ? "Complete" : "In Progress" }}
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-row>
</template>

<script setup lang="ts">
import { Ref, ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useQuery } from "@urql/vue";

const searchQuery: Ref<string> = ref("");
const router = useRouter();
const result = useQuery({
  query: `
  query($searchQuery: String!){
  find_archive(query: $searchQuery, method: "fast"){
    archive{
      id
      name
      insert_date
      sha256
      sha1
      extract_status
      file_collection_id
    }
    distance
    }
  }
`,
  variables: { searchQuery },
  pause: true,
});
const data = result.data;
const fetching = result.fetching;
const queryError = result.error;

function search() {
  if (searchQuery.value.length > 1) {
    result.executeQuery();
  } else return;
}

function redirect(id: number): void {
  router.push({
    name: "Package Detail",
    params: { id: id },
  });
}
</script>
