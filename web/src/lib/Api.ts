import axios from "axios";
const baseURL = "/api/v1/";

/**
 * Fetches data from the api and returns the data for the chat
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getChatResponse(params: object, url: string) {
    const response = await axios({
        method: "post",
        url: url,
        baseURL: baseURL,
        data: params,
    });
    return response;
}

/**
 * Fetches data from the api and returns the data for the user
 * @param params: parameters for api
 * @param url: url for api
 */
export async function getUserInfo(params: object, url: string) {
    const response = await axios({
        method: "post",
        url: url,
        baseURL: baseURL,
        data: params,
    });
    return response;
}

export async function getTemplate(url: string) {
    const response = await axios({
        method: "get",
        url: url,
        baseURL: baseURL,
    });
    return response;
}

export async function createTest(params: object, url: string = "/saveLocal") {
    const response = await axios({
        method: "post",
        url,
        baseURL: baseURL,
        data: params,
    });
    return response;
}