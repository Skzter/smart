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

/** Validates the prompt by sending it to the /validationRes endpoint
 * @param body: object containing userId, conversationId, and prompt
 */
export async function validatePrompt(body: { userId: string; conversationId: string; prompt: string }) {
    return axios.post("validationRes", body, { baseURL });
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

export async function saveTestLocal(
    params: object,
    url: string = "/saveLocal",
) {
    const response = await axios({
        method: "post",
        url,
        baseURL: baseURL,
        data: params,
    });
    return response;
}

export async function runContainer(params: object, url: string = "/run") {
    const response = await axios({
        method: "post",
        url,
        baseURL: baseURL,
        data: params,
    });
    return response;
}
