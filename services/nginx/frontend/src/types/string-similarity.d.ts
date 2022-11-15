declare module 'string-similarity';

declare function compareTwoStrings(str1: string, str2: string): number;

type Rating = {
    target: string
    rating: number
}
type Matches = {
    ratings: Rating[]
    bestMatch: Rating
}
declare function findBestMatch(mainString: string, targetStrings: string[]): Matches;