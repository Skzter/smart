import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock clipboard BEFORE importing the component
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.defineProperty(navigator, "clipboard", {
    value: {
        writeText: mockWriteText,
    },
    writable: true,
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

        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
    });

    it("displays Check icon after successful copy", async () => {
        const user = userEvent.setup();

        render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        // Check icon should be displayed after click
        const buttonAfter = screen.getByRole("button");
        expect(buttonAfter).toBeInTheDocument();
    });

    it("reverts to Copy icon after 3 seconds", async () => {
        vi.useFakeTimers();
        const user = userEvent.setup({ delay: null });

        render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        await vi.advanceTimersByTimeAsync(3000);

        const buttonAfter = screen.getByRole("button");
        expect(buttonAfter).toBeInTheDocument();
    });

    it("does not revert to Copy icon before 3 seconds", async () => {
        vi.useFakeTimers();
        const user = userEvent.setup({ delay: null });

        render(CopyButton, {
            props: {
                code: "test code",
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        await vi.advanceTimersByTimeAsync(2000);

        const buttonAfter = screen.getByRole("button");
        expect(buttonAfter).toBeInTheDocument();
    });
});
