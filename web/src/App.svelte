<script lang="ts">
    import Prompt from "./components/Prompt.svelte";
    import Box from "./components/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse } from "./lib/Api.ts";
    import { onMount } from "svelte";
    import { auth } from "./lib/authService";
    import LoginButton from "./components/LoginButton.svelte";
    import LogoutButton from "./components/LogoutButton.svelte";

    let prompt = $state("");
    let convo = $state<{ id: number; question: string; answer: string }[]>([]);

    // get ConversationId for api calls from cookies
    // Note: We should ideally get this from a backend state associated with the user
    var conversationId = localStorage.getItem("conversationId") || "";

    onMount(async () => {
        await auth.initAuth();
    });

    let userId: string | undefined;
    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            userId = $auth.user.sub;
        } else {
            userId = undefined;
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
        if (!userId) {
            console.error("User is not authenticated.");
            return;
        }

        const userQuestion = prompt;
        prompt = "";
        convo.push({
            question: userQuestion,
            answer: "",
        });
        isLoading = true;
        paramsChatRequest.message.data = userQuestion;
        paramsChatRequest.userId = userId;

        try {
            const answer = await getChatResponse(paramsChatRequest, chatUrl);
            convo[convo.length - 1].answer = answer.data.message.data;
            setConversationId(answer.data.conversationId);
        } catch (err) {
            if (err.isAxiosError) {
                convo[convo.length - 1].answer = err.response.data.message;
            } else {
                convo[convo.length - 1].answer =
                    "no axios error returned - something went horribly wrong";
            }
        } finally {
            isLoading = false;
        }
    }

    function setConversationId(newConversationId: string) {
        if (conversationId === "") {
            conversationId = newConversationId;
            localStorage.setItem("conversationId", conversationId);
            paramsChatRequest.conversationId = conversationId;
        }
    }

    let container: HTMLElement | undefined = $state();
    // Effect to trigger scrolling on relevant changes
    $effect(() => {
        if (container && (isLoading || convo.length > 0)) {
            // Small timeout to ensure DOM updates are complete
            setTimeout(() => {
                container.scrollTo({
                    top: container.scrollHeight,
                    behavior: "smooth",
                });
            }, 50);
        }
    });
</script>

{#if $auth.isAuthenticated}
    <div class="fixed left-4 top-4">
        <LogoutButton />
    </div>
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
{:else}
    <main
        class="flex h-screen w-screen items-center justify-center bg-gray-100 dark:bg-gray-900"
    >
        <div
            class="w-full max-w-sm rounded-lg bg-white p-8 text-center shadow-xl dark:bg-gray-800"
        >
            <h1 class="mb-4 text-2xl font-bold text-gray-900 dark:text-white">
                Welcome to Project Autotester
            </h1>
            <p class="mb-8 text-gray-500 dark:text-gray-400">
                Please log in to continue
            </p>
            <LoginButton />
        </div>
    </main>
{/if}
