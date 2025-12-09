import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect } from "vitest";
import EditButton from "$lib/components/EditButton.svelte";
import "@testing-library/jest-dom/vitest";

describe("EditButton", () => {
    it("renders a button with text 'Bearbeiten'", () => {
        const activeTab = "view";

        render(EditButton, {
            props: {
                activeTab,
            },
        });

        const button = screen.getByRole("button", { name: /bearbeiten/i });
        expect(button).toBeInTheDocument();
    });

    it("displays 'Bearbeiten' text", () => {
        const activeTab = "view";

        render(EditButton, {
            props: {
                activeTab,
            },
        });

        const text = screen.getByText("Bearbeiten");
        expect(text).toBeInTheDocument();
    });

    it("displays SquarePen icon", () => {
        const activeTab = "view";

        const { container } = render(EditButton, {
            props: {
                activeTab,
            },
        });

        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
    });

    it("sets activeTab to 'edit' when clicked", async () => {
        const user = userEvent.setup();
        let activeTab = "view";

        render(EditButton, {
            props: {
                get activeTab() {
                    return activeTab;
                },
                set activeTab(value) {
                    activeTab = value;
                },
            },
        });

        expect(activeTab).toBe("view");

        const button = screen.getByRole("button", { name: /bearbeiten/i });
        await user.click(button);

        expect(activeTab).toBe("edit");
    });

    it("changes activeTab from any value to 'edit'", async () => {
        const user = userEvent.setup();
        let activeTab = "preview";

        render(EditButton, {
            props: {
                get activeTab() {
                    return activeTab;
                },
                set activeTab(value) {
                    activeTab = value;
                },
            },
        });

        expect(activeTab).toBe("preview");

        const button = screen.getByRole("button", { name: /bearbeiten/i });
        await user.click(button);

        expect(activeTab).toBe("edit");
    });

    it("can be clicked multiple times", async () => {
        const user = userEvent.setup();
        let activeTab = "view";

        render(EditButton, {
            props: {
                get activeTab() {
                    return activeTab;
                },
                set activeTab(value) {
                    activeTab = value;
                },
            },
        });

        const button = screen.getByRole("button", { name: /bearbeiten/i });

        await user.click(button);
        expect(activeTab).toBe("edit");

        // Change it back
        activeTab = "view";

        await user.click(button);
        expect(activeTab).toBe("edit");

        // Change it again
        activeTab = "preview";

        await user.click(button);
        expect(activeTab).toBe("edit");
    });

    it("sets activeTab to 'edit' even when already 'edit'", async () => {
        const user = userEvent.setup();
        let activeTab = "edit";

        render(EditButton, {
            props: {
                get activeTab() {
                    return activeTab;
                },
                set activeTab(value) {
                    activeTab = value;
                },
            },
        });

        const button = screen.getByRole("button", { name: /bearbeiten/i });
        await user.click(button);

        expect(activeTab).toBe("edit");
    });

    it("is always enabled", () => {
        const activeTab = "view";

        render(EditButton, {
            props: {
                activeTab,
            },
        });

        const button = screen.getByRole("button", { name: /bearbeiten/i });
        expect(button).not.toBeDisabled();
    });
});
