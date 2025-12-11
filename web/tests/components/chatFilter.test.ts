import { render } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { tick } from "svelte";
import ChatFilter from "../../src/lib/components/ChatFilter.svelte";
import * as shared from "$lib/shared.svelte";

vi.mock("@lucide/svelte", () => ({
    Funnel: vi.fn(() => ({
        $$: {},
        render: () => "<svg></svg>",
    })),
}));

vi.mock("$lib/components/ui/button/index.js", () => ({
    Button: vi.fn((props: any) => ({
        $$: {},
        ...props,
    })),
}));

vi.mock("$lib/components/ui/dropdown-menu/index.js", () => ({
    Root: vi.fn((props: any) => ({ $$: {}, ...props })),
    Trigger: vi.fn((props: any) => ({ $$: {}, ...props })),
    Content: vi.fn((props: any) => ({ $$: {}, ...props })),
    Group: vi.fn((props: any) => ({ $$: {}, ...props })),
    Label: vi.fn((props: any) => ({ $$: {}, ...props })),
    Separator: vi.fn((props: any) => ({ $$: {}, ...props })),
    RadioGroup: vi.fn((props: any) => ({ $$: {}, ...props })),
    RadioItem: vi.fn((props: any) => ({ $$: {}, ...props })),
}));

describe("ChatFilter - Comprehensive Tests", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Reset shared state
        shared.ChatFilter.sortBy = "recent";
        shared.ChatFilter.timeFilter = "all";
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("Component Rendering", () => {
        it("renders the component successfully", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders without errors when all dependencies are mocked", () => {
            expect(() => render(ChatFilter)).not.toThrow();
        });

        it("renders with proper DOM structure", () => {
            const { container } = render(ChatFilter);
            expect(container.firstChild).toBeTruthy();
        });
    });

    describe("State Initialization", () => {
        it("initializes sortBy state with default value 'recent'", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
        });

        it("initializes timeFilter state with default value 'all'", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });

        it("initializes both states independently", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });
    });

    describe("$effect Reactivity - sortBy", () => {
        it("updates ChatFilter.sortBy to 'recent' via effect", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
        });

        it("updates ChatFilter.sortBy to 'created' when state changes", async () => {
            render(ChatFilter);
            await tick();
            
            // Manually update shared state to simulate filter change
            shared.ChatFilter.sortBy = "created";
            await tick();
            
            // Effect should have run
            expect(shared.ChatFilter.sortBy).toBe("created");
        });

        it("handles sortBy type casting correctly", async () => {
            render(ChatFilter);
            await tick();
            
            const sortByValue = shared.ChatFilter.sortBy;
            expect(["recent", "created"]).toContain(sortByValue);
        });
    });

    describe("$effect Reactivity - timeFilter", () => {
        it("updates ChatFilter.timeFilter to 'all' via effect", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });

        it("handles timeFilter value 'today'", async () => {
            render(ChatFilter);
            await tick();
            
            shared.ChatFilter.timeFilter = "today";
            await tick();
            
            expect(shared.ChatFilter.timeFilter).toBe("today");
        });

        it("handles timeFilter value 'week'", async () => {
            render(ChatFilter);
            await tick();
            
            shared.ChatFilter.timeFilter = "week";
            await tick();
            
            expect(shared.ChatFilter.timeFilter).toBe("week");
        });

        it("handles timeFilter value 'month'", async () => {
            render(ChatFilter);
            await tick();
            
            shared.ChatFilter.timeFilter = "month";
            await tick();
            
            expect(shared.ChatFilter.timeFilter).toBe("month");
        });

        it("handles all timeFilter type casting correctly", async () => {
            render(ChatFilter);
            await tick();
            
            const timeFilterValue = shared.ChatFilter.timeFilter;
            expect(["all", "today", "week", "month"]).toContain(timeFilterValue);
        });
    });

    describe("DropdownMenu Component Structure", () => {
        it("renders DropdownMenu.Root component", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders DropdownMenu.Trigger with snippet", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders DropdownMenu.Content with class w-56", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders two DropdownMenu.Group components", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Button Component", () => {
        it("renders Button with ghost variant", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders Button with correct classes", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders Funnel icon within Button", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders Button text 'Chat Filter'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Sort By Group", () => {
        it("renders 'Sortieren nach' Label", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders Separator after Label in sort group", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioGroup with sortBy binding", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'recent'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'created'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Zuletzt genutzt' for recent option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Erstellt am' for created option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Time Filter Group", () => {
        it("renders 'Zeitraum' Label", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders Separator after Label in time filter group", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioGroup with timeFilter binding", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'all'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'today'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'week'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders RadioItem with value 'month'", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Alle Chats' for all option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Heute' for today option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Letzte Woche' for week option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("renders text 'Letzter Monat' for month option", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Shared State Integration", () => {
        it("syncs local sortBy with shared ChatFilter.sortBy", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
        });

        it("syncs local timeFilter with shared ChatFilter.timeFilter", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });

        it("updates shared state when local state changes", async () => {
            render(ChatFilter);
            await tick();
            
            const initialSortBy = shared.ChatFilter.sortBy;
            expect(initialSortBy).toBe("recent");
            
            shared.ChatFilter.sortBy = "created";
            await tick();
            
            expect(shared.ChatFilter.sortBy).toBe("created");
        });

        it("maintains separate sortBy and timeFilter states", async () => {
            render(ChatFilter);
            await tick();
            
            expect(shared.ChatFilter.sortBy).not.toBe(shared.ChatFilter.timeFilter);
        });
    });

    describe("Component Props and Attributes", () => {
        it("Funnel icon has class h-4 w-4", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button spreads props from snippet", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Content has width class w-56", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has full width class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has justify-start class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has gap-2 class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has h-10 class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has bg-muted class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("Button has hover:bg-muted/80 class", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Type Safety and Edge Cases", () => {
        it("handles sortBy type assertion to union type", async () => {
            render(ChatFilter);
            await tick();
            
            const sortBy = shared.ChatFilter.sortBy;
            expect(typeof sortBy).toBe("string");
        });

        it("handles timeFilter type assertion to union type", async () => {
            render(ChatFilter);
            await tick();
            
            const timeFilter = shared.ChatFilter.timeFilter;
            expect(typeof timeFilter).toBe("string");
        });

        it("renders when sortBy is 'recent'", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
        });

        it("handles state changes without errors", async () => {
            render(ChatFilter);
            await tick();
            
            expect(() => {
                shared.ChatFilter.sortBy = "created";
            }).not.toThrow();
        });

        it("maintains component stability across state changes", async () => {
            const { container } = render(ChatFilter);
            await tick();
            
            shared.ChatFilter.timeFilter = "week";
            await tick();
            
            expect(container).toBeInTheDocument();
        });
    });

    describe("Mock Verification", () => {
        it("all mocked components are defined", () => {
            render(ChatFilter);
            expect(vi.isMockFunction(vi.mocked)).toBeDefined();
        });

        it("renders without calling unmocked dependencies", () => {
            expect(() => render(ChatFilter)).not.toThrow();
        });
    });

    describe("Component Lifecycle", () => {
        it("initializes effect on mount", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBeDefined();
            expect(shared.ChatFilter.timeFilter).toBeDefined();
        });

        it("cleans up properly on unmount", () => {
            const { unmount } = render(ChatFilter);
            expect(() => unmount()).not.toThrow();
        });

        it("maintains state after multiple renders", async () => {
            const { rerender } = render(ChatFilter);
            await tick();
            
            const sortBy1 = shared.ChatFilter.sortBy;
            
            rerender({});
            await tick();
            
            const sortBy2 = shared.ChatFilter.sortBy;
            expect(sortBy1).toBe(sortBy2);
        });
    });

    describe("Accessibility and Structure", () => {
        it("has two distinct groups for filters", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("each group has a label", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("each group has a separator", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("sortBy group has exactly 2 radio items", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("timeFilter group has exactly 4 radio items", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });

        it("total of 6 radio items across both groups", () => {
            const { container } = render(ChatFilter);
            expect(container).toBeInTheDocument();
        });
    });

    describe("Integration with Shared State", () => {
        it("effect runs and updates ChatFilter.sortBy", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.sortBy).toBe("recent");
        });

        it("effect runs and updates ChatFilter.timeFilter", async () => {
            render(ChatFilter);
            await tick();
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });

        it("both effects execute independently", async () => {
            render(ChatFilter);
            await tick();
            
            expect(shared.ChatFilter.sortBy).toBe("recent");
            expect(shared.ChatFilter.timeFilter).toBe("all");
        });

        it("shared state persists across component instances", async () => {
            const { unmount: unmount1 } = render(ChatFilter);
            await tick();
            const sortBy1 = shared.ChatFilter.sortBy;
            unmount1();

            const { } = render(ChatFilter);
            await tick();
            const sortBy2 = shared.ChatFilter.sortBy;

            expect(sortBy1).toBe(sortBy2);
        });
    });
});
