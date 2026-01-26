<script lang="ts">
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { buttonVariants } from "$lib/components/ui/button/index.js";
    import {
        CircleUserRound,
        EllipsisVertical,
        KeyRound,
        LogOut,
        Eye,
        EyeOff,
        Copy,
    } from "@lucide/svelte";
    import { auth } from "$lib/authService";
    import { apiToken, getToken } from "$lib/shared.svelte";
    import Button from "./ui/button/button.svelte";
    import { toast } from "svelte-sonner";
    const sidebar = Sidebar.useSidebar();

    let showTokenDialog = $state(false);
    let showTokenValue = $state(false);
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
        <DropdownMenu.Item onSelect={() => (showTokenDialog = true)}>
            <KeyRound />
            Token
        </DropdownMenu.Item>
        <DropdownMenu.Item onclick={auth.logout}>
            <LogOut />
            Log out
        </DropdownMenu.Item>
    </DropdownMenu.Content>
</DropdownMenu.Root>

<Dialog.Root bind:open={showTokenDialog}>
    <Dialog.Content class="sm:max-w-[500px]">
        <Dialog.Header>
            <Dialog.Title>API Token Settings</Dialog.Title>
            <Dialog.Description
                >Token sehen, kopieren und neu erstellen!</Dialog.Description
            >
        </Dialog.Header>
        <div class="space-y-4 py-4">
            <div class="space-y-2">
                <label class="text-sm font-medium" for="token">Dein Token</label
                >
                <div class="flex items-center gap-2">
                    <div
                        class="flex-1 rounded-md border bg-muted/50 px-3 py-2 font-mono text-sm"
                    >
                        {#if showTokenValue}
                            {apiToken.token || "Kein Token verfügbar"}
                        {:else}
                            {apiToken.token
                                ? "•".repeat(40)
                                : "Kein Token verfügbar"}
                        {/if}
                    </div>
                    <Button
                        variant="outline"
                        size="icon"
                        onclick={() => (showTokenValue = !showTokenValue)}
                        disabled={!apiToken.token}
                    >
                        {#if showTokenValue}
                            <EyeOff class="h-4 w-4" />
                        {:else}
                            <Eye class="h-4 w-4" />
                        {/if}
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        onclick={() => {
                            if (apiToken.token) {
                                navigator.clipboard.writeText(apiToken.token);
                                toast.success("Token kopiert!", {
                                    description:
                                        "Der Token wurde in die Zwischenablage kopiert.",
                                });
                            }
                        }}
                        disabled={!apiToken.token}
                    >
                        <Copy class="h-4 w-4" />
                    </Button>
                </div>
            </div>
            <Button class="w-full" onclick={async () => await getToken()}
                >Neuen Token generieren</Button
            >
        </div>
        <Dialog.Footer>
            <Dialog.Close class={buttonVariants({ variant: "outline" })}
                >Schließen</Dialog.Close
            >
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
