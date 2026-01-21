<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import AppSidebar from "$lib/components/Sidebar.svelte";
    import Main from "$lib/components/Main.svelte";
    import LoginButton from "$lib/components/LoginButton.svelte";
    import { user, apiToken } from "$lib/shared.svelte";
    import { onMount } from "svelte";
    import { Toaster } from "$lib/components/ui/sonner";
    import { auth } from "$lib/authService";
    import type { ApiToken } from "$types/api";
    import { getApiToken } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { AxiosError } from "axios";

    onMount(() => {
        const media = window.matchMedia("(prefers-color-scheme: dark)");

        const updateTheme = () => {
            document.documentElement.classList.toggle("dark", media.matches);
        };

        media.addEventListener("change", updateTheme);
        updateTheme();
    });

    onMount(async () => {
        await auth.initAuth();
    });

    let tokenFetched = false;

    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            user.id = $auth.user.sub;
            if (!tokenFetched) {
                tokenFetched = true;
                getToken();
            }
        } else {
            user.id = undefined;
            tokenFetched = false;
        }
    });

    async function getToken(): Promise<ApiToken | null> {
        try {
            let token = (await getApiToken()) as ApiToken;
            apiToken.token = token.token;
            return token;
        } catch (err) {
            if (err instanceof AxiosError) {
                let error = err.message;
                toast.error(error, {
                    description: "Das war wohl nichts mit der Historie.",
                });
            }
            apiToken.token = null;
            return null;
        }
    }
</script>

{#if $auth.isAuthenticated}
    <Sidebar.Provider>
        <AppSidebar />
        <Sidebar.Inset>
            <header
                class="flex justify-between h-16 shrink-0 items-center gap-2 px-4"
            >
                <p>CHECK24</p>
                <p>Playwright Test AI</p>
            </header>
            <Main />
        </Sidebar.Inset>
    </Sidebar.Provider>

    <Toaster richColors position="top-right" />
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
