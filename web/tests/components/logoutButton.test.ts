import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import LogoutButton from "../../src/components/LogoutButton.svelte";
import { auth } from "../../src/lib/authService";

vi.mock("../../src/lib/authService", () => {
    return {
        auth: {
            logout: vi.fn(),
        },
    };
});

describe("LogoutButton", () => {
    it("renders a logout button", () => {
        render(LogoutButton);
        const button = screen.getByRole("button", { name: /log out/i });
        expect(button).toBeInTheDocument();
    });

    it("calls auth.logout when clicked", async () => {
        const user = userEvent.setup();
        render(LogoutButton);
        const button = screen.getByRole("button", { name: /log out/i });
        await user.click(button);
        expect(auth.logout).toHaveBeenCalled();
    });
});
