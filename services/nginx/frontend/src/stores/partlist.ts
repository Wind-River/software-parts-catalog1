import { defineStore } from "pinia"
import { Ref, ref } from "vue"

// ## Graphql Schema Definition
// type PartList {
//   id: Int64!
//   name: String!
//   parent_id: Int64
// }

export type PartList = {
  id: number
  name: string
  parent_id: number
  children: PartList[]
}

export const usePartListStore = defineStore("partlist", () => {
  const partLists: Ref<PartList[]> = ref([])
  function setPartLists(list: PartList[]) {
    partLists.value = list
  }
  function deletePartList(id: number) {
    partLists.value = partLists.value.filter((value) => {
      return value.id !== id
    })
  }
  function addPartList(newList: PartList) {
    partLists.value.push(newList)
  }
  return {
    partLists,
    setPartLists,
    deletePartList,
    addPartList,
  }
})
