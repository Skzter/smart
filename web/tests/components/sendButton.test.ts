import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import SendButton from "$lib/components/SendButton.svelte";
import { chat } from "$lib/shared.svelte";
import '@testing-library/jest-dom/vitest';

describe("SendButton", () => {
    let mockOnClick: () => void;

    beforeEach(() => {
        mockOnClick = vi.fn() as () => void;
        chat.isLoading = false;
    });

    it("renders a button", () => {
        render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });
        const button = screen.getByRole("button");
        expect(button).toBeInTheDocument();
    });

    it("displays Send icon", () => {
        const { container } = render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });

        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });

    it("calls onclick when clicked with valid input", async () => {
        const user = userEvent.setup();

        render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        await user.click(button);

        expect(mockOnClick).toHaveBeenCalledTimes(1);
    });

    it("is disabled when input is empty", () => {
        render(SendButton, {
            props: {
                input: "",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeDisabled();
    });

    it("is disabled when input contains only whitespace", () => {
        render(SendButton, {
            props: {
                input: "   ",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeDisabled();
    });

    it("is disabled when input contains only tabs and newlines", () => {
        render(SendButton, {
            props: {
                input: "\t\n  \n\t",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeDisabled();
    });

    it("is enabled when input has valid text", () => {
        render(SendButton, {
            props: {
                input: "Hello",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).not.toBeDisabled();
    });

    it("is enabled when input has text with leading/trailing spaces", () => {
        render(SendButton, {
            props: {
                input: "  Hello  ",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).not.toBeDisabled();
    });

    it("is disabled when chat.isLoading is true", () => {
        chat.isLoading = true;

        render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeDisabled();
    });

    it("is disabled when chat.isLoading is true even with valid input", () => {
        chat.isLoading = true;

        render(SendButton, {
            props: {
                input: "valid message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeDisabled();
    });

    it("is enabled when chat.isLoading is false and input is valid", () => {
        chat.isLoading = false;

        render(SendButton, {
            props: {
                input: "valid message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");
        expect(button).not.toBeDisabled();
    });

    it("does not call onclick when disabled due to empty input", async () => {
        const user = userEvent.setup();

        render(SendButton, {
            props: {
                input: "",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");

        await user.click(button);

        expect(mockOnClick).not.toHaveBeenCalled();
    });

    it("does not call onclick when disabled due to loading state", async () => {
        const user = userEvent.setup();
        chat.isLoading = true;

        render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");

        await user.click(button);

        expect(mockOnClick).not.toHaveBeenCalled();
    });

    it("can be clicked multiple times when enabled", async () => {
        const user = userEvent.setup();

        render(SendButton, {
            props: {
                input: "test message",
                onclick: mockOnClick,
            },
        });

        const button = screen.getByRole("button");

        await user.click(button);
        await user.click(button);
        await user.click(button);

        expect(mockOnClick).toHaveBeenCalledTimes(3);
    });
});