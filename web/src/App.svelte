<script lang="ts">
    import Prompt from "./lib/Prompt.svelte";
    import Box from "./lib/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, getUserInfo } from "./lib/Api.svelte.ts";
    import { getCookie } from "typescript-cookie";
    //import { setCookie } from "typescript-cookie";

    let prompt = $state("");
    let convo = $state([]);

    // get UserId and ConversationId for api calls from cookies
    var userId = getCookie("userId");
    var conversationId = getCookie("conversationId");
    userId = "ashdf";

    //var allConversations = [];

    // load UserData when opening the page
    const userInfoUrl = "/userInfo";

    window.onload = async () => {
        if (userId !== undefined) {
            const userData = await getUserInfo(userId, userInfoUrl);
            console.log(userData);
        }
    };

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
        const UserQuestion = prompt;
        prompt = "";
        // for real api - currently not working
        paramsChatRequest.message.data = UserQuestion;
        const LLMAnswer = getChatResponse(params, url1);
        convo.push({
            id: convo.length,
            question: UserQuestion,
            answer: LLMAnswer,
        });
        /*
	if (userId === undefined){
	    userId = LLMAnswer.userId;
	    setCookie("userId", userId);
	}
	if (conversationId === undefined){
	    conversationId = LLMAnswer.conversationId;
	    setCookie("conversationId", conversation);
	}
	*/
    }
</script>

<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2 overflow-auto px-4 pt-4 pb-28">
        {#each convo as c (c.id)}
            <Box msg={c.question} name="User" />
            {#await c.answer}
                <Spinner color="blue" />
            {:then answer}
                <Box msg={answer} name="Bot" />
            {:catch error}
                {console.log(error)}
                <Box msg={error} name="Bot" />
            {/await}
        {/each}
        <Prompt bind:input={prompt} {onclick} />
    </div>
</div>
