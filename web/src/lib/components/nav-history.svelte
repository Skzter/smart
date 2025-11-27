<script lang="ts">
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import type { ComponentProps } from "svelte";
  import { chats, selectedChat, selectChat, createChat } from "$lib/stores/chats";
    import MessageSquareIcon from "@lucide/svelte/icons/message-square";

  let {
    ref = $bindable(null),
    class: className,
    ...restProps
  }: ComponentProps<typeof Sidebar.Group> = $props();

  // We'll use the store's $ syntax directly in the template ($chats, $selectedChat)
</script>

<!-- Show a single History icon when the sidebar is collapsed (hidden by default, visible in collapsed state) -->
<Sidebar.Menu>
  <Sidebar.MenuItem class="px-3 py-2 hidden group-data-[collapsible=icon]:flex">
    <Sidebar.MenuButton
      tooltipContent="History"
      class="w-full justify-center"
      onclick={() => {
        // Select the most recent chat, or create a new one
        // Use the exported `chats` store to get the first item (latest)
        // We have to subscribe since we are in a callback; use a short-lived subscription
        let selected = null;
        const unsubscribe = chats.subscribe(($c) => {
          selected = $c[0];
        });
        unsubscribe();
        if (selected) {
          selectChat(selected.id);
        } else {
          // create a new chat and select it
          createChat();
        }
      }}
    >
      <MessageSquareIcon />
    </Sidebar.MenuButton>
  </Sidebar.MenuItem>
</Sidebar.Menu>

<!-- Full history group visible when the sidebar is expanded -->
<Sidebar.Group class={`${className ?? ""} group-data-[collapsible=icon]:hidden`} {...restProps}>
  <Sidebar.GroupLabel>Saved Tests</Sidebar.GroupLabel>
  <Sidebar.Menu>
    <!-- Leer als Platzhalter -->
  </Sidebar.Menu>
</Sidebar.Group>
