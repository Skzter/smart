import { render, screen } from "@testing-library/svelte";
import { describe, it, expect, beforeEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";

import OutputView from "../../src/lib/components/OutputView.svelte";
import type { Runner } from "../../src/lib/runner.svelte";

/* ---------------- Types used by OutputView ---------------- */

type Summary = {
    status: "idle" | "running" | "passed" | "failed";
    durationMs?: number;
};

type Step = {
    kind?: "group" | "step";
    label: string;
    status?: "running" | "done" | "failed";
    children?: Step[];
};

type Model = {
    summary: Summary;
    steps: Step[];
};

type TestRunnerMock = {
    model: Model;
    logStatus: "idle" | "connecting" | "connected" | "error";
    fetchMediaUrl: () => void;
    clearVideoUrl: () => void;
};

/* ---------------- Helpers ---------------- */

function makeRunner(model: Model): TestRunnerMock {
    return {
        model,
        logStatus: "idle",
        fetchMediaUrl: vi.fn(),
        clearVideoUrl: vi.fn(),
    };
}

/* ---------------- Tests ---------------- */

describe("OutputView", () => {
    let runner: TestRunnerMock;

    beforeEach(() => {
        runner = makeRunner({
            summary: { status: "idle" },
            steps: [],
        });
    });

    it("renders main container", () => {
        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(document.querySelector(".flex.flex-col")).toBeInTheDocument();
    });

    it("renders log area", () => {
        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(
            document.querySelector(".flex-1.overflow-auto"),
        ).toBeInTheDocument();
    });

    it("renders footer", () => {
        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(document.querySelector(".border-t")).toBeInTheDocument();
    });

    it("shows idle state initially", () => {
        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText("inaktiv")).toBeInTheDocument();
    });

    it("shows no connection initially", () => {
        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText("Keine Verbindung")).toBeInTheDocument();
    });

    it("renders running summary", () => {
        runner = makeRunner({
            summary: { status: "running" },
            steps: [],
        });

        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText("Test läuft…")).toBeInTheDocument();
    });

    it("renders passed summary with duration", () => {
        runner = makeRunner({
            summary: { status: "passed", durationMs: 1300 },
            steps: [],
        });

        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText(/Test erfolgreich/)).toBeInTheDocument();
        expect(screen.getByText("1.3 s")).toBeInTheDocument();
    });

    it("renders failed summary with duration", () => {
        runner = makeRunner({
            summary: { status: "failed", durationMs: 3000 },
            steps: [],
        });

        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText(/Test fehlgeschlagen/)).toBeInTheDocument();
        expect(screen.getByText("3.0 s")).toBeInTheDocument();
    });

    it("renders steps", () => {
        runner = makeRunner({
            summary: { status: "passed", durationMs: 100 },
            steps: [
                {
                    kind: "step",
                    label: "Compile",
                    status: "done",
                    children: [],
                },
                {
                    kind: "step",
                    label: "Run tests",
                    status: "done",
                    children: [],
                },
            ],
        });

        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText("Compile")).toBeInTheDocument();
        expect(screen.getByText("Run tests")).toBeInTheDocument();
    });

    it("renders nested steps", () => {
        runner = makeRunner({
            summary: { status: "passed", durationMs: 100 },
            steps: [
                {
                    kind: "group",
                    label: "Suite A",
                    status: "done",
                    children: [
                        {
                            kind: "step",
                            label: "Test 1",
                            status: "done",
                            children: [],
                        },
                    ],
                },
            ],
        });

        render(OutputView, {
            props: { testRunner: runner as unknown as Runner },
        });
        expect(screen.getByText("Suite A")).toBeInTheDocument();
        expect(screen.getByText("Test 1")).toBeInTheDocument();
    });
});
