import axios from "axios";
// added dummy api for testing purposes
const baseURLdummy = "https://fakerapi.it/api/v2";
const baseURL = "/api/v1";

/**
 * Fetches data from the api and returns the data for the chat
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getChatResponse(params: object, url: string) {
    try {
        const response = await axios({
            method: "get",
            url: url,
            baseURL: baseURLdummy,
            params: params, //bei post muss data heißen
        });
        const answer = {
            message: {
                data: response.data.data[0].description,
                agent: "system",
            },
            userId: "1",
            conversationId: "1",
        };
        return answer;
    } catch (error) {
        return error;
    }
}

/**
 * Fetches data from the api and returns the data for the user
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getUserInfo(params: object, url: string) {
    try {
        const response = await axios({
            method: "post",
            url: url,
            baseURL: baseURL,
            data: params,
        });
        return response;
    } catch (error) {
        return error;
    }
}
