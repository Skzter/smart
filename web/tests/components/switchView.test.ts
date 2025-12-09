import { render, screen, fireEvent } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";

import ViewSwitcher from "../../src/lib/components/SwitchView.svelte";

describe("SwitchView", () => {

    it("renders 3 view buttons", () => {
        render(ViewSwitcher);

        const buttons = screen.getAllByRole("button");

        expect(buttons.length).toBeGreaterThanOrEqual(3);
    });

    it("defaults to code view", () => {
        render(ViewSwitcher);

        const buttons = screen.getAllByRole("button");
        const codeButton = buttons[0];

        expect(codeButton).toHaveClass("bg-primary");
    });

    it("changes view to split", async () => {
        render(ViewSwitcher);

        const buttons = screen.getAllByRole("button");
        const splitButton = buttons[1];

        await fireEvent.click(splitButton);

        expect(splitButton).toHaveClass("bg-primary");
    });

    it("changes view to browser", async () => {
        render(ViewSwitcher);

        const buttons = screen.getAllByRole("button");
        const browserButton = buttons[2];

        await fireEvent.click(browserButton);

        expect(browserButton).toHaveClass("bg-primary");
    });
});
