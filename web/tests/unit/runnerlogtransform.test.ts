import { describe, it, expect } from "vitest";
import { buildStepTree } from "../../src/lib/runnerlogtransform";

type Case = {
    name: string;
    logs: { begin: string }[];
    expected: {
        status: "idle" | "running" | "passed" | "failed";
        duration?: number;
        finalLabel?: string;
        hasGroup?: boolean;
        failedStep?: string;
    };
};

const RAW_LOGS_OK = [
    {
        begin: JSON.stringify({
            type: "run:start",
            timestamp: 1000,
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:start",
            stepIndex: 1,
            step: "Before Hooks",
            timestamp: 1100,
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:start",
            stepIndex: 2,
            step: "Compile",
            timestamp: 1150,
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:end",
            stepIndex: 2,
            timestamp: 1200,
            status: "passed",
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:end",
            stepIndex: 1,
            timestamp: 1250,
            status: "passed",
        }),
    },
    {
        begin: JSON.stringify({
            type: "run:end",
            timestamp: 2300,
            status: "passed",
        }),
    },
];

const RAW_LOGS_FAILED = [
    {
        begin: JSON.stringify({
            type: "run:start",
            timestamp: 1000,
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:start",
            stepIndex: 1,
            step: "Compile",
            timestamp: 1100,
        }),
    },
    {
        begin: JSON.stringify({
            type: "step:end",
            stepIndex: 1,
            timestamp: 1300,
            status: "failed",
        }),
    },
    {
        begin: JSON.stringify({
            type: "run:end",
            timestamp: 1500,
            status: "failed",
        }),
    },
];

const CASES: Case[] = [
    {
        name: "idle when no logs",
        logs: [],
        expected: {
            status: "idle",
        },
    },
    {
        name: "running when run:start but no run:end",
        logs: [
            {
                begin: JSON.stringify({
                    type: "run:start",
                    timestamp: 1000,
                }),
            },
        ],
        expected: {
            status: "running",
        },
    },
    {
        name: "successful run builds tree and duration",
        logs: RAW_LOGS_OK,
        expected: {
            status: "passed",
            duration: 1300, // 2300 - 1000
            finalLabel: "Test abgeschlossen",
            hasGroup: true,
        },
    },
    {
        name: "failed run marks failure and duration",
        logs: RAW_LOGS_FAILED,
        expected: {
            status: "failed",
            duration: 500, // 1500 - 1000
            finalLabel: "Test fehlgeschlagen",
            failedStep: "Compile",
        },
    },
];

describe("buildStepTree", () => {
    it.each(CASES)("$name", ({ logs, expected }) => {
        const result = buildStepTree(logs);

        // Summary status
        expect(result.summary.status).toBe(expected.status);

        // Duration
        if (expected.duration !== undefined) {
            expect(result.summary.durationMs).toBe(expected.duration);
        }

        // Final step label
        if (expected.finalLabel) {
            const last = result.steps[result.steps.length - 1];
            expect(last.label).toBe(expected.finalLabel);
        }

        // Groups
        if (expected.hasGroup) {
            const group = result.steps.find((s) => s.kind === "group");
            expect(group).toBeDefined();
            expect(group!.children.length).toBeGreaterThan(0);
        }

        // Failed step
        if (expected.failedStep) {
            const failed = result.steps.find(
                (s) => s.label === expected.failedStep,
            );
            expect(failed).toBeDefined();
            expect(failed!.status).toBe("failed");
        }
    });
});
