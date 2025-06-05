<script lang="ts">
    import {Button, Input} from "flowbite-svelte";
    import axios from "axios";
    // geht nicht
    const url = "http://localhost:8080/api/hello/";
    function apiReq() {
        const input = document.getElementById("user-msg") as HTMLInputElement;
        const msg = input.value;
        axios({
            method: "get",
            url: url,
            //data: { message: msg }
        }).then(function (response) {
            console.log(response.data);
        });
    }
</script>

<div class="flex">
    <Input type="text" id="user-msg" placeholder="Prompt" required />
    <Button color="purple" onclick={apiReq}>Send</Button>
</div>
