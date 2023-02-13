<!-- About Page -->
<template>
  <div class="d-flex flex-column align-center fill-height">
    <v-container>
      <v-img src="catalog-logo2.png" height="150" class="mt-12" />
      <div class="my-4 px-6 text-body-2">
        Maintaining a catalog of software components (parts) is a requirement
        whether generating SBOMs for managing license compliance, security
        assurance, export control, or safety certification. We developed a
        highly searchable scalable software parts database (catalog) enabling
        companies to seamlessly manage 1000s, 10,000s if not 100,000s of
        software parts from which their products are comprised.
      </div>
    </v-container>
    <div class="d-flex w-100 mt-auto pr-8 bg-secondary justify-end">
      <div class="mx-1 text-grey-darken-1 text-caption">Catalog: {{ catalogName }}</div>
      <div class="mx-1 text-grey-darken-1 text-caption">Total Parts: {{ totalParts }}</div>
      <div class="mx-1 text-grey-darken-1 text-caption">UI: {{ uiVersion.toPrecision(3) }}</div>
      <div class="mx-1 text-grey-darken-1 text-caption">Server: {{ serverVersion.toPrecision(3) }}</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import router from '@/router';

const catalogName: string = window.location.href.substring(0, window.location.href.indexOf("/app")) ||"windriver.com/spc-prod-1"
const serverVersion: number = 1.0
const uiVersion: number = 1.0
const totalParts: string = "TBD"
</script>
