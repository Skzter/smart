<script lang="ts">
	import * as Collapsible from "$lib/components/ui/collapsible/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import NewChat from "@lucide/svelte/icons/plus";
	import { toast } from "svelte-sonner";
	import Calendar from "@lucide/svelte/icons/calendar";
	import Filter from "@lucide/svelte/icons/filter";
	import { Button } from "$lib/components/ui/button/index.js";
	import { createChat } from "$lib/stores/chats";
	import ChevronRightIcon from "@lucide/svelte/icons/chevron-right";
	import * as Popover from "$lib/components/ui/popover/index.js";
	import { useSidebar } from "$lib/components/ui/sidebar/index.js";
	import { buttonVariants } from "$lib/components/ui/button/index.js";

	let {
		items,
	}: {
		items: {
			title: string;
			url: string;
			// this should be `Component` after @lucide/svelte updates types
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			icon?: any;
			isActive?: boolean;
			items?: {
				title: string;
				url: string;
			}[];
		}[];
	} = $props();
    
	// Change this to any lucide icon to swap the icon used for the Login button
	let loginIcon: any = NewChat;
	const sidebar = useSidebar();
</script>

<!-- Add a top button above the Platform group -->
<Sidebar.Menu>
	<Sidebar.MenuItem class="px-3 py-2" onclick={() => { createChat(); toast.success('New Chat started!'); }}>
		<!-- Use the Button UI component inside a MenuItem; avoid nested buttons by not using Sidebar.MenuButton here -->
		<div class="px-1">
			<Button
				variant="outline"
				size={sidebar.state === "collapsed" ? "icon" : "lg"}
				class={sidebar.state === "collapsed" ? "w-full justify-center gap-0" : "w-full justify-start gap-3 h-13"}
			>
				<NewChat class="size-5" />
				{#if sidebar.state !== "collapsed"}
					New Chat
				{/if}
			</Button>
		</div>
	</Sidebar.MenuItem>

	<Sidebar.MenuItem class="px-3 py-2">
		<div class="px-1">
			<Button
				variant="outline"
				size={sidebar.state === "collapsed" ? "icon" : "lg"}
				class={sidebar.state === "collapsed" ? "w-full justify-center gap-0 " : "w-full justify-start gap-3 h-9"}
			>
				<Calendar class="size-5" />
				{#if sidebar.state !== "collapsed"}
					Calendar
				{/if}
			</Button>
		</div>
	</Sidebar.MenuItem>

	<Sidebar.MenuItem class="px-3 py-2">
		<div class="px-1">
			<Button
				variant="outline"
				size={sidebar.state === "collapsed" ? "icon" : "lg"}
				class={sidebar.state === "collapsed" ? "w-full justify-center gap-0" : "w-full justify-start gap-3 h-9"}
			>
				<Filter class="size-5" />
				{#if sidebar.state !== "collapsed"}
					Filter
				{/if}
			</Button>
		</div>
	</Sidebar.MenuItem>

</Sidebar.Menu>

<Sidebar.Group>
	<Sidebar.GroupLabel>History</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each items as item (item.title)}
			<Collapsible.Root open={item.isActive} class="group/collapsible">
				{#snippet child({ props }: { props: Record<string, unknown> })}
					<Sidebar.MenuItem {...props}>
						<Collapsible.Trigger>
							{#snippet child({ props }: { props: Record<string, unknown> })}
								<Sidebar.MenuButton {...props} tooltipContent={item.title}>
									{#if item.icon}
										<item.icon />
									{/if}
									<span>{item.title}</span>
									<ChevronRightIcon class="ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"/>
								</Sidebar.MenuButton>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content>
							<Sidebar.MenuSub>
								{#each item.items ?? [] as subItem (subItem.title)}
									<Sidebar.MenuSubItem>
										<Sidebar.MenuSubButton>
											{#snippet child({ props })}
												<a href={subItem.url} {...props}>
													<span>{subItem.title}</span>
												</a>
											{/snippet}
										</Sidebar.MenuSubButton>
									</Sidebar.MenuSubItem>
								{/each}
							</Sidebar.MenuSub>
						</Collapsible.Content>
					</Sidebar.MenuItem>
				{/snippet}
			</Collapsible.Root>
		{/each}
	</Sidebar.Menu>
</Sidebar.Group>
