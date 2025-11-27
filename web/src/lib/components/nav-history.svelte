<script lang="ts">
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import type { ComponentProps } from "svelte";
  import { chats, selectedChat, selectChat, createChat } from "$lib/stores/chats";
  import BookmarkIcon from "@lucide/svelte/icons/bookmark";

  let {
    ref = $bindable(null),
    class: className,
    ...restProps
  }: ComponentProps<typeof Sidebar.Group> = $props();
</script>

