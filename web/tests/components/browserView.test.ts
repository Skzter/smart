import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";

import BrowserView from "../../src/lib/components/BrowserView.svelte";

describe("BrowserView", () => {
    it("renders the preview header", () => {
        const { container } = render(BrowserView);

        const header = container.querySelector(".border-b.bg-muted\\/50");
        expect(header).toBeInTheDocument();
        expect(header?.textContent?.trim()).toBe("Vorschau");
    });

    it("renders the monitor icon", () => {
        const { container } = render(BrowserView);

        const icon = container.querySelector("svg");
        expect(icon).toBeInTheDocument();
    });

    it("displays browser preview text", () => {
        const { container } = render(BrowserView);

        const previewText = container.querySelector("p.text-sm");
        expect(previewText?.textContent).toBe("Browser Vorschau");
    });

    it("has correct layout structure", () => {
        const { container } = render(BrowserView);

        const mainContainer = container.querySelector(".flex.flex-col");
        const contentArea = container.querySelector(".flex-1.overflow-auto");

        expect(mainContainer).toBeInTheDocument();
        expect(contentArea).toBeInTheDocument();
    });

    it("centers the content", () => {
        const { container } = render(BrowserView);

        const centeredContent = container.querySelector(
            ".flex.items-center.justify-center",
        );
        expect(centeredContent).toBeInTheDocument();
    });
});
