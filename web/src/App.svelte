<script lang="ts">
    import Prompt from "./components/Prompt.svelte";
    import Box from "./components/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, validatePrompt } from "./lib/Api.ts";
    import { onMount } from "svelte";
    import { auth } from "./lib/authService";
    import LoginButton from "./components/LoginButton.svelte";
    import LogoutButton from "./components/LogoutButton.svelte";

    let prompt = $state("");
    let convo = $state<{ id: number; question: string; answer: string }[]>([]);

    // get ConversationId for api calls from cookies
    // Note: We should ideally get this from a backend state associated with the user
    var conversationId = $state("");

    onMount(() => {
        const media = window.matchMedia("(prefers-color-scheme: dark)");

        const updateTheme = () => {
            document.documentElement.setAttribute(
                "data-theme",
                media.matches ? "dark" : "light",
            );
        };

        media.addEventListener("change", updateTheme);
        updateTheme();
    });

    onMount(async () => {
        await auth.initAuth();
    });

    let userId = $state<string | undefined>(undefined);
    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            userId = $auth.user.sub;
        } else {
            userId = undefined;
        }
    });

    let paramsChatRequest = $derived({
        prompt: "",
        userId: userId,
        conversationId: conversationId,
    });
    const chatUrl = "/chat";

    let isLoading = $state(false);
    async function onclick() {
        if (!userId) {
            console.error("User is not authenticated.");
            return;
        }

        const userQuestion = prompt;
        prompt = "";
        // push question early so UI shows it immediately
        convo.push({
            question: userQuestion,
            answer: "",
        });

        // Validate prompt first
        isLoading = true;
        try {
            const valResp = await validatePrompt({ userId, conversationId, prompt: userQuestion });

            // Backend may return either a simple { isValid, message } (spec) or the project's
            // ResponseForUser shape { message: { body }, userId, chatId }.
            const data = valResp.data ?? {};

            // If API uses explicit isValid flag, honour it
            if (typeof data.isValid === "boolean") {
                if (!data.isValid) {
                    const msg = data.message ?? "Prompt invalid";
                    convo[convo.length - 1].answer = typeof msg === "string" ? msg : msg?.body ?? JSON.stringify(msg);
                    isLoading = false;
                    return;
                }
            } else if (data.message) {
                // If backend returns ResponseForUser, check message body. If it is not the success message,
                // treat it as invalid and show the message returned.
                const body = (data.message && (data.message.body ?? data.message)) as string;
                if (body && body !== "Prompt validated successfully!") {
                    convo[convo.length - 1].answer = body;
                    // capture conversation id if backend returned it
                    conversationId = data.chatId ?? data.conversationId ?? conversationId;
                    isLoading = false;
                    return;
                }
                // if success, keep going
                conversationId = data.chatId ?? data.conversationId ?? conversationId;
            }

            // If validation passed, proceed to chat/generation
            paramsChatRequest.prompt = userQuestion;
            paramsChatRequest.userId = userId;
            paramsChatRequest.conversationId = conversationId;

            try {
                const answer = await getChatResponse(paramsChatRequest, chatUrl);
                convo[convo.length - 1].answer = answer.data.message.body;
                conversationId = answer.data.conversationId ?? "";
            } catch (err) {
                if (err && (err as any).isAxiosError) {
                    convo[convo.length - 1].answer = (err as any).response?.data?.message ?? "Server error";
                } else {
                    convo[convo.length - 1].answer = "no axios error returned - something went horribly wrong";
                }
            }
        } catch (err) {
            // validation call failed (network/server error)
            if (err && (err as any).isAxiosError) {
                convo[convo.length - 1].answer = (err as any).response?.data?.message ?? "Validation failed";
            } else {
                convo[convo.length - 1].answer = "Validation failed: unknown error";
            }
        } finally {
            isLoading = false;
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
                <Box msg={c.question} name="User" {userId} {conversationId} />
                {#if c.answer}
                    <Box
                        msg={c.answer}
                        name="Bot"
                        {userId}
                        {conversationId}
                        isCode={c.answer.startsWith("import")}
                    />
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
