import { render } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";

import BrowserView from "../../src/lib/components/BrowserView.svelte";

// Mock runner with videoUrl, screenshotUrl, getCurTest and required methods
function createMockRunner(
    overrides: {
        videoUrl?: string | null;
        screenshotUrl?: string | null;
        getCurTest?: () => string;
    } = {},
) {
    const {
        videoUrl = null,
        screenshotUrl = null,
        getCurTest = () => "",
    } = overrides;
    return {
        videoUrl,
        screenshotUrl,
        getCurTest: vi.fn(getCurTest),
        model: {
            summary: { status: "idle" as const },
            steps: [],
        },
        fetchMediaUrl: vi.fn(),
        clearVideoUrl: vi.fn(),
    };
}

describe("BrowserView", () => {
    beforeEach(() => {
        vi.spyOn(window, "addEventListener");
        vi.spyOn(window, "removeEventListener");
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("renders the preview header", () => {
        const { container } = render(BrowserView, {
            props: { runner: createMockRunner() },
        });

        const header = container.querySelector(".border-b.bg-muted\\/50");
        expect(header).toBeInTheDocument();
        expect(header?.textContent).toContain("Vorschau");
    });

    it("centers the content", () => {
        const { container } = render(BrowserView, {
            props: { runner: createMockRunner() },
        });

        const centeredContent = container.querySelector(
            ".flex.justify-center.items-center.flex-1.overflow-hidden",
        );
        expect(centeredContent).toBeInTheDocument();
    });

    it("renders a video element with controls when videoUrl is set", () => {
        const { container } = render(BrowserView, {
            props: {
                runner: createMockRunner({
                    videoUrl: "http://example.com/video.mp4",
                }),
            },
        });

        const video = container.querySelector("video");
        expect(video).toBeInTheDocument();
        expect(video).toHaveAttribute("controls");
    });

    it("shows placeholder text when videoUrl is null", () => {
        const { container } = render(BrowserView, {
            props: { runner: createMockRunner({ videoUrl: null }) },
        });

        const video = container.querySelector("video");
        expect(video).not.toBeInTheDocument();
        expect(container.textContent).toContain(
            "Video wird nach Fehlschlag angezeigt",
        );
    });

    it("has a container with h-full class for proper sizing", () => {
        const { container } = render(BrowserView, {
            props: { runner: createMockRunner() },
        });

        const mainContainer = container.querySelector("#container");
        expect(mainContainer).toBeInTheDocument();
        expect(mainContainer).toHaveClass("h-full");
    });

    it("adds resize event listener on mount", () => {
        render(BrowserView, {
            props: { runner: createMockRunner() },
        });

        expect(window.addEventListener).toHaveBeenCalledWith(
            "resize",
            expect.any(Function),
        );
    });

    it("removes resize event listener on unmount", () => {
        const { unmount } = render(BrowserView, {
            props: { runner: createMockRunner() },
        });
        unmount();

        expect(window.removeEventListener).toHaveBeenCalledWith(
            "resize",
            expect.any(Function),
        );
    });

    it("calculates video size based on parent height", async () => {
        const { container } = render(BrowserView, {
            props: {
                runner: createMockRunner({
                    videoUrl: "http://example.com/video.mp4",
                }),
            },
        });

        const mainContainer = container.querySelector(
            "#container",
        ) as HTMLElement;
        const video = container.querySelector("video") as HTMLVideoElement;

        // Mock parent element with specific dimensions
        Object.defineProperty(mainContainer, "parentElement", {
            value: {
                clientHeight: 500,
                clientWidth: 800,
            },
            writable: true,
        });

        // Trigger resize event to recalculate
        window.dispatchEvent(new Event("resize"));

        // Wait for the update
        await new Promise((resolve) => setTimeout(resolve, 0));

        // Expected: availableHeight = 500 - 50 = 450
        // calculatedWidth = 450 * (16/9) = 800
        expect(video.style.height).toBe("450px");
        expect(video.style.width).toBe("800px");
    });

    it("does not set video size if parent element is missing", async () => {
        const { container } = render(BrowserView, {
            props: {
                runner: createMockRunner({
                    videoUrl: "http://example.com/video.mp4",
                }),
            },
        });

        const mainContainer = container.querySelector(
            "#container",
        ) as HTMLElement;
        const video = container.querySelector("video") as HTMLVideoElement;

        // Mock no parent element
        Object.defineProperty(mainContainer, "parentElement", {
            value: null,
            writable: true,
        });

        // Clear any styles that might have been set
        video.style.width = "";
        video.style.height = "";

        // Trigger resize event
        window.dispatchEvent(new Event("resize"));

        await new Promise((resolve) => setTimeout(resolve, 0));

        // Video style should remain unchanged (empty or not set)
        expect(video.style.width).toBe("");
        expect(video.style.height).toBe("");
    });
});
