import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, beforeEach, vi } from "vitest";
import TimeFilter from "../../src/lib/components/TimeFilter.svelte";
import { ChatDate } from "$lib/shared.svelte";
import "@testing-library/jest-dom/vitest";

vi.mock("$lib/components/ui/range-calendar", () => ({
    RangeCalendar: () => {
        return {};
    },
}));

describe("TimeFilter", () => {
    beforeEach(() => {
        ChatDate.Range = undefined;
    });

    it("renders the wrapper with correct width", () => {
        const { container } = render(TimeFilter);
        const wrapper = container.querySelector(".w-full");
        expect(wrapper).toBeInTheDocument();
    });

    it("renders TimeFilterButton component", () => {
        const { container } = render(TimeFilter);
        expect(container.firstChild).toBeInTheDocument();
    });

    it("does not show calendar wrapper initially", () => {
        const { container } = render(TimeFilter);
        const calendarWrapper = container.querySelector(".scale-\\[0\\.85\\]");
        expect(calendarWrapper).not.toBeInTheDocument();
    });

    it("shows calendar wrapper when button is clicked", async () => {
        const user = userEvent.setup();
        const { container } = render(TimeFilter);
        const button = screen.getByRole("button");

        await user.click(button);

        await waitFor(() => {
            const calendarWrapper = container.querySelector(
                ".scale-\\[0\\.85\\]",
            );
            expect(calendarWrapper).toBeInTheDocument();
        });
    });

    it("applies scale transformation to calendar wrapper", async () => {
        const user = userEvent.setup();
        const { container } = render(TimeFilter);
        const button = screen.getByRole("button");

        await user.click(button);

        await waitFor(() => {
            const calendarWrapper = container.querySelector(
                ".scale-\\[0\\.85\\]",
            );
            expect(calendarWrapper).toBeInTheDocument();
            expect(calendarWrapper).toHaveClass("-my-7");
        });
    });
});
