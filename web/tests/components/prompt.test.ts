import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { expect, test } from "vitest";
import { vi } from "vitest";
import Prompt from "../../src/components/Prompt.svelte";

describe("Prompt Component", () => {
    test("renders textarea and button", () => {
        render(Prompt);

        expect(screen.getByPlaceholderText("Prompt")).toBeInTheDocument();
        expect(screen.getByText("Send")).toBeInTheDocument();
    });

    test("textarea has correct classes", () => {
        render(Prompt);
        const textarea = screen.getByPlaceholderText("Prompt");

        expect(textarea).toHaveClass(
            "w-9/10",
            "resize-none",
            "overflow-y-auto",
        );
    });

    test("button has correct classes", () => {
        render(Prompt);
        const button = screen.getByText("Send");

        // Flowbite button is nested, so we check the parent button element
        const buttonElement = button.closest("button");
        expect(buttonElement).toHaveClass("w-1/10", "h-1/3");
    });

    test("container has correct classes", () => {
        render(Prompt);
        const container = screen.getByPlaceholderText("Prompt").parentElement;

        expect(container).toHaveClass(
            "flex",
            "flex-row",
            "w-screen",
            "items-center",
            "bg-white",
            "p-4",
            "border-t",
            "gap-2",
        );
    });

    test("button is disabled when input is empty", () => {
        render(Prompt);
        expect(screen.getByText("Send")).toBeDisabled();
    });

    test("button is enabled when input has value", () => {
        render(Prompt, { input: "Test" });
        expect(screen.getByText("Send")).not.toBeDisabled();
    });

    test("pressing Enter calls onclick", async () => {
        const user = userEvent.setup();
        const mockFn = vi.fn();
        render(Prompt, { onclick: mockFn, input: "Test" });

        const textarea = screen.getByPlaceholderText("Prompt");
        textarea.focus();
        await user.keyboard("[Enter]");

        expect(mockFn).toHaveBeenCalled();
    });

    test("Shift+Enter doesn't call onclick", async () => {
        const user = userEvent.setup();
        const mockFn = vi.fn();
        render(Prompt, { onclick: mockFn, input: "Test" });

        const textarea = screen.getByPlaceholderText("Prompt");
        textarea.focus();
        await user.keyboard("[ShiftLeft>][Enter][/ShiftLeft]");

        expect(mockFn).not.toHaveBeenCalled();
    });

    test("doesn't call onclick when input is empty", async () => {
        const user = userEvent.setup();
        const mockFn = vi.fn();
        render(Prompt, { onclick: mockFn, input: "" });

        const textarea = screen.getByPlaceholderText("Prompt");
        textarea.focus();
        await user.keyboard("[Enter]");

        expect(mockFn).not.toHaveBeenCalled();
    });

    test("doesn't call onclick when input is only whitespace", async () => {
        const user = userEvent.setup();
        const mockFn = vi.fn();
        render(Prompt, { onclick: mockFn, input: "   " });

        const textarea = screen.getByPlaceholderText("Prompt");
        textarea.focus();
        await user.keyboard("[Enter]");

        expect(mockFn).not.toHaveBeenCalled();
    });
});
