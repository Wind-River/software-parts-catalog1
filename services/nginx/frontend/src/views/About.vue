<!-- About Page -->
<template>
  <div class="d-flex flex-column align-center fill-height">
    <v-container>
      <v-img src="catalog-logo2.png" height="200" class="mt-12" />
      <h4 class="my-6 px-6">
        Maintaining a catalog of software components (parts) is a requirement
        whether generating SBOMs for managing license compliance, security
        assurance, export control, or safety certification. We developed a
        highly searchable scalable software parts database (catalog) enabling
        companies to seamlessly manage 1000s, 10,000s if not 100,000s of
        software parts from which their products are comprised.
      </h4>
    </v-container>
    <div class="d-flex w-100 mt-auto pr-8 bg-secondary justify-end">
      <h4 class="mx-2 text-grey-darken-1">Catalog: {{ catalogName }}</h4>
      <h4 class="mx-2 text-grey-darken-1">Total Parts: {{ totalParts }}</h4>
      <h4 class="mx-2 text-grey-darken-1">UI: {{ uiVersion.toPrecision(3) }}</h4>
      <h4 class="mx-2 text-grey-darken-1">Server: {{ serverVersion.toPrecision(3) }}</h4>
    </div>
  </div>
</template>
<script setup lang="ts">
const catalogName: string = "windriver.com/spc-prod-1"
const serverVersion: number = 1.0
const uiVersion: number = 1.0
const totalParts: string = "TBD"
</script>
