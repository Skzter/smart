import { render, fireEvent } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock Monaco Editor
vi.mock("monaco-editor", () => ({
    editor: {
        create: vi.fn(() => ({
            getValue: vi.fn(() => ""),
            setValue: vi.fn(),
            getModel: vi.fn(() => ({
                onDidChangeContent: vi.fn(() => ({ dispose: vi.fn() })),
            })),
            getPosition: vi.fn(),
            setPosition: vi.fn(),
            dispose: vi.fn(),
        })),
    },
}));

// Mock the worker
vi.mock("monaco-editor/esm/vs/language/typescript/ts.worker?worker", () => ({
    default: class MockWorker { },
}));

import type { Runner } from "../../src/lib/runner.svelte";
import RunWindowTestWrapper from "../helpers/RunWindowTestWrapper.svelte";

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

describe("RunWindow", () => {
    let testRunner: Runner;

    beforeEach(() => {
        testRunner = createMockRunner();
    });

    it("renders the dialog content with correct styling", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "const test = 'hello';",
                activeTab: "edit",
                testRunner,
            },
        });

        const dialogContent = document.body.querySelector(
            ".sm\\:max-w-\\[90vw\\]",
        );
        expect(dialogContent).toBeInTheDocument();
    });

    it("renders the dialog title", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const title = document.body.querySelector(".text-lg.font-semibold");
        expect(title).toBeInTheDocument();
        expect(title?.textContent).toBe("Button Click Test");
    });

    it("renders header with correct structure", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const header = document.body.querySelector(
            ".flex.flex-row.items-center.justify-between.border-b",
        );
        expect(header).toBeInTheDocument();
    });

    it("renders CloseButton when activeTab is 'edit'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const closeButtons = document.body.querySelectorAll("button");
        const hasCloseButton = Array.from(closeButtons).some((btn: Element) =>
            btn.querySelector("svg"),
        );
        expect(hasCloseButton).toBe(true);
    });

    it("renders SwitchView when activeTab is 'run'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const buttons = document.body.querySelectorAll("button");
        expect(buttons.length).toBeGreaterThan(0);
    });

    it("renders CloseButton when activeTab is 'result'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "result",
                testRunner,
            },
        });
        const closeButtons = document.body.querySelectorAll("button");
        const hasCloseButton = Array.from(closeButtons).some((btn: Element) =>
            btn.querySelector("svg"),
        );
        expect(hasCloseButton).toBe(true);
    });

    it("renders TabsView component", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const tabsContainer = document.body.querySelector(".px-6");
        expect(tabsContainer).toBeInTheDocument();
    });

    it("renders EditView when activeTab is 'edit'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const gridLayout = document.body.querySelector(".flex-1.grid");
        expect(gridLayout).toBeInTheDocument();
    });

    it("renders split view when activeTab is 'run' and view is 'split'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const splitLayout = document.body.querySelector(".grid.grid-cols-2");
        expect(splitLayout).toBeInTheDocument();
    });

    it("renders OutputView when activeTab is 'run'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const outputHeader = Array.from(
            document.body.querySelectorAll(".px-4.py-2.bg-muted\\/50"),
        ).find((el: Element) => el.textContent?.includes("Test Output"));
        expect(outputHeader).toBeInTheDocument();
    });

    it("renders BrowserView when activeTab is 'run'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const browserPreview = Array.from(
            document.body.querySelectorAll(".px-4.py-2.border-b"),
        ).find((el: Element) => el.textContent?.includes("Vorschau"));
        expect(browserPreview).toBeInTheDocument();
    });

    it("renders ResultView when activeTab is 'result'", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "result",
                testRunner,
            },
        });
        const resultPlaceholder = document.body.querySelector(".text-center");
        expect(resultPlaceholder).toBeInTheDocument();
    });

    it("renders hidden Dialog.Close button", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const hiddenClose = document.body.querySelector("[data-dialog-close]");
        expect(hiddenClose).toBeInTheDocument();
    });

    it("renders with different testRunner instances", () => {
        const customRunner = createMockRunner({ result: [] });
        const { container } = render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner: customRunner,
            },
        });
        expect(container).toBeInTheDocument();
    });

    it("renders all required components", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const dialogContent = document.body.querySelector(
            ".sm\\:max-w-\\[90vw\\]",
        );
        expect(dialogContent).toBeInTheDocument();
        const title = document.body.querySelector(".text-lg.font-semibold");
        expect(title).toBeInTheDocument();
        const tabs = document.body.querySelector(".px-6");
        expect(tabs).toBeInTheDocument();
    });

    it("handles empty code string", () => {
        const { container } = render(RunWindowTestWrapper, {
            props: {
                code: "",
                activeTab: "edit",
                testRunner,
            },
        });
        expect(container).toBeInTheDocument();
    });

    it("handles long code strings", () => {
        const longCode = "const test = 'hello';\n".repeat(100);
        const { container } = render(RunWindowTestWrapper, {
            props: {
                code: longCode,
                activeTab: "edit",
                testRunner,
            },
        });
        expect(container).toBeInTheDocument();
    });

    it("renders with all three activeTab options", () => {
        const tabs = ["edit", "run", "result"];
        tabs.forEach((tab) => {
            const { container } = render(RunWindowTestWrapper, {
                props: {
                    code: "test",
                    activeTab: tab,
                    testRunner,
                },
            });
            expect(container).toBeInTheDocument();
        });
    });

    it("maintains proper layout structure", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const dialogContent = document.body.querySelector(".flex.flex-col");
        expect(dialogContent).toBeInTheDocument();
        const header = document.body.querySelector(".border-b");
        expect(header).toBeInTheDocument();
    });

    it("renders overflow handling correctly", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const overflowContainer =
            document.body.querySelector(".overflow-hidden");
        expect(overflowContainer).toBeInTheDocument();
    });

    it("handles all conditional rendering branches for header controls", () => {
        const tabs = ["edit", "run", "result"];
        tabs.forEach((tab) => {
            render(RunWindowTestWrapper, {
                props: {
                    code: "test",
                    activeTab: tab,
                    testRunner,
                },
            });
            const controlsContainer = document.body.querySelector(
                ".flex.items-center.gap-2",
            );
            expect(controlsContainer).toBeInTheDocument();
        });
    });

    it("renders content area correctly for each tab", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test",
                activeTab: "edit",
                testRunner,
            },
        });
        expect(
            document.body.querySelector(".flex-1.overflow-visible"),
        ).toBeInTheDocument();

        render(RunWindowTestWrapper, {
            props: {
                code: "test",
                activeTab: "run",
                testRunner,
            },
        });
        expect(document.body.querySelector(".flex-1")).toBeInTheDocument();

        render(RunWindowTestWrapper, {
            props: {
                code: "test",
                activeTab: "result",
                testRunner,
            },
        });
        expect(
            document.body.querySelector(".flex-1.overflow-auto"),
        ).toBeInTheDocument();
    });

    it("integrates all child components properly", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const splitLayout = document.body.querySelector(".grid.grid-cols-2");
        expect(splitLayout).toBeInTheDocument();
        const children = splitLayout?.children;
        expect(children?.length).toBe(2);
    });

    it.skip("passes testRunner prop correctly to child components", () => {
        testRunner.result = "Test result from runner";
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        const outputHeader = Array.from(
            document.body.querySelectorAll(".px-4.py-2"),
        ).find((el: Element) =>
            el.textContent?.includes("Test result from runner"),
        );
        expect(outputHeader).toBeDefined();
    });

    it("renders with proper responsive classes", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const dialogContent = document.body.querySelector(
            ".sm\\:max-w-\\[90vw\\].md\\:max-w-\\[80vw\\].lg\\:max-w-\\[1170px\\]",
        );
        expect(dialogContent).toBeInTheDocument();
    });

    it("renders height classes correctly", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });
        const dialogContent = document.body.querySelector(".h-\\[85vh\\]");
        expect(dialogContent).toBeInTheDocument();
    });

    it("handles view state in run tab - code view", () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "run",
                testRunner,
            },
        });
        // Default should be split view
        const splitLayout = document.body.querySelector(".grid.grid-cols-2");
        expect(splitLayout).toBeInTheDocument();
    });

    it("renders all conditional branches", () => {
        ["edit", "run", "result"].forEach((tab) => {
            render(RunWindowTestWrapper, {
                props: {
                    code: "test",
                    activeTab: tab,
                    testRunner,
                },
            });
            expect(
                document.body.querySelector(".flex.flex-col"),
            ).toBeInTheDocument();
        });
    });

    it("handles tab change event", async () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });

        // Find and click the Run tab
        const runTab = Array.from(
            document.body.querySelectorAll('button[role="tab"]'),
        ).find((btn: Element) =>
            btn.textContent?.includes("Run"),
        ) as HTMLElement;

        expect(runTab).toBeTruthy();
        await fireEvent.click(runTab);

        // Verify the run view is displayed
        const splitLayout = document.body.querySelector(".grid.grid-cols-2");
        expect(splitLayout).toBeInTheDocument();
    });

    it("handles close button click", async () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });

        // Find close button
        const closeButton = Array.from(
            document.body.querySelectorAll("button"),
        ).find((btn: Element) => btn.querySelector("svg")) as HTMLElement;

        expect(closeButton).toBeTruthy();

        // Mock the dialog close button click
        const dialogCloseBtn = document.body.querySelector(
            "[data-dialog-close]",
        ) as HTMLElement;
        const clickSpy = vi.fn();
        if (dialogCloseBtn) {
            dialogCloseBtn.addEventListener("click", clickSpy);
            await fireEvent.click(closeButton);
        }
    });

    it("switches tabs via tab change handler", async () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });

        // Click Result tab
        const resultTab = Array.from(
            document.body.querySelectorAll('button[role="tab"]'),
        ).find((btn: Element) =>
            btn.textContent?.includes("Result"),
        ) as HTMLElement;

        expect(resultTab).toBeTruthy();
        await fireEvent.click(resultTab);

        // Verify result view is displayed
        const resultView = document.body.querySelector(".text-center");
        expect(resultView).toBeInTheDocument();
    });

    it("changes activeTab when switching between tabs", async () => {
        render(RunWindowTestWrapper, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner,
            },
        });

        const editTab = Array.from(
            document.body.querySelectorAll('button[role="tab"]'),
        ).find((btn: Element) =>
            btn.textContent?.includes("Edit"),
        ) as HTMLElement;
        const runTab = Array.from(
            document.body.querySelectorAll('button[role="tab"]'),
        ).find((btn: Element) =>
            btn.textContent?.includes("Run"),
        ) as HTMLElement;

        // Switch to Run
        await fireEvent.click(runTab);
        expect(
            document.body.querySelector(".grid.grid-cols-2"),
        ).toBeInTheDocument();

        // Switch back to Edit
        await fireEvent.click(editTab);
        expect(document.body.querySelector(".flex-1.grid")).toBeInTheDocument();
    });
});
