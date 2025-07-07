<script lang="ts">
    import Prompt from "./components/Prompt.svelte";
    import Box from "./components/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, getUserInfo } from "./lib/Api.ts";
    import { getCookie } from "typescript-cookie";
    import { onMount } from "svelte";

    let prompt = $state("");
    let convo = $state<{ id: number; question: string; answer: string }[]>([]);

    // get UserId and ConversationId for api calls from cookies
    var userId = getCookie("userId") || "";
    var conversationId = getCookie("conversationId") || "";

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

    let isLoading = $state(false);
    async function onclick() {
        const userQuestion = prompt;
        prompt = "";
        convo.push({
            question: userQuestion,
            answer: "",
        });
        isLoading = true;
        paramsChatRequest.message.data = userQuestion;
        try {
            const answer = await getChatResponse(paramsChatRequest, chatUrl);
            convo[convo.length - 1].answer = answer.data.message.data; // so oder so ähnlich
            setIdsAsCookie(answer); // resp id and user id setten
        } catch (err) {
            console.error("api call failed", err);
            convo[convo.length - 1].answer = "Interner Server Error - Bitte nochmal versuchen!";
        } finally {
            isLoading = false;
        }
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

    let container: HTMLElement;
    // Effect to trigger scrolling on relevant changes
    $effect(() => {
        if (container && (isLoading || convo.length > 0)) {
            // Small timeout to ensure DOM updates are complete
            setTimeout(() => {
                container.scrollTo({
                    top: container.scrollHeight,
                    behavior: "smooth",
                });
            }, 0);
        }
    });
</script>

<main class="flex w-screen justify-center">
    <div
        class="flex flex-col h-[calc(100vh-132px)] overflow-y-auto w-8/10 gap-2 px-4 pt-4"
        bind:this={container}
    >
        {#each convo as c}
            <Box msg={c.question} name="User" />
            {#if c.answer}
                <Box msg={c.answer} name="Bot" />
            {/if}
        {/each}
        {#if isLoading}
            <div class="flex justify-start p-4">
                <Spinner color="blue" />
            </div>
        {/if}
    </div>
</main>
<footer class="mt-[0px] mb-[-10px] pb-[5px] h-[100px]">
    <Prompt bind:input={prompt} {onclick} />
</footer>
