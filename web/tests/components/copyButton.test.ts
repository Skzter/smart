import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock clipboard BEFORE importing the component
const mockWriteText = vi.fn();

Object.defineProperty(navigator, "clipboard", {
    value: {
        writeText: mockWriteText,
    },
    writable: false,
    configurable: true,
});

import CopyButton from "../../src/lib/components/CopyButton.svelte";

describe("CopyButton", () => {
    beforeEach(() => {
        mockWriteText.mockClear();
        mockWriteText.mockResolvedValue(undefined);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it("renders a button", () => {
        render(CopyButton, {
            props: {
                code: "test code",
            },
        });
        const button = screen.getByRole("button");
        expect(button).toBeInTheDocument();
    });

    it("displays Copy icon initially", () => {
        const { container } = render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const copyIcon = container.querySelector("svg.lucide-copy");
        expect(copyIcon).toBeInTheDocument();
    });

    it("displays Check icon after successful copy", async () => {
        const user = userEvent.setup();

        const { container } = render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        const checkIcon = container.querySelector("svg.lucide-check");
        expect(checkIcon).toBeInTheDocument();
    });

    it("reverts to Copy icon after 3 seconds", async () => {
        vi.useFakeTimers();
        const user = userEvent.setup({ delay: null });

        const { container } = render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        // Check icon should be visible
        expect(container.querySelector("svg.lucide-check")).toBeInTheDocument();

        await vi.advanceTimersByTimeAsync(3000);

        // Copy icon should be back
        expect(container.querySelector("svg.lucide-copy")).toBeInTheDocument();
    });

    it("does not revert to Copy icon before 3 seconds", async () => {
        vi.useFakeTimers();
        const user = userEvent.setup({ delay: null });

        const { container } = render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        await vi.advanceTimersByTimeAsync(2000);

        // Check icon should still be visible
        expect(container.querySelector("svg.lucide-check")).toBeInTheDocument();
    });
});
