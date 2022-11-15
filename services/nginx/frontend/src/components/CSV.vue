<!-- Component used to display csv content within a table layout -->
<template>
  <table class="color">
    <tr class="header">
      <th v-for="(element, index) in header" :key="index">{{ element }}</th>
    </tr>
    <tr v-for="(row, index) in body" :key="index">
      <td v-for="(element, index) in row" :key="index">{{ element }}</td>
    </tr>
  </table>
</template>

<script setup lang="ts">
import { Ref, ref, computed } from "vue";

const props = defineProps<{
  text: String,
}>();

const test: Ref<string> = ref(
  "A,B,C,D\n2,22,222,2222\n3,33,333,3333\n4,44,444,4444\n5,55,555,5555\n"
);

const rows = computed(function getRows(): string[][] {
  return props.text.split("\n").reduce((array: string[][], element: string) => {
    if (element !== "") {
      array.push(element.split(","));
    }
    return array;
  }, []);
});

const header = computed(function getHeader(): string[] {
  return rows.value[0];
});

const body = computed(function getBody(): string[][] {
  return rows.value.slice(1);
});
</script>

<style scoped>
table,
tr {
  border: 1px solid black;
  border-collapse: collapse;
}

tr.header {
  background-color: red;
  text-shadow: 1px 1px 0 #771c1c;
  color: white;
}

table.color {
  background-color: #eaeaea;
}

td:not(.full),
th:not(.full) {
  border-bottom: 1px solid black;
  border-right: 1px solid black;
}

.color {
  background-color: #fafafa;
}

tr:nth-child(even) {
  background-color: #fafafa;
}
</style>
