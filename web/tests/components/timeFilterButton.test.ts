import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect } from "vitest";
import TimeFilterButton from "$lib/components/TimeFilterButton.svelte";
import '@testing-library/jest-dom/vitest';

describe("TimeFilterButton", () => {
    it("renders a button with text 'Time Filter'", () => {
        const showCalendar = false;

        render(TimeFilterButton, {
            props: {
                showCalendar,
            },
        });

        const button = screen.getByRole("button", { name: /time filter/i });
        expect(button).toBeInTheDocument();
    });

    it("displays 'Time Filter' text", () => {
        const showCalendar = false;

        render(TimeFilterButton, {
            props: {
                showCalendar,
            },
        });

        const text = screen.getByText("Time Filter");
        expect(text).toBeInTheDocument();
    });

    it("displays Calendar icon", () => {
        const showCalendar = false;

        const { container } = render(TimeFilterButton, {
            props: {
                showCalendar,
            },
        });

        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });

    it("toggles showCalendar from false to true when clicked", async () => {
        const user = userEvent.setup();
        let showCalendar = false;

        render(TimeFilterButton, {
            props: {
                get showCalendar() {
                    return showCalendar;
                },
                set showCalendar(value) {
                    showCalendar = value;
                },
            },
        });

        expect(showCalendar).toBe(false);

        const button = screen.getByRole("button", { name: /time filter/i });
        await user.click(button);

        expect(showCalendar).toBe(true);
    });

    it("toggles showCalendar from true to false when clicked", async () => {
        const user = userEvent.setup();
        let showCalendar = true;

        render(TimeFilterButton, {
            props: {
                get showCalendar() {
                    return showCalendar;
                },
                set showCalendar(value) {
                    showCalendar = value;
                },
            },
        });

        expect(showCalendar).toBe(true);

        const button = screen.getByRole("button", { name: /time filter/i });
        await user.click(button);

        expect(showCalendar).toBe(false);
    });

    it("can be clicked multiple times and toggles correctly", async () => {
        const user = userEvent.setup();
        let showCalendar = false;

        render(TimeFilterButton, {
            props: {
                get showCalendar() {
                    return showCalendar;
                },
                set showCalendar(value) {
                    showCalendar = value;
                },
            },
        });

        const button = screen.getByRole("button", { name: /time filter/i });

        // First click: false -> true
        await user.click(button);
        expect(showCalendar).toBe(true);

        // Second click: true -> false
        await user.click(button);
        expect(showCalendar).toBe(false);

        // Third click: false -> true
        await user.click(button);
        expect(showCalendar).toBe(true);

        // Fourth click: true -> false
        await user.click(button);
        expect(showCalendar).toBe(false);
    });

    it("is always enabled", () => {
        const showCalendar = false;

        render(TimeFilterButton, {
            props: {
                showCalendar,
            },
        });

        const button = screen.getByRole("button", { name: /time filter/i });
        expect(button).not.toBeDisabled();
    });

    it("is enabled when showCalendar is true", () => {
        const showCalendar = true;

        render(TimeFilterButton, {
            props: {
                showCalendar,
            },
        });

        const button = screen.getByRole("button", { name: /time filter/i });
        expect(button).not.toBeDisabled();
    });
});
