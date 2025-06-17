<script lang="ts">
    import Prompt from "./lib/Prompt.svelte";
    import Box from "./lib/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, getUserInfo } from "./lib/Api.ts";
    import { getCookie, setCookie } from "typescript-cookie";
    import { onMount } from "svelte";

    let prompt = $state("");
    let convo = $state([]);

    // get UserId and ConversationId for api calls from cookies
    var userId = getCookie("userId");
    var conversationId = getCookie("conversationId");

    // load UserData when opening the page
    const userInfoUrl = "/userInfo";
    let paramsUserInfo = {
        userId: userId,
    };

    let userData = {};
    onMount(async () => {
        if (userId !== undefined) {
            userData = await getUserInfo(paramsUserInfo, userInfoUrl);
            console.log(userData);
        }
    });

    //const chatUrl = "/chat";

    // Params for real api
    let paramsChatRequest = {
        message: { data: "", agent: 0 },
        userId: userId,
        conversationId: conversationId,
    };

    // Params for Dummy API
    const params = {
        _locale: "de_DE",
        _quantity: 1,
    };
    // Url for Dummy API
    const url1 = "/books";

    async function onclick() {
        const userQuestion = prompt;
        prompt = "";
        // for real api - currently not working
        paramsChatRequest.message.data = userQuestion;
        const answerPromise = getChatResponse(params, url1);
        convo.push({
            id: convo.length,
            question: userQuestion,
            answerPromise: answerPromise,
        });
        if (userId === "undefined") {
            console.log("giving userID");
            userId = answerPromise;
            setCookie("userId", userId);
        }
        if (conversationId === "undefined") {
            console.log("giving conversationId");
            conversationId = answerPromise;
            setCookie("conversationId", conversationId);
        }
        console.log("u: " + userId + " c: " + conversationId);
    }
    $inspect(convo);
</script>

<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2 overflow-auto px-4 pt-4 pb-28">
        {#each convo as c (c.id)}
            <Box msg={c.question} name="User" />
            {#await c.answerPromise}
                <Spinner color="blue" />
            {:then result}
                <Box msg={result.message.data} name="Bot" />
            {:catch error}
                {console.log(error)}
                <Box msg={error} name="Bot" />
            {/await}
        {/each}
        <Prompt bind:input={prompt} {onclick} />
    </div>
</div>
