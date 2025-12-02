<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import AppSidebar from "$lib/components/app-sidebar.svelte";
    import Main from "$lib/components/Main.svelte";
</script>

<Sidebar.Provider>
    <AppSidebar />
    <Sidebar.Inset>
        <header
            class="flex justify-between h-16 shrink-0 items-center gap-2 border-b px-4"
        >
            <p>CHECK24</p>
            <p>Playwright Test AI</p>
        </header>
        <Main />
        <!-- In inset gehört die Main Component -->
    </Sidebar.Inset>
</Sidebar.Provider>
