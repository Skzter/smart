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
    if(userId === "undefined") {
	userId = "";
    }
    var conversationId = getCookie("conversationId");
    if(conversationId === "undefined") {
	conversationId = "";
    }

    let paramsUserInfo = {
        userId: userId,
    };
    const userInfoUrl = "/userInfo";
    let userData = {};

    // load UserData when opening the page
    onMount(async () => {
        if (userId !== undefined) {
            userData = await getUserInfo(paramsUserInfo, userInfoUrl);
            console.log(userData);
        }
    });

    let paramsChatRequest = {
        message: { data: "", agent: "user" },
        userId: userId,
        conversationId: conversationId,
    };
    const chatUrl = "/chat";

    async function onclick() {
        const userQuestion = prompt;
        prompt = "";
        paramsChatRequest.message.data = userQuestion;
        const answerPromise = getChatResponse(paramsChatRequest, chatUrl);
        convo.push({
            id: convo.length,
            question: userQuestion,
            answerPromise: answerPromise,
        });
    }

    function setIdsAsCookie(data) {
        if (userId === "") {
            userId = data.data.userId;
            setCookie("userId", userId);
        }
        if (conversationId === "") {
            conversationId = data.data.conversationId;
            setCookie("conversationId", conversationId);
        }
    }
</script>

<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2 overflow-auto px-4 pt-4 pb-28">
        {#each convo as c (c.id)}
            <Box msg={c.question} name="User" />
            {#await c.answerPromise}
                <Spinner color="blue" />
            {:then result}
                {setIdsAsCookie(result)}
                <Box msg={result.data.message.data} name="Bot" />
            {:catch error}
                {console.log(error)}
                <Box msg={error} name="Bot" />
            {/await}
        {/each}
        <Prompt bind:input={prompt} {onclick} />
    </div>
</div>
