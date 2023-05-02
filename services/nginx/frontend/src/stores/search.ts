import {defineStore} from "pinia"
import {ref, Ref} from "vue"

// ## Graphql Schema Definition
// type Archive {
//     sha256: String!
//     size: Int64
//     part_id: UUID
//     part: Part
//     md5: String
//     sha1: String
//     name: String
//     insert_date: Time!
//   }

type SearchResult = {
    archive: {
        part_id: string
        part: {
            license: string
            automation_license: string
        }
        insert_date: string
        name: string
        sha1: string
        sha256: string
        size: number
        md5: string
    }
    distance: number
}

export const useSearchStore = defineStore("search", () => {
    const results: Ref<SearchResult[]> = ref([])
    function setResults(values: SearchResult[]) {
        results.value = values
    }
    return { results, setResults }
})