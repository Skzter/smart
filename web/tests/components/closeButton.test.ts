import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, beforeEach, vi } from "vitest";
import CloseButton from "../../src/lib/components/CloseButton.svelte";
import '@testing-library/jest-dom/vitest';

describe("CloseButton", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it("renders a button with X icon", () => {
        const { container } = render(CloseButton);

        const button = screen.getByRole("button");
        expect(button).toBeInTheDocument();

        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
        expect(svg).toHaveClass('h-4', 'w-4');
    });

    it("applies correct styling classes", () => {
        const { container } = render(CloseButton);

        const button = container.querySelector('button');
        expect(button).toHaveClass('h-6', 'w-6', 'p-0', 'cursor-pointer');
    });

    it("calls onCloseClick when clicked", async () => {
        const user = userEvent.setup();
        const onCloseClick = vi.fn();

        render(CloseButton, { props: { onCloseClick } });

        const button = screen.getByRole("button");
        await user.click(button);

        expect(onCloseClick).toHaveBeenCalledTimes(1);
    });

    it("calls onCloseClick multiple times when clicked multiple times", async () => {
        const user = userEvent.setup();
        const onCloseClick = vi.fn();

        render(CloseButton, { props: { onCloseClick } });

        const button = screen.getByRole("button");

        await user.click(button);
        await user.click(button);
        await user.click(button);

        expect(onCloseClick).toHaveBeenCalledTimes(3);
    });

    it("works without onCloseClick prop provided", async () => {
        const user = userEvent.setup();

        render(CloseButton);

        const button = screen.getByRole("button");
        await expect(user.click(button)).resolves.not.toThrow();
    });

    it("can be triggered with keyboard", async () => {
        const user = userEvent.setup();
        const onCloseClick = vi.fn();

        render(CloseButton, { props: { onCloseClick } });

        const button = screen.getByRole("button");
        button.focus();

        await user.keyboard('{Enter}');

        expect(onCloseClick).toHaveBeenCalledTimes(1);
    });

    it("is always enabled", () => {
        render(CloseButton);

        const button = screen.getByRole("button");
        expect(button).not.toBeDisabled();
    });
});