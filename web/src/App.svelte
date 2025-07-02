<script lang="ts">
    import Prompt from "./components/Prompt.svelte";
    import Box from "./components/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, getUserInfo } from "./lib/Api.ts";
    import { getCookie } from "typescript-cookie";
    import { onMount } from "svelte";

    let prompt = $state("");
    let convo = $state([]);

    // get UserId and ConversationId for api calls from cookies
    var userId = getCookie("userId");
    if (userId === undefined) {
        userId = "";
    }
    var conversationId = getCookie("conversationId");
    if (conversationId === undefined) {
        conversationId = "";
    }

    // for next sprint
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
            //setCookie("userId", userId); // for future oder so
            paramsChatRequest.userId = userId;
        }
        if (conversationId === "") {
            conversationId = data.data.conversationId;
            //setCookie("conversationId", conversationId); //for future oder so
            paramsChatRequest.conversationId = conversationId;
        }
    }

    const scrollToBottom = (node) => {
        const scroll = () =>
            node.scroll({
                top: node.scrollHeight,
                behavior: "smooth",
            });
        scroll();

        return { update: scroll };
    };
</script>

<main class="flex w-screen justify-center">
    <div
        class="flex flex-col h-[calc(100vh-132px)] overflow-y-auto w-8/10 gap-2 px-4 pt-4 pb-[132px]"
        use:scrollToBottom={convo}
    >
        {#each convo as c (c.id)}
            <Box msg={c.question} name="User" />
            {#await c.answerPromise}
                <Spinner color="blue" />
            {:then result}
                {setIdsAsCookie(result)}
                <Box msg={result.data.message.data} name="Bot" />
            {:catch error}
                <Box msg={error} name="Bot" />
            {/await}
        {/each}
    </div>
</main>
<footer class="mt-[0px] mb-[-10px] pb-[5px] h-[100px]">
    <Prompt bind:input={prompt} {onclick} />
</footer>
