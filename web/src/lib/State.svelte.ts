// Testapi Fkt aber brauch man eigene Go Server für

import axios from "axios";
const url = "http://localhost:8080/api/hello" 
export function apiReq(msg: string) {
    console.log(msg);
    axios({
	method: "get",
	url: url,
    }).then(function (response) {
	console.log(response.data);
    }).catch((error) => {
	console.log(error);
    });
}
