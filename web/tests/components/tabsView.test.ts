import { render, screen, fireEvent } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";

import TabsComponent from "../../src/lib/components/TabsView.svelte";

describe("TabsComponent", () => {
    it("renders all tabs", () => {
        render(TabsComponent);

        expect(screen.getByText("Edit")).toBeInTheDocument();
        expect(screen.getByText("Run")).toBeInTheDocument();
        expect(screen.getByText("Result")).toBeInTheDocument();
    });

    it("sets the default active tab to 'run'", () => {
        render(TabsComponent);

        const activeTab = screen.getByRole("tab", { name: "Run" });
        expect(activeTab).toHaveAttribute("aria-selected", "true");
    });

    it("changes active tab when clicked", async () => {
        render(TabsComponent);

        const editTab = screen.getByRole("tab", { name: "Edit" });

        await fireEvent.click(editTab);

        expect(editTab).toHaveAttribute("aria-selected", "true");
    });

    it("applies `px-6` class to root element", () => {
        const { container } = render(TabsComponent);

        const root = container.querySelector(".px-6");
        expect(root).toBeInTheDocument();
    });

    it("updates activeTab when a new tab is clicked", async () => {
        render(TabsComponent);

        const resultTab = screen.getByRole("tab", { name: "Result" });

        await fireEvent.click(resultTab);

        expect(resultTab).toHaveAttribute("aria-selected", "true");
    });
});
