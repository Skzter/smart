import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import LoginButton from "../../src/components/LoginButton.svelte";
import { auth } from "../../src/lib/authService";

vi.mock("../../src/lib/authService", () => {
    return {
        auth: {
            login: vi.fn(),
        },
    };
});

describe("LoginButton", () => {
    it("renders a login button", () => {
        render(LoginButton);
        const button = screen.getByRole("button", { name: /log in/i });
        expect(button).toBeInTheDocument();
    });

    it("calls auth.login when clicked", async () => {
        const user = userEvent.setup();
        render(LoginButton);
        const button = screen.getByRole("button", { name: /log in/i });
        await user.click(button);
        expect(auth.login).toHaveBeenCalled();
    });
});
