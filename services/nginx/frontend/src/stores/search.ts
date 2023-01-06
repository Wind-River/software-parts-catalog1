import {defineStore} from "pinia"
import {ref, Ref} from "vue"

// ## Graphql Schema Definition
// type Archive {
//     id: Int64!
//     file_collection_id: Int64
//     file_collection: FileCollection
//     name: String
//     path: String
//     size: Int64
//     sha1: String
//     sha256: String
//     md5: String
//     insert_date: Time!
//     extract_status: Int!
//   }
type SearchResult = {
    archive: {
        extract_status: number
        file_collection_id: number
        file_collection: {
            license: {
                name: string
            }
        }
        id: number
        insert_date: string
        name: string
        sha1: string
        sha256: string
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