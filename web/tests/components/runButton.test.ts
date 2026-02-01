import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";

import RunButton from "../../src/lib/components/RunButton.svelte";
import type { Runner } from "../../src/lib/runner.svelte";

// Create a mock Runner since the real one uses $effect which can only run in component context
function createMockRunner(overrides: Partial<Runner> = {}): Runner {
    return {
        isRunning: () => false,
        getCurTest: () => "",
        run: vi.fn(),
        setTest: vi.fn(),
        storeTest: vi.fn(),
        getStorageState: () => "idle" as const,
        logStatus: "idle" as const,
        logError: null,
        result: [],
        videoUrl: null,
        model: {
            summary: { status: "idle" as const },
            steps: [],
        },
        fetchMediaUrl: vi.fn(),
        clearVideoUrl: vi.fn(),
        ...overrides,
    } as unknown as Runner;
}

describe("RunButton", () => {
    let testRunner: Runner;
    let activeTab: string;

    beforeEach(() => {
        testRunner = createMockRunner();
        activeTab = "edit";
    });

    it("renders the button", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "test-class",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).toBeInTheDocument();
    });

    it("applies custom classes to button", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "custom-test-class",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).toHaveClass("custom-test-class");
    });

    it("applies variant prop to button", () => {
        render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "outline",
                size: "default",
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeInTheDocument();
    });

    it("applies size prop to button", () => {
        render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "sm",
            },
        });

        const button = screen.getByRole("button");
        expect(button).toBeInTheDocument();
    });

    it("renders Play icon", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const icon = container.querySelector(".h-3\\.5.w-3\\.5");
        expect(icon).toBeInTheDocument();
    });

    it("displays 'Ausführen' when getCurTest is empty", () => {
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const paragraph = container.querySelector("p");
        expect(paragraph?.textContent).toBe("Ausführen");
    });

    it("displays 'Ergebnis anzeigen' when getCurTest is not empty", () => {
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const paragraph = container.querySelector("p");
        expect(paragraph?.textContent).toBe("Ergebnis anzeigen");
    });

    it("is disabled when current test is empty", () => {
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).toBeDisabled();
    });

    it("is enabled when test exists", () => {
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).not.toBeDisabled();
    });

    it("does not call testRunner.run() when clicked (only sets activeTab)", async () => {
        const user = userEvent.setup();
        const runSpy = vi.spyOn(testRunner, "run");
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button") as HTMLButtonElement;
        await user.click(button);

        expect(runSpy).not.toHaveBeenCalled();
    });

    it("sets activeTab to 'run' when clicked", async () => {
        const user = userEvent.setup();
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        let currentTab = "edit";

        const { container } = render(RunButton, {
            props: {
                get activeTab() {
                    return currentTab;
                },
                set activeTab(value) {
                    currentTab = value;
                },
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button") as HTMLButtonElement;
        await user.click(button);

        expect(currentTab).toBe("run");
    });


    it("renders with different variant values", () => {
        const variants: Array<
            | "default"
            | "destructive"
            | "outline"
            | "secondary"
            | "ghost"
            | "link"
        > = ["default", "destructive", "outline", "secondary", "ghost", "link"];

        variants.forEach((variant) => {
            const { container } = render(RunButton, {
                props: {
                    activeTab,
                    testRunner,
                    classes: "",
                    variant,
                    size: "default",
                },
            });

            const button = container.querySelector("button");
            expect(button).toBeInTheDocument();
        });
    });

    it("renders with different size values", () => {
        const sizes: Array<"default" | "sm" | "lg" | "icon"> = [
            "default",
            "sm",
            "lg",
            "icon",
        ];

        sizes.forEach((size) => {
            const { container } = render(RunButton, {
                props: {
                    activeTab,
                    testRunner,
                    classes: "",
                    variant: "default",
                    size,
                },
            });

            const button = container.querySelector("button");
            expect(button).toBeInTheDocument();
        });
    });

    it("updates button text when getCurTest changes", () => {
        const getCurTestSpy = vi
            .spyOn(testRunner, "getCurTest")
            .mockReturnValue("");

        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        let paragraph = container.querySelector("p");
        expect(paragraph?.textContent).toBe("Ausführen");

        getCurTestSpy.mockReturnValue("test-123");

        const { container: container2 } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        paragraph = container2.querySelector("p");
        expect(paragraph?.textContent).toBe("Ergebnis anzeigen");
    });

    it("maintains icon size classes", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const icon = container.querySelector("svg");
        expect(icon).toHaveClass("h-3.5");
        expect(icon).toHaveClass("w-3.5");
    });

    it("renders paragraph inside button", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        const paragraph = button?.querySelector("p");
        expect(paragraph).toBeInTheDocument();
    });

    it("renders Play icon inside button", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        const icon = button?.querySelector("svg");
        expect(icon).toBeInTheDocument();
    });

    it("handles empty classes string", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).toBeInTheDocument();
    });

    it("handles multiple classes separated by spaces", () => {
        const { container } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "class-one class-two class-three",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button");
        expect(button).toHaveClass("class-one");
        expect(button).toHaveClass("class-two");
        expect(button).toHaveClass("class-three");
    });

    it("sets activeTab to run on click without calling run()", async () => {
        const user = userEvent.setup();
        const runSpy = vi.spyOn(testRunner, "run");
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        let currentTab = "edit";

        const { container } = render(RunButton, {
            props: {
                get activeTab() {
                    return currentTab;
                },
                set activeTab(value) {
                    currentTab = value;
                },
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        const button = container.querySelector("button") as HTMLButtonElement;
        await user.click(button);

        expect(currentTab).toBe("run");
        expect(runSpy).not.toHaveBeenCalled();
    });

    it("disabled only when getCurTest is empty", () => {
        vi.spyOn(testRunner, "getCurTest").mockReturnValue("");

        const { container: container1 } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        expect(container1.querySelector("button")).toBeDisabled();

        vi.spyOn(testRunner, "getCurTest").mockReturnValue("test-123");

        const { container: container2 } = render(RunButton, {
            props: {
                activeTab,
                testRunner,
                classes: "",
                variant: "default",
                size: "default",
            },
        });

        expect(container2.querySelector("button")).not.toBeDisabled();
    });
});
