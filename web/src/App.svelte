<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import AppSidebar from "$lib/components/app-sidebar.svelte";
    import Main from "$lib/components/Main.svelte";
    import { user } from "$lib/shared.svelte";
    import SaveToaster from "$lib/components/SaveToaster.svelte";

    // hier dann auth
    user.id = "auth0|687270280dca20b77cfdcf73";
</script>

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
<SaveToaster />
