import axios from "axios";
// added dummy api for testing purposes
const baseURL = "https://fakerapi.it/api/v2";
const url = "/books";

// Type for conversation between user and llm
interface RequestResponse {
    question: string;
    answer: string;
}

// Params for API
const params = {
    _locale: "de_DE",
    _quantity: 1,
};

/**
 * Fetches data from the api
 * @param msg - Message from the user
 * @param convo - Array of Object of the whole conversation.New message with answer get pushed to array.
 * @example
 * getResponse("Gib mir eine zufällige Buchbeschreibung", [])
 */
export async function getResponse(msg: string): Promise<string> {
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
