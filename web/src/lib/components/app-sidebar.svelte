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

	/* ---------------- Sidebar Data ---------------- */
	const sidebarData = {
		user: {
			name: "shadcn",
			email: "m@example.com",
			avatar: "/avatars/shadcn.jpg"
		},
		teams: [
			{ name: "Acme Inc", logo: GalleryVerticalEndIcon, plan: "Enterprise" },
			{ name: "Acme Corp.", logo: AudioWaveformIcon, plan: "Startup" },
			{ name: "Evil Corp.", logo: CommandIcon, plan: "Free" }
		],
		navMain: [
			{
				title: "Playground",
				url: "#",
				icon: SquareTerminalIcon,
				isActive: true,
				items: [
					{ title: "History", url: "#" },
					{ title: "Starred", url: "#" },
					{ title: "Settings", url: "#" }
				]
			},
			{
				title: "Models",
				url: "#",
				icon: BotIcon,
				items: [
					{ title: "Genesis", url: "#" },
					{ title: "Explorer", url: "#" },
					{ title: "Quantum", url: "#" }
				]
			},
			{
				title: "Documentation",
				url: "#",
				icon: BookOpenIcon,
				items: [
					{ title: "Introduction", url: "#" },
					{ title: "Get Started", url: "#" },
					{ title: "Tutorials", url: "#" },
					{ title: "Changelog", url: "#" }
				]
			},
			{
				title: "Settings",
				url: "#",
				icon: Settings2Icon,
				items: [
					{ title: "General", url: "#" },
					{ title: "Team", url: "#" },
					{ title: "Billing", url: "#" },
					{ title: "Limits", url: "#" }
				]
			}
		],
		projects: [
			{ name: "Design Engineering", url: "#", icon: FrameIcon },
			{ name: "Sales & Marketing", url: "#", icon: ChartPieIcon },
			{ name: "Travel", url: "#", icon: MapIcon }
		]
	};

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
		<NavMain items={sidebarData.navMain} />
		<NavHistory />
	</Sidebar.Content>

	<Sidebar.Footer>
		<NavUser user={sidebarData.user} />
	</Sidebar.Footer>

	<Sidebar.Rail />
</Sidebar.Root>
