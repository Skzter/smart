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
export function getResponse(msg: string, convo) {
    axios({
        method: "get",
        url: url,
        baseURL: baseURL,
        params: params,
    })
        .then(function (response) {
            console.log(response);
            let QuestionAnswer: RequestResponse = {
                question: msg,
                answer: response.data.data[0].description,
            };
            convo.push(QuestionAnswer);
        })
        .catch((error) => {
            console.log(error);
        });
}
