<script lang="ts">
	import * as Collapsible from "$lib/components/ui/collapsible/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import NewChat from "@lucide/svelte/icons/plus";
	import Calendar from "@lucide/svelte/icons/calendar";
	import { Button } from "$lib/components/ui/button/index.js";
	import ChevronRightIcon from "@lucide/svelte/icons/chevron-right";
	import * as Popover from "$lib/components/ui/popover/index.js";
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
</script>

<!-- Add a top button above the Platform group -->
<Sidebar.Menu>
	<Sidebar.MenuItem class="px-3 py-2">
		<!-- Use the Button UI component inside a MenuItem; avoid nested buttons by not using Sidebar.MenuButton here -->
		<div class="px-1">
			<Button variant="outline" size="lg" class="w-full justify-start gap-3 h-14">
				<!-- static icon usage (simple) -->
				<NewChat class="size-5" />
				New Chat
			</Button>
		</div>
	</Sidebar.MenuItem>

	<Sidebar.MenuItem class="px-3 py-2">
		<!-- Use the Button UI component inside a MenuItem; avoid nested buttons by not using Sidebar.MenuButton here -->
		<div class="px-1">
			<Button variant="outline" size="lg" class="w-full justify-start gap-3 h-14">
				<!-- static icon usage (simple) -->
				<Calendar class="size-5" />
				Calendar
			</Button>
		</div>
	</Sidebar.MenuItem>

</Sidebar.Menu>

<Sidebar.Group>
	<Sidebar.GroupLabel>Platform</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each items as item (item.title)}
			<Collapsible.Root open={item.isActive} class="group/collapsible">
				{#snippet child({ props })}
					<Sidebar.MenuItem {...props}>
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<Sidebar.MenuButton {...props} tooltipContent={item.title}>
									{#if item.icon}
										<item.icon />
									{/if}
									<span>{item.title}</span>
									<ChevronRightIcon
										class="ms-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
									/>
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
