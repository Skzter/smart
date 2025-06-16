import axios from "axios";
// added dummy api for testing purposes
const baseURLdummy = "https://fakerapi.it/api/v2";
const baseURL = "/api/v1";

/**
 * Fetches data from the api and returns the data for the chat
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getChatResponse(params: object, url: string): string {
    try {
        const response = await axios({
            method: "get",
            url: url,
            baseURL: baseURLdummy,
            params: params,
        });

        const data = await response;
        return data.data.data[0].description;
    } catch (error) {
        console.log(error);
        throw error;
    }
}

/**
 * Fetches data from the api and returns the data for the user
 * @param params: parameters for api
 * @param url: url for api
 */
// needs fixing
export async function getUserInfo(params: object, url: string): object {
    try {
        const response = await axios({
            method: "post",
            url: url,
            baseURL: baseURL,
            data: params,
        });

        const data = await response;
        return data;
    } catch (error) {
        console.log(error);
        throw error;
    }
}
