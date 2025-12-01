<script lang="ts">
	/* ---------------- Icons ---------------- */
	import AudioWaveformIcon from "@lucide/svelte/icons/audio-waveform";
	import BookOpenIcon from "@lucide/svelte/icons/book-open";
	import BotIcon from "@lucide/svelte/icons/bot";
	import ChartPieIcon from "@lucide/svelte/icons/chart-pie";
	import CommandIcon from "@lucide/svelte/icons/command";
	import FrameIcon from "@lucide/svelte/icons/frame";
	import GalleryVerticalEndIcon from "@lucide/svelte/icons/gallery-vertical-end";
	import MapIcon from "@lucide/svelte/icons/map";
	import Settings2Icon from "@lucide/svelte/icons/settings-2";
	import SquareTerminalIcon from "@lucide/svelte/icons/square-terminal";

	/* ---------------- Components ---------------- */
	import NavMain from "./nav-main.svelte";
	import NavHistory from "./nav-history.svelte";
	import NavUser from "./nav-user.svelte";
	import TeamSwitcher from "./team-switcher.svelte";
	import * as Sidebar from "$lib/components/ui/sidebar";
	import type { ComponentProps } from "svelte";
    import { Title } from "./ui/sheet";
	import { chats } from "$lib/stores/chats";

	/* ---------------- Sidebar Data ---------------- */
	const sidebarData = {
		user: {
			name: "Johannes",
			email: "johannes@check24.com",
			avatar: "/avatars/shadcn.jpg"
		},
		teams: [
			{ name: "Check24", logo: GalleryVerticalEndIcon, plan: "Frontend Tests" },
		],
	};

	let navMainItems = $derived([
		{
			title: "Sidebar",
			url: "#",
			isActive: true,
			items: $chats.map(chat => ({ title: chat.title, url: "#" }))
		},
		{
			title: "Login",
			url: "#",
			isActive: true,
			items: [
				{ title: "Homepage Login", url: "#",},

			]
		},
		{
			title: "Vacation",
			url: "#",
			isActive: true,
			items: [
				{ title: "Pauschalreisen", url: "#",},

			]
		}
	]);

	/* ---------------- Props ---------------- */
	let {
		ref = $bindable(null),
		collapsible = "icon",
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();
</script>

<Sidebar.Root {collapsible} {...restProps}>
	<Sidebar.Header>
		<TeamSwitcher teams={sidebarData.teams} />
	</Sidebar.Header>

	<Sidebar.Content>
		<NavMain items={navMainItems} />
		<NavHistory />
	</Sidebar.Content>

	<Sidebar.Footer>
		<NavUser user={sidebarData.user} />
	</Sidebar.Footer>

	<Sidebar.Rail />
</Sidebar.Root>
