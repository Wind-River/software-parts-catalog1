<!-- This page displays profile information such as cve list or bug list -->
<template>
  <v-container>
    <h3 v-if="profileFetching">Fetching...</h3>
    <h3 v-if="profileError">Error...</h3>
    <h3 v-if="profileData && JSON.parse(profileData.profile[0].document).label">
      {{ JSON.parse(profileData.profile[0].document).label }}
    </h3>
    <v-table v-if="profileData">
      <thead>
        <tr>
          <th
            class="bg-primary"
            v-if="profileKey === 'security'"
            v-for="(field, index) in Object.keys(
              JSON.parse(profileData.profile[0].document).cve_list[0],
            )"
            :key="index"
          >
            {{ field }}
          </th>
          <th
            class="bg-primary"
            v-if="profileKey === 'quality'"
            v-for="(field, index) in Object.keys(
              JSON.parse(profileData.profile[0].document).bug_list[0],
            )"
            :key="index"
          >
            {{ field }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="profileKey === 'security'" v-for="(cve, index) in JSON.parse(profileData.profile[0].document).cve_list" :key="index">
          <td
            v-for="(value, index) in Object.values(cve)"
            :index="index"
          >
            {{ value }}
          </td>
        </tr>
        <tr v-if="profileKey === 'quality'" v-for="(bug, index) in JSON.parse(profileData.profile[0].document).bug_list" :key="index">
          <td
            v-for="(value, index) in Object.values(bug)"
            :index="index"
          >
            {{ value }}
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
