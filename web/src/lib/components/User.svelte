<script lang="ts">
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { CircleUserRound, EllipsisVertical, LogOut } from "@lucide/svelte";
    import { auth } from "$lib/authService";
    const sidebar = Sidebar.useSidebar();
</script>

<DropdownMenu.Root>
    <DropdownMenu.Trigger>
        <Sidebar.MenuButton
            size="lg"
            class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
        >
            <CircleUserRound />
            <div class="grid flex-1 text-start text-sm leading-tight">
                <span class="truncate font-medium">Autotester</span>
            </div>
            <EllipsisVertical />
        </Sidebar.MenuButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content
        class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
        side={sidebar.isMobile ? "bottom" : "right"}
        align="end"
        sideOffset={4}
    >
        <DropdownMenu.Item onclick={auth.logout}>
            <LogOut />
            Log out
        </DropdownMenu.Item>
    </DropdownMenu.Content>
</DropdownMenu.Root>
