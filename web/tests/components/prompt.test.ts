import { render, waitFor } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom/vitest";

// Mock API
const mockGetTemplate = vi.fn();
vi.mock("../../src/lib/api", () => ({
    getTemplate: () => mockGetTemplate(),
}));

// Mock toast
const mockToastError = vi.fn();
vi.mock("svelte-sonner", () => ({
    toast: {
        error: (...args: unknown[]) => mockToastError(...args),
    },
}));

import Prompt from "../../src/lib/components/Prompt.svelte";
import { chat, user } from "../../src/lib/shared.svelte";

describe("Prompt", () => {
    let inputValue: string;
    let onclickMock: () => void;

    beforeEach(() => {
        inputValue = "";
        onclickMock = vi.fn();
        chat.isLoading = false;
        user.id = "test-user-123"; // Set user ID so the effect runs
        mockGetTemplate.mockClear();
        mockToastError.mockClear();
    });

    afterEach(() => {
        vi.clearAllTimers();
    });

    it("renders the InputGroup.Root component", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const inputGroup = container.querySelector(".w-full");
        expect(inputGroup).toBeInTheDocument();
    });

    it("renders textarea with correct placeholder", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea");
        expect(textarea).toHaveAttribute("placeholder", "Send a message...");
    });

    it("renders textarea with correct initial rows", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea");
        expect(textarea).toHaveAttribute("rows", "1");
    });

    it("renders textarea with resize-none class", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea.resize-none");
        expect(textarea).toBeInTheDocument();
    });

    it("renders textarea with min-h-11 class", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea.min-h-11");
        expect(textarea).toBeInTheDocument();
    });

    it("loads template on mount successfully", async () => {
        const mockTemplate = `
Erzeuge Playwright-Tests via Autoplaywright für meine lokale Seite.
    Base-URL: localhost:8082
    Szenario: Nutzer nutzt die Suche.
    Ablauf: Zuerst gibt der Nutzer beim Reiseziel „Mallorca" ein.
    Assertions: Im Reiseziel-Feld steht „Mallorca".
    Testdaten/Setup: Reiseziel Mallorca
`;
        mockGetTemplate.mockResolvedValue(mockTemplate);

        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        await waitFor(() => {
            expect(mockGetTemplate).toHaveBeenCalled();
        });

        const textarea = container.querySelector("textarea");
        await waitFor(() => {
            expect(textarea?.value).toContain("Erzeuge Playwright-Tests");
        });
    });

    it("shows toast error when template loading fails", async () => {
        const errorMessage = "Failed to load template";
        mockGetTemplate.mockRejectedValue(new Error(errorMessage));

        render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        await waitFor(() => {
            expect(mockGetTemplate).toHaveBeenCalled();
        });

        await waitFor(() => {
            expect(mockToastError).toHaveBeenCalledWith(errorMessage, {
                description: "Das war wohl nichts mit dem Template.",
            });
        });
    });

    it("calls onclick when Enter key is pressed with non-empty input", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "Test message",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        expect(textarea).toBeInTheDocument();

        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).toHaveBeenCalled();
    });

    it("does not call onclick when Enter key is pressed with empty input", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).not.toHaveBeenCalled();
    });

    it("does not call onclick when Enter key is pressed with whitespace-only input", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "   ",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).not.toHaveBeenCalled();
    });

    it("does not call onclick when Shift+Enter is pressed", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "Test message",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Shift>}{Enter}{/Shift}");

        expect(onclickMock).not.toHaveBeenCalled();
    });

    it("prevents default behavior when Enter is pressed with valid input", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "Test message",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);

        const enterEvent = new KeyboardEvent("keydown", {
            key: "Enter",
            bubbles: true,
            cancelable: true,
        });
        const preventDefaultSpy = vi.spyOn(enterEvent, "preventDefault");

        textarea.dispatchEvent(enterEvent);

        expect(preventDefaultSpy).toHaveBeenCalled();
    });

    it("disables textarea when chat is loading", () => {
        chat.isLoading = true;

        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea");
        expect(textarea).toBeDisabled();
    });

    it("enables textarea when chat is not loading", () => {
        chat.isLoading = false;

        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea");
        expect(textarea).not.toBeDisabled();
    });

    it("binds input value correctly", async () => {
        const { container } = render(Prompt, {
            props: {
                input: "Initial value",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        expect(textarea.value).toBe("Initial value");
    });

    it("handles keyboard events other than Enter", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "Test",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("a");

        expect(onclickMock).not.toHaveBeenCalled();
    });

    it("trims input before checking if empty on Enter press", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "  Valid input  ",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).toHaveBeenCalled();
    });

    it("handles template with Mallorca scenario", async () => {
        const mallocaTemplate = `
Erzeuge Playwright-Tests via Autoplaywright für meine lokale Seite.
    Base-URL: localhost:8082
    Szenario: Nutzer nutzt die Suche.
    Ablauf: Zuerst gibt der Nutzer beim Reiseziel „Mallorca" ein.
    Assertions: Im Reiseziel-Feld steht „Mallorca".
    Testdaten/Setup: Reiseziel Mallorca
`;
        mockGetTemplate.mockResolvedValue(mallocaTemplate);

        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        await waitFor(() => {
            expect(mockGetTemplate).toHaveBeenCalled();
        });

        const textarea = container.querySelector("textarea");
        await waitFor(() => {
            expect(textarea?.value).toContain("Mallorca");
            expect(textarea?.value).toContain("localhost:8082");
            expect(textarea?.value).toContain("Autoplaywright");
        });
    });

    it("handles non-Error objects in catch block", async () => {
        mockGetTemplate.mockRejectedValue("String error");

        render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        await waitFor(() => {
            expect(mockGetTemplate).toHaveBeenCalled();
        });

        // Should still handle the error without crashing
        await waitFor(() => {
            expect(mockToastError).toHaveBeenCalled();
        });
    });

    it("maintains w-full class on InputGroup.Root", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const inputGroup = container.querySelector(".w-full");
        expect(inputGroup).toBeInTheDocument();
    });

    it("applies all required textarea classes", () => {
        const { container } = render(Prompt, {
            props: {
                input: inputValue,
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector("textarea");
        expect(textarea).toHaveClass("w-full");
        expect(textarea).toHaveClass("resize-none");
        expect(textarea).toHaveClass("min-h-11");
    });

    it("handles multiple Enter key presses", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "Test message",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).toHaveBeenCalledTimes(1);
    });

    it("handles edge case with only newline characters", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: "\n\n",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).not.toHaveBeenCalled();
    });

    it("handles edge case with mixed whitespace", async () => {
        const user = userEvent.setup();

        const { container } = render(Prompt, {
            props: {
                input: " \t \n ",
                onclick: onclickMock,
            },
        });

        const textarea = container.querySelector(
            "textarea",
        ) as HTMLTextAreaElement;
        await user.click(textarea);
        await user.keyboard("{Enter}");

        expect(onclickMock).not.toHaveBeenCalled();
    });
});
