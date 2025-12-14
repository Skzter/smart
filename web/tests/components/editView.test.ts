import { render } from "@testing-library/svelte";
import { describe, it, expect, vi } from "vitest";
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
    default: class MockWorker {},
}));

import EditView from "../../src/lib/components/EditView.svelte";
import { Runner } from "../../src/lib/runner.svelte";

describe("EditView", () => {
    const mockRunner = new Runner();

    it("renders the monaco editor", () => {
        const { container } = render(EditView, {
            props: {
                code: "const test = 'hello';",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        // MonacoEditor component should be present (check for parent container)
        const editorContainer = container.querySelector(
            ".h-full.overflow-y-visible",
        );
        expect(editorContainer).toBeInTheDocument();
    });

    it("renders test information panel", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const heading = container.querySelector("h1.text-md.font-semibold");
        expect(heading?.textContent).toContain("Test Information");
    });

    it("displays test statistics labels", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const labels = container.querySelectorAll(".space-y-3 p");
        expect(labels.length).toBeGreaterThanOrEqual(3);

        const labelTexts = Array.from(labels).map((l) => l.textContent);
        expect(labelTexts).toContain("Zeilen:");
        expect(labelTexts).toContain("Zeichen:");
        expect(labelTexts).toContain("Status:");
    });

    it("renders quick actions panel", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const headings = container.querySelectorAll("h1.text-md.font-semibold");
        const quickActionsHeading = Array.from(headings).find((h) =>
            h.textContent?.includes("Schnellaktionen"),
        );
        expect(quickActionsHeading).toBeInTheDocument();
    });

    it("renders run and save buttons", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const buttons = container.querySelectorAll(".flex.flex-col button");
        expect(buttons.length).toBeGreaterThanOrEqual(2);
    });

    it("has correct grid layout", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const gridContainer = container.querySelector(".flex-1.grid");
        expect(gridContainer).toBeInTheDocument();
        expect(gridContainer).toHaveStyle({
            "grid-template-columns": "70% 30%",
        });
    });

    it("renders sidebar with correct styling", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const sidebar = container.querySelector(".bg-gray-300.border-l");
        expect(sidebar).toBeInTheDocument();
    });

    it("renders info cards with correct styling", () => {
        const { container } = render(EditView, {
            props: {
                code: "test code",
                activeTab: "edit",
                testRunner: mockRunner,
            },
        });

        const cards = container.querySelectorAll(".bg-gray-50.rounded-lg.p-6");
        expect(cards.length).toBe(2);
    });
});
