<!-- This page displays profile information such as cve list or bug list -->
<template>
  <v-container style="max-width:100%;">
    <h3 v-if="profileFetching">Fetching...</h3>
    <h3 v-if="profileError">Error...</h3>
    <h3 v-if="profileData && profileData.profile[0].document.label">
      {{ profileData.profile[0].document.label }}
    </h3>
    <v-table v-if="profileData" fixed-header>
      <thead>
        <tr>
          <th
            class="bg-primary"
            v-if="profileKey === 'security'"
            v-for="(field, index) in Object.keys(
              profileData.profile[0].document.cve_list[0],
            )"
            :key="index"
          >
            {{ field }}
          </th>
          <th class="bg-primary" v-if="profileKey === 'licensing'">
            license_expression
          </th>
          <th class="bg-primary" v-if="profileKey === 'licensing'">
            analysis_type
          </th>
          <th class="bg-primary" v-if="profileKey === 'licensing'">
            comments
          </th>
          <th
            class="bg-primary"
            v-if="profileKey === 'quality'"
            v-for="(field, index) in Object.keys(
              profileData.profile[0].document.bug_list[0],
            )"
            :key="index"
          >
            {{ field }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="profileKey === 'security'" v-for="(cve, index) in profileData.profile[0].document.cve_list" :key="index">
          <td
            v-for="(value, index) in Object.values(cve)"
            :index="index"
          >
            {{ value? value : "None" }}
          </td>
        </tr>
        <tr v-if="profileKey === 'licensing'" v-for="(license, index) in profileData.profile[0].document.license_analysis" :key="index">
          <td>
            {{ license.license_expression? license.license_expression : "None" }}
          </td>
          <td>
            {{ license.analysis_type? license.analysis_type : "None" }}
          </td>
          <td>
            {{ license.comments? license.comments : "None" }}
          </td>
        </tr>
        <tr v-if="profileKey === 'quality'" v-for="(bug, index) in profileData.profile[0].document.bug_list" :key="index">
          <td
            v-for="(value, index) in Object.values(bug)"
            :index="index"
          >
            {{ value? value : "None" }}
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
</template>
<script setup lang="ts">
import { useQuery } from "@urql/vue"
import { onBeforeMount, Ref, ref } from "vue"
import { useRoute } from "vue-router"

//This section retrieves a profile based on key and part id
const route = useRoute()
const pid: Ref<string> = ref("")
const profileKey: Ref<string> = ref("")

const profileQuery = useQuery({
  query: `
        query($id: UUID!, $key: String!){
            profile(id: $id, key: $key){
                document
            }
        }`,
  variables: { id: pid, key: profileKey },
})

const profileData = profileQuery.data
const profileError = profileQuery.error
const profileFetching = profileQuery.fetching

//Collects information from route params
onBeforeMount(() => {
  var id: string
  var key: string
  if (typeof route.params.id === "string") {
    id = route.params.id
  } else {
    id = route.params.id[0]
  }
  if (typeof route.params.key === "string") {
    key = route.params.key
  } else {
    key = route.params.key[0]
  }

  pid.value = id
  profileKey.value = key
})
</script>
