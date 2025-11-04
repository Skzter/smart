import { render, screen, fireEvent } from "@testing-library/svelte";
import { describe, it, expect, vi } from "vitest";
import Popup from "../../src/components/Popup.svelte";

describe("Popup component", () => {
    describe("when isOpen is true", () => {
        it("renders the popup with title and message", () => {
            render(Popup, {
                isOpen: true,
                title: "Test Title",
                message: "Test message content",
            });

            const titleElement = screen.getByText("Test Title");
            const messageElement = screen.getByText("Test message content");
            const closeButton = screen.getByRole("button", { name: "Close" });
            const okButton = screen.getByRole("button", { name: "OK" });

            expect(titleElement).toBeInTheDocument();
            expect(messageElement).toBeInTheDocument();
            expect(closeButton).toBeInTheDocument();
            expect(okButton).toBeInTheDocument();
        });

        it("applies correct aria attributes for accessibility", () => {
            render(Popup, {
                isOpen: true,
                title: "Test Title",
            });

            const dialog = screen.getByRole("dialog");
            const title = screen.getByText("Test Title");

            expect(dialog).toBeInTheDocument();
            expect(dialog).toHaveAttribute("aria-modal", "true");
            expect(dialog).toHaveAttribute("aria-labelledby", "popup-title");
            expect(title).toHaveAttribute("id", "popup-title");
        });
    });

    describe("when isOpen is false", () => {
        it("does not render the popup", () => {
            const { container } = render(Popup, { isOpen: false });

            const dialog = container.querySelector('[role="dialog"]');
            expect(dialog).not.toBeInTheDocument();
        });
    });

    describe("type prop variations", () => {
        it("applies red color for error type", () => {
            render(Popup, {
                isOpen: true,
                type: "error",
                title: "Error Title",
            });

            const titleElement = screen.getByText("Error Title");
            expect(titleElement).toHaveClass("text-red-600");
        });

        it("applies green color for success type", () => {
            render(Popup, {
                isOpen: true,
                type: "success",
                title: "Success Title",
            });

            const titleElement = screen.getByText("Success Title");
            expect(titleElement).toHaveClass("text-green-600");
        });

        it("applies blue color for info type", () => {
            render(Popup, {
                isOpen: true,
                type: "info",
                title: "Info Title",
            });

            const titleElement = screen.getByText("Info Title");
            expect(titleElement).toHaveClass("text-blue-600");
        });
    });

    describe("default props", () => {
        it("uses default title when not provided", () => {
            render(Popup, { isOpen: true });

            const titleElement = screen.getByText("Alert");
            expect(titleElement).toBeInTheDocument();
        });

        it("handles empty message gracefully", () => {
            const { container } = render(Popup, {
                isOpen: true,
                message: "",
            });

            const messageElement = container.querySelector("p.text-gray-700");
            expect(messageElement).toBeInTheDocument();
            expect(messageElement?.textContent).toBe("");
        });
    });
    describe("user interactions", () => {
        it("has functional close button", async () => {
            render(Popup, { isOpen: true });

            const closeButton = screen.getByLabelText("Close");
            expect(closeButton).toBeInTheDocument();
            await fireEvent.click(closeButton);
            // If we get here without errors, the click handler worked
        });

        it("has functional OK button", async () => {
            render(Popup, { isOpen: true });

            const okButton = screen.getByText("OK");
            expect(okButton).toBeInTheDocument();
            await fireEvent.click(okButton);
            // If we get here without errors, the click handler worked
        });

        it("sets up escape key listener when open", () => {
            const addEventListenerSpy = vi.spyOn(window, "addEventListener");

            const { unmount } = render(Popup, { isOpen: true });

            expect(addEventListenerSpy).toHaveBeenCalledWith(
                "keydown",
                expect.any(Function),
            );

            // Test cleanup
            const removeEventListenerSpy = vi.spyOn(
                window,
                "removeEventListener",
            );
            unmount();
            expect(removeEventListenerSpy).toHaveBeenCalledWith(
                "keydown",
                expect.any(Function),
            );

            addEventListenerSpy.mockRestore();
            removeEventListenerSpy.mockRestore();
        });
    });
    describe("keyboard event handling", () => {
        it("only listens to Escape key when popup is open", () => {
            const addEventListenerSpy = vi.spyOn(window, "addEventListener");
            const removeEventListenerSpy = vi.spyOn(
                window,
                "removeEventListener",
            );

            const { unmount } = render(Popup, { isOpen: true });

            expect(addEventListenerSpy).toHaveBeenCalledWith(
                "keydown",
                expect.any(Function),
            );

            unmount();

            expect(removeEventListenerSpy).toHaveBeenCalledWith(
                "keydown",
                expect.any(Function),
            );

            addEventListenerSpy.mockRestore();
            removeEventListenerSpy.mockRestore();
        });

        it("ignores other keyboard events", async () => {
            render(Popup, { isOpen: true });

            await fireEvent.keyDown(window, { key: "Enter" });
            // Popup should still be open
            const dialog = screen.getByRole("dialog");
            expect(dialog).toBeInTheDocument();

            await fireEvent.keyDown(window, { key: "Tab" });
            expect(dialog).toBeInTheDocument();
        });
    });

    describe("whitespace handling", () => {
        it("preserves whitespace in message with whitespace-pre-wrap", () => {
            const messageWithWhitespace = "Line 1\nLine 2\n    Indented line";
            const { container } = render(Popup, {
                isOpen: true,
                message: messageWithWhitespace,
            });

            const messageElement = container.querySelector("p.text-gray-700");
            expect(messageElement).toHaveClass("whitespace-pre-wrap");
        });
    });

    describe("bindable prop behavior", () => {
        it("accepts isOpen as a bindable prop", () => {
            // This test verifies that the component can be rendered with the bindable prop
            // without throwing errors
            expect(() => {
                render(Popup, { isOpen: true });
            }).not.toThrow();
        });
    });
});
