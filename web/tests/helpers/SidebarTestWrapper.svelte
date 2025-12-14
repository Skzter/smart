<script lang="ts">
    import Sidebar from "../../src/lib/components/Sidebar.svelte";
    import * as SidebarUI from "../../src/lib/components/ui/sidebar/index.js";

    let open = $state(true);
</script>

<SidebarUI.Provider bind:open>
    <SidebarUI.Inset>
        <Sidebar />
    </SidebarUI.Inset>
</SidebarUI.Provider>
