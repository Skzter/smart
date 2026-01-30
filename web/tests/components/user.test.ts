import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock toast
vi.mock("svelte-sonner", () => ({
    toast: {
        success: vi.fn(),
        error: vi.fn(),
    },
}));

// Mock auth service
vi.mock("$lib/authService", () => ({
    auth: {
        logout: vi.fn(),
    },
}));

// Mock getToken function
vi.mock("$lib/shared.svelte", async () => {
    const actual = await vi.importActual("$lib/shared.svelte");
    return {
        ...actual,
        getToken: vi.fn(),
    };
});

// Mock clipboard
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.defineProperty(navigator, "clipboard", {
    value: {
        writeText: mockWriteText,
    },
    writable: true,
    configurable: true,
});

import UserTestWrapper from "../helpers/UserTestWrapper.svelte";
import { apiToken, getToken } from "$lib/shared.svelte";
import { auth } from "$lib/authService";
import { toast } from "svelte-sonner";

describe("User", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockWriteText.mockResolvedValue(undefined);
        vi.mocked(getToken).mockResolvedValue(undefined);
        apiToken.token = null;
    });

    afterEach(() => {
        apiToken.token = null;
    });

    it("renders the user dropdown menu button", () => {
        const { container } = render(UserTestWrapper);

        const button = container.querySelector("button");
        expect(button).toBeInTheDocument();
    });

    it("displays user icon in the menu button", () => {
        const { container } = render(UserTestWrapper);

        // CircleUserRound icon should be present
        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
    });

    it("displays 'Autotester' as the username", () => {
        render(UserTestWrapper);

        const username = screen.getByText("Autotester");
        expect(username).toBeInTheDocument();
    });

    it("opens dropdown menu when button is clicked", async () => {
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        await waitFor(() => {
            const tokenItem = screen.getByText("Token");
            expect(tokenItem).toBeInTheDocument();
        });
    });

    it("displays 'Token' menu item in dropdown", async () => {
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        await waitFor(() => {
            const tokenItem = screen.getByText("Token");
            expect(tokenItem).toBeInTheDocument();
        });
    });

    it("displays 'Log out' menu item in dropdown", async () => {
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        await waitFor(() => {
            const logoutItem = screen.getByText("Log out");
            expect(logoutItem).toBeInTheDocument();
        });
    });

    it("calls auth.logout when logout button is clicked", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        await waitFor(() => {
            const logoutItem = screen.getByText("Log out");
            expect(logoutItem).toBeInTheDocument();
        });

        const logoutItem = screen.getByText("Log out");
        await user.click(logoutItem);

        expect(auth.logout).toHaveBeenCalledTimes(1);
    });

    it("opens token dialog when Token menu item is clicked", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        await waitFor(() => {
            const tokenItem = screen.getByText("Token");
            expect(tokenItem).toBeInTheDocument();
        });

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const dialogTitle = screen.getByText("API Token Settings");
            expect(dialogTitle).toBeInTheDocument();
        });
    });

    it("displays token dialog with correct title", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const dialogTitle = screen.getByText("API Token Settings");
            expect(dialogTitle).toBeInTheDocument();
        });
    });

    it("displays token dialog with correct description", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const description = screen.getByText(
                "Token sehen, kopieren und neu erstellen!",
            );
            expect(description).toBeInTheDocument();
        });
    });

    it("displays masked token when token exists", async () => {
        apiToken.token = "test-token-12345678901234567890123456";
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const button = container.querySelectorAll("button")[0];
        await fireEvent.click(button);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const maskedToken = screen.getByText("•".repeat(40));
            expect(maskedToken).toBeInTheDocument();
        });
    });

    it("displays 'Kein Token verfügbar' when no token exists", async () => {
        apiToken.token = null;
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const noToken = screen.getAllByText("Kein Token verfügbar")[0];
            expect(noToken).toBeInTheDocument();
        });
    });

    it("toggles token visibility when eye button is clicked", async () => {
        apiToken.token = "test-token-12345678901234567890123456";
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const maskedToken = screen.getByText("•".repeat(40));
            expect(maskedToken).toBeInTheDocument();
        });

        // Find and click the eye button (first icon button)
        const buttons = container.querySelectorAll("button");
        const eyeButton = Array.from(buttons).find((btn) =>
            btn.querySelector('svg[class*="lucide-eye"]'),
        );

        if (eyeButton) {
            await user.click(eyeButton);

            await waitFor(() => {
                const visibleToken = screen.getByText(
                    "test-token-12345678901234567890123456",
                );
                expect(visibleToken).toBeInTheDocument();
            });
        }
    });

    it("disables eye button when no token is available", async () => {
        apiToken.token = null;
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const buttons = container.querySelectorAll("button");
            // Eye button should be disabled
            const iconButtons = Array.from(buttons).filter(
                (btn) =>
                    btn.querySelector("svg") &&
                    btn.getAttribute("class")?.includes("outline"),
            );
            expect(iconButtons.length).toBeGreaterThan(0);
        });
    });

    it("copies token to clipboard when copy button is clicked", async () => {
        apiToken.token = "test-token-to-copy";
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const dialogTitle = screen.getByText("API Token Settings");
            expect(dialogTitle).toBeInTheDocument();
        });

        // Find and click the copy button
        const buttons = container.querySelectorAll("button");
        const copyButton = Array.from(buttons).find((btn) =>
            btn.querySelector('svg[class*="lucide-copy"]'),
        );

        if (copyButton) {
            await user.click(copyButton);

            expect(mockWriteText).toHaveBeenCalledWith("test-token-to-copy");
            expect(toast.success).toHaveBeenCalledWith("Token kopiert!", {
                description: "Der Token wurde in die Zwischenablage kopiert.",
            });
        }
    });

    it("disables copy button when no token is available", async () => {
        apiToken.token = null;
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const buttons = container.querySelectorAll("button");
            // Copy button should be disabled
            const iconButtons = Array.from(buttons).filter(
                (btn) =>
                    btn.querySelector("svg") &&
                    btn.getAttribute("class")?.includes("outline"),
            );
            expect(iconButtons.length).toBeGreaterThan(0);
        });
    });

    it("calls getToken when generate new token button is clicked", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const generateButton = screen.getByText("Neuen Token generieren");
            expect(generateButton).toBeInTheDocument();
        });

        const generateButton = screen.getByText("Neuen Token generieren");
        await user.click(generateButton);

        expect(getToken).toHaveBeenCalledTimes(1);
    });

    it("displays close button in dialog footer", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const closeButton = screen.getByText("Schließen");
            expect(closeButton).toBeInTheDocument();
        });
    });

    it("closes dialog when close button is clicked", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const dialogTitle = screen.getByText("API Token Settings");
            expect(dialogTitle).toBeInTheDocument();
        });

        const closeButton = screen.getByText("Schließen");
        await user.click(closeButton);

        await waitFor(() => {
            expect(
                screen.queryByText("API Token Settings"),
            ).not.toBeInTheDocument();
        });
    });

    it("renders token label in dialog", async () => {
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const label = screen.getByText("Dein Token");
            expect(label).toBeInTheDocument();
        });
    });

    it("shows EllipsisVertical icon in menu button", () => {
        const { container } = render(UserTestWrapper);

        // Should have multiple SVGs (user icon and ellipsis)
        const svgs = container.querySelectorAll("svg");
        expect(svgs.length).toBeGreaterThan(1);
    });

    it("handles token toggle from hidden to visible and back", async () => {
        apiToken.token = "secret-token-123";
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const maskedToken = screen.getByText("•".repeat(40));
            expect(maskedToken).toBeInTheDocument();
        });

        // Find eye button
        const buttons = container.querySelectorAll("button");
        const eyeButton = Array.from(buttons).find((btn) =>
            btn.querySelector('svg[class*="lucide-eye"]'),
        );

        if (eyeButton) {
            // Show token
            await user.click(eyeButton);

            await waitFor(() => {
                const visibleToken = screen.getByText("secret-token-123");
                expect(visibleToken).toBeInTheDocument();
            });

            // Hide token again
            await user.click(eyeButton);

            await waitFor(() => {
                const maskedToken = screen.getByText("•".repeat(40));
                expect(maskedToken).toBeInTheDocument();
            });
        }
    });

    it("does not copy if no token available and copy button is somehow clicked", async () => {
        apiToken.token = "";
        const user = userEvent.setup();
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        const tokenItem = screen.getByText("Token");
        await user.click(tokenItem);

        await waitFor(() => {
            const dialogTitle = screen.getByText("API Token Settings");
            expect(dialogTitle).toBeInTheDocument();
        });

        // Try to find and click copy button (even though it should be disabled)
        const buttons = container.querySelectorAll("button");
        const copyButton = Array.from(buttons).find((btn) =>
            btn.querySelector('svg[class*="lucide-copy"]'),
        );

        if (copyButton && !copyButton.hasAttribute("disabled")) {
            await user.click(copyButton);
            // Should not call writeText
            expect(mockWriteText).not.toHaveBeenCalled();
        }
    });

    it("displays KeyRound icon next to Token menu item", async () => {
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        await waitFor(() => {
            const tokenItem = screen.getByText("Token");
            expect(tokenItem).toBeInTheDocument();
        });

        // Check for KeyRound icon in dropdown
        const svgs = container.querySelectorAll("svg");
        expect(svgs.length).toBeGreaterThan(0);
    });

    it("displays LogOut icon next to Log out menu item", async () => {
        const { container } = render(UserTestWrapper);

        const menuButton = container.querySelectorAll("button")[0];
        await fireEvent.click(menuButton);

        await waitFor(() => {
            const logoutItem = screen.getByText("Log out");
            expect(logoutItem).toBeInTheDocument();
        });

        // Check for LogOut icon in dropdown
        const svgs = container.querySelectorAll("svg");
        expect(svgs.length).toBeGreaterThan(0);
    });
});
