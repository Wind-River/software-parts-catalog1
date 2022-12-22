import {defineStore} from "pinia"
import {ref, Ref} from "vue"

export const useSearchStore = defineStore("search", () => {
    const results: Ref<any[]> = ref([])
    function setResults(values: any[]) {
        results.value = values
    }
    return { results, setResults }
})