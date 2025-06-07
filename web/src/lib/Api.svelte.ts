import axios from "axios";
const baseURL = "https://fakerapi.it/api/v2";
const url = "/books";

interface RequestResponse {
    question: string;
    answer: string;
}

const params = { 
    _locale: "de_DE",
    _quantity: 1,
}

export function getResponse(msg: string, convo) {
    axios({
	method: "get",
	url: url,
	baseURL: baseURL,
	params: params,
    }).then(function (response) {
	console.log(response);
	let QuestionAnswer: RequestResponse = {question: msg, answer: response.data.data[0].description} 
	convo.push(QuestionAnswer);
    }).catch((error) => {
	console.log(error);
    });
}
