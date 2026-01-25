<script lang="ts">
    import User from "../../src/lib/components/User.svelte";
    import * as SidebarUI from "../../src/lib/components/ui/sidebar/index.js";

    let open = $state(true);
</script>

<SidebarUI.Provider bind:open>
    <User />
</SidebarUI.Provider>
