<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import AppSidebar from "$lib/components/Sidebar.svelte";
    import Main from "$lib/components/Main.svelte";
    import LoginButton from "$lib/components/LoginButton.svelte";
    import { user, getToken } from "$lib/shared.svelte";
    import { onMount } from "svelte";
    import { Toaster } from "$lib/components/ui/sonner";
    import { auth } from "$lib/authService";

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

    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            user.id = $auth.user.sub;
            getToken();
        } else {
            user.id = undefined;
        }
    });
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
