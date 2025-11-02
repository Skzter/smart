import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { expect, test } from "vitest";
import { vi } from "vitest";
import Prompt from "../../src/components/Prompt.svelte";

vi.mock("../../src/lib/Api.ts", () => ({
    getTemplate: vi.fn(),
}));

import { getTemplate } from "../../src/lib/Api.ts";

beforeEach(() => {
    vi.clearAllMocks();
});

if (!Element.prototype.animate) {
    Element.prototype.animate = () => ({
        cancel: () => {},
        finish: () => {},
        play: () => {},
        pause: () => {},
        reverse: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        finished: Promise.resolve(),
        playState: "finished",
    });
}
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

    test("clicking Template calls getTemplate and fills textarea on success", async () => {
        const user = userEvent.setup();
        const templateString = "This is a loaded template";
        getTemplate.mockResolvedValue({
            status: 200,
            data: { template: templateString },
        });

        render(Prompt, { input: "" });

        const templateButton = screen.getByText("Template");
        await user.click(templateButton);

        await waitFor(() => {
            expect(screen.getByPlaceholderText("Prompt")).toHaveValue(
                templateString,
            );
        });

        expect(getTemplate).toHaveBeenCalledTimes(1);
        expect(getTemplate).toHaveBeenCalledWith("/template");
    });

    test("clicking Template shows alert on non-200 response and doesn't change textarea", async () => {
        const user = userEvent.setup();
        getTemplate.mockRejectedValue({ status: 404, data: {} });

        render(Prompt, { input: "" });

        const templateButton = screen.getByText("Template");
        await user.click(templateButton);

        await waitFor(() => {
            expect(screen.getByText("NO TEMPLATE FOUND!")).toBeInTheDocument();
            expect(screen.getByText("Template Error")).toBeInTheDocument();
        });

        expect(screen.getByPlaceholderText("Prompt")).toHaveValue("");
        expect(getTemplate).toHaveBeenCalledTimes(1);
        expect(getTemplate).toHaveBeenCalledWith("/template");
    });
});
