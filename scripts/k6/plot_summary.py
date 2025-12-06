#!/usr/bin/env python3
"""
Plot latency percentiles and RPS from a k6 summary JSON.

Usage (with uv):
  uv run scripts/k6/plot_summary.py --input k6-summary.json --output k6-summary.png --duration 60

Requires: matplotlib (uv will install automatically if not present).
"""

import argparse
import json
from pathlib import Path

import matplotlib.pyplot as plt


def load_summary(path: Path) -> dict:
  with path.open(encoding="utf-8") as f:
    return json.load(f)


def extract_percentiles(metric: dict) -> dict:
  vals = metric.get("values", {})
  def nz(key: str) -> float:
    val = vals.get(key)
    return float(val) if val is not None else 0.0

  return {
      "avg": nz("avg"),
      "p50": nz("p(50)"),
      "p90": nz("p(90)"),
      "p95": nz("p(95)"),
      "p99": nz("p(99)"),
  }


def plot_latency(ax, label: str, percentiles: dict):
  order = ["avg", "p50", "p90", "p95", "p99"]
  data = [percentiles.get(k) for k in order]
  ax.bar(order, data, color="#4e79a7")
  ax.set_title(f"{label} (ms)")
  ax.set_ylabel("ms")
  ax.grid(axis="y", linestyle="--", alpha=0.4)


def plot_rps(ax, http_reqs: dict, duration_seconds: float):
  total = http_reqs.get("count", 0)
  rps = total / duration_seconds if duration_seconds > 0 else 0
  ax.bar(["rps"], [rps], color="#f28e2b")
  ax.set_title("Requests per second")
  ax.set_ylabel("req/s")
  ax.grid(axis="y", linestyle="--", alpha=0.4)


def main():
  parser = argparse.ArgumentParser(description="Plot k6 summary JSON to PNG")
  parser.add_argument("--input", default="k6-summary.json", help="Path to k6 summary JSON")
  parser.add_argument("--output", default="k6-summary.png", help="Output PNG path")
  parser.add_argument("--duration", type=float, default=60.0, help="Test duration in seconds (for RPS)")
  args = parser.parse_args()

  summary = load_summary(Path(args.input))
  metrics = summary.get("metrics", {})

  http_req_duration = extract_percentiles(metrics.get("http_req_duration", {}))
  custom_latency = extract_percentiles(metrics.get("suproxy_response_time", {}))
  http_reqs = metrics.get("http_reqs", {})

  fig, axes = plt.subplots(1, 3, figsize=(10, 4))
  plot_latency(axes[0], "http_req_duration", http_req_duration)
  plot_latency(axes[1], "suproxy_response_time", custom_latency)
  plot_rps(axes[2], http_reqs, args.duration)

  plt.tight_layout()
  plt.savefig(args.output, dpi=150)
  print(f"Wrote {args.output}")


if __name__ == "__main__":
  main()

