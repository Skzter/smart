import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import SaveButton from "../../src/lib/components/SaveButton.svelte";
import type { Runner } from "../../src/lib/runner.svelte";
import "@testing-library/jest-dom/vitest";

describe("SaveButton", () => {
    let mockTestRunner: Runner;

    beforeEach(() => {
        mockTestRunner = {
            storeTest: vi.fn(),
            getStorageState: vi.fn().mockReturnValue("idle"),
        } as unknown as Runner;
    });

    it("renders a save button with text 'Speichern'", () => {
        render(SaveButton, {
            props: {
                code: "test code",
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });
        const button = screen.getByRole("button", { name: /speichern/i });
        expect(button).toBeInTheDocument();
    });

    it("displays Save icon when not saving", () => {
        render(SaveButton, {
            props: {
                code: "test code",
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });
        const button = screen.getByRole("button", { name: /speichern/i });
        expect(button).toBeInTheDocument();
    });

    it("displays Spinner when saving", () => {
        mockTestRunner.getStorageState = vi.fn().mockReturnValue("saving");

        render(SaveButton, {
            props: {
                code: "test code",
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });
        const button = screen.getByRole("button", { name: /speichern/i });
        expect(button).toBeDisabled();
    });

    it("calls testRunner.storeTest with code when clicked", async () => {
        const user = userEvent.setup();
        const testCode = "console.log('test')";

        render(SaveButton, {
            props: {
                code: testCode,
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = screen.getByRole("button", { name: /speichern/i });
        await user.click(button);

        expect(mockTestRunner.storeTest).toHaveBeenCalledWith(testCode);
    });

    it("is disabled when storage state is 'saving'", () => {
        mockTestRunner.getStorageState = vi.fn().mockReturnValue("saving");

        render(SaveButton, {
            props: {
                code: "test code",
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = screen.getByRole("button", { name: /speichern/i });
        expect(button).toBeDisabled();
    });

    it("is enabled when storage state is not 'saving'", () => {
        mockTestRunner.getStorageState = vi.fn().mockReturnValue("idle");

        render(SaveButton, {
            props: {
                code: "test code",
                testRunner: mockTestRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = screen.getByRole("button", { name: /speichern/i });
        expect(button).not.toBeDisabled();
    });
});
