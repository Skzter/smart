import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import '@testing-library/jest-dom/vitest';

import ControlButtons from "../../src/lib/components/ControlButtons.svelte";

describe("ControlButtons", () => {
    it("renders all three buttons", () => {
        render(ControlButtons);

        const buttons = screen.getAllByRole("button");
        expect(buttons).toHaveLength(3);
    });

    it("renders pause button", () => {
        const { container } = render(ControlButtons);

        const pauseIcon = container.querySelector('svg.lucide-pause');
        expect(pauseIcon).toBeInTheDocument();
    });

    it("renders rotate button", () => {
        const { container } = render(ControlButtons);

        const rotateIcon = container.querySelector('svg.lucide-rotate-ccw');
        expect(rotateIcon).toBeInTheDocument();
    });

    it("renders close button", () => {
        const { container } = render(ControlButtons);

        const closeIcon = container.querySelector('svg.lucide-x');
        expect(closeIcon).toBeInTheDocument();
    });

    it("calls onCloseClick when close button is clicked", async () => {
        const user = userEvent.setup();
        const mockOnClose = vi.fn();

        render(ControlButtons, {
            props: {
                onCloseClick: mockOnClose,
            },
        });

        const buttons = screen.getAllByRole("button");
        const closeButton = buttons[2]; // Third button is the close button
        
        await user.click(closeButton);
        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it("has correct button styling", () => {
        const { container } = render(ControlButtons);

        const buttons = container.querySelectorAll('button');
        buttons.forEach(button => {
            expect(button).toHaveClass('h-6', 'w-6', 'p-0');
        });
    });
});
