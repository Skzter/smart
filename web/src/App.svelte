<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import AppSidebar from "$lib/components/app-sidebar.svelte";
    import Main from "$lib/components/Main.svelte";

    // hier dann auth
    let userId = $state("auth0|687270280dca20b77cfdcf73");
</script>

<Sidebar.Provider>
    <AppSidebar {userId} />
    <Sidebar.Inset>
        <header
            class="flex justify-between h-16 shrink-0 items-center gap-2 border-b px-4"
        >
            <p>CHECK24</p>
            <p>Playwright Test AI</p>
        </header>
        <Main {userId} />
        <!-- In inset gehört die Main Component -->
    </Sidebar.Inset>
</Sidebar.Provider>
