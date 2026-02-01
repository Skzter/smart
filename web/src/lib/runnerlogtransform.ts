export type StepNode = {
    id: number;
    time: string;
    label: string;
    kind: "group" | "step";
    status: "running" | "done" | "failed";
    children: StepNode[];
};

export type RunSummary = {
    status: "idle" | "running" | "passed" | "failed";
    durationMs?: number;
};

type RunStartEvent = {
    type: "run:start";
    timestamp: number;
};

type RunEndEvent = {
    type: "run:end";
    timestamp: number;
    status: "passed" | "failed";
};

type StepStartEvent = {
    type: "step:start";
    stepIndex: number;
    step: string;
    timestamp: number;
};

type StepEndEvent = {
    type: "step:end";
    stepIndex: number;
    timestamp: number;
    status: "passed" | "failed";
};

type PlaywrightEvent =
    | RunStartEvent
    | RunEndEvent
    | StepStartEvent
    | StepEndEvent;

function isGroup(label: string): boolean {
    return (
        label === "Before Hooks" ||
        label === "After Hooks" ||
        label === "Worker Cleanup" ||
        label.startsWith('Fixture "') ||
        label.startsWith("auto-playwright.ai")
    );
}

function parseEvent(raw: string): PlaywrightEvent | null {
    try {
        const parsed = JSON.parse(raw) as unknown;

        if (
            typeof parsed === "object" &&
            parsed !== null &&
            "type" in parsed &&
            typeof (parsed as { type: unknown }).type === "string"
        ) {
            return parsed as PlaywrightEvent;
        }
    } catch {
        // Skip non-JSON lines
    }
    return null;
}

export function buildStepTree(rawLogs: { begin: string }[]): {
    steps: StepNode[];
    summary: RunSummary;
} {
    const roots: StepNode[] = [];
    const byId = new Map<number, StepNode>();

    let currentGroup: StepNode | null = null;
    let runStart: number | null = null;
    let runEnd: { status: "passed" | "failed"; time: number } | null = null;

    for (const entry of rawLogs) {
        const evt = parseEvent(entry.begin);
        if (!evt || !("timestamp" in evt)) continue;

        const time = new Date(evt.timestamp).toLocaleTimeString("de-DE", {
            hour12: false,
        });

        /* ---------- RUN START ---------- */
        if (evt.type === "run:start") {
            runStart = evt.timestamp;
        }

        /* ---------- STEP START ---------- */
        if (evt.type === "step:start") {
            const node: StepNode = {
                id: evt.stepIndex,
                time,
                label: evt.step,
                kind: isGroup(evt.step) ? "group" : "step",
                status: "running",
                children: [],
            };

            byId.set(evt.stepIndex, node);

            if (node.kind === "group") {
                if (
                    currentGroup &&
                    currentGroup.status === "running" &&
                    !currentGroup.children.some((c) => c.status === "failed")
                ) {
                    currentGroup.status = "done";
                }

                roots.push(node);
                currentGroup = node;
            } else {
                if (currentGroup) currentGroup.children.push(node);
                else roots.push(node);
            }
        }

        if (evt.type === "step:end") {
            const node = byId.get(evt.stepIndex);
            if (!node) continue;

            node.status = evt.status === "failed" ? "failed" : "done";
        }

        if (evt.type === "run:end") {
            runEnd = {
                status: evt.status,
                time: evt.timestamp,
            };

            if (
                currentGroup &&
                currentGroup.status === "running" &&
                !currentGroup.children.some((c) => c.status === "failed")
            ) {
                currentGroup.status = "done";
            }

            roots.push({
                id: Number.MAX_SAFE_INTEGER,
                time,
                label:
                    evt.status === "failed"
                        ? "Test fehlgeschlagen"
                        : "Test abgeschlossen",
                kind: "step",
                status: evt.status === "failed" ? "failed" : "done",
                children: [],
            });
        }
    }

    let summary: RunSummary;

    if (runStart === null) {
        summary = { status: "idle" };
    } else if (runEnd === null) {
        summary = { status: "running" };
    } else {
        summary = {
            status: runEnd.status,
            durationMs: runEnd.time - runStart,
        };
    }

    return { steps: roots, summary };
}
