#!/usr/bin/env python3
import json
import argparse
import matplotlib.pyplot as plt
from pathlib import Path


def load_summary(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def main():
    parser = argparse.ArgumentParser(description="Create ONE PNG containing all k6 load test graphs")
    parser.add_argument("-i", "--input", required=True, type=Path, help="Path to k6-summary.json")
    parser.add_argument("-o", "--output", required=True, type=Path, help="Output PNG file (will be overwritten)")
    args = parser.parse_args()

    if not args.input.exists():
        raise SystemExit(f"Input file not found: {args.input}")

    summary = load_summary(args.input)
    metrics = summary["metrics"]

    # ---- Create combined figure ----
    fig, axes = plt.subplots(1, 3, figsize=(18, 5))  # 3 plots side-by-side

    # === Plot 1: Duration percentiles ===
    dur = metrics["http_req_duration"]
    labels = ["avg", "p50", "p90", "p95", "p99"]
    values = [
        dur.get("avg"),
        dur.get("p(50)"),
        dur.get("p(90)"),
        dur.get("p(95)"),
        dur.get("p(99)"),
    ]
    ax = axes[0]
    ax.bar(labels, values)
    ax.set_ylabel("Duration (ms)")
    ax.set_title("HTTP request duration")
    for i, v in enumerate(values):
        ax.text(i, v, f"{v:.2f}", ha="center", va="bottom", fontsize=8)

    # === Plot 2: Timing breakdown ===
    blocked = metrics["http_req_blocked"]["avg"]
    connecting = metrics["http_req_connecting"]["avg"]
    sending = metrics["http_req_sending"]["avg"]
    waiting = metrics["http_req_waiting"]["avg"]
    receiving = metrics["http_req_receiving"]["avg"]

    labels2 = ["blocked", "connecting", "sending", "waiting", "receiving"]
    values2 = [blocked, connecting, sending, waiting, receiving]

    ax = axes[1]
    ax.bar(labels2, values2)
    ax.set_ylabel("Duration (ms)")
    ax.set_title("HTTP timing breakdown")
    for i, v in enumerate(values2):
        ax.text(i, v, f"{v:.3f}", ha="center", va="bottom", fontsize=7)

    # === Plot 3: RPS ===
    http_reqs = metrics["http_reqs"]
    iterations = metrics.get("iterations", {})

    labels3 = ["http_reqs/s"]
    values3 = [http_reqs["rate"]]

    if "rate" in iterations:
        labels3.append("iterations/s")
        values3.append(iterations["rate"])

    ax = axes[2]
    ax.bar(labels3, values3)
    ax.set_ylabel("Rate (per second)")
    ax.set_title("Load (rates)")
    for i, v in enumerate(values3):
        ax.text(i, v, f"{v:.2f}", ha="center", va="bottom", fontsize=8)

    # ---- Save combined PNG ----
    plt.tight_layout()
    plt.savefig(args.output, dpi=150)
    plt.close()

    print(f"Created PNG: {args.output}")


if __name__ == "__main__":
    main()
