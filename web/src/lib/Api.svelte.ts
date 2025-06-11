import axios from "axios";
// added dummy api for testing purposes
const baseURL = "https://fakerapi.it/api/v2";
const url = "/books";

/**
 * Fetches data from the api and returns the data
 * @param msg - Message from the user
 * @example
 * getResponse("Gib mir eine zufällige Buchbeschreibung")
 */
export async function getResponse(msg: string, params: object): Promise<string> {
    try {
        const response = await axios({
            method: "get",
            url: url,
            baseURL: baseURL,
            params: params,
        });

        const data = await response;
        console.log(data);
        return data.data.data[0].description;
    } catch (error) {
        throw error;
    }
}
