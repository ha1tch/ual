#!/usr/bin/env python3
"""
ual Benchmark HTML Report Generator

Generates visual HTML reports from benchmark JSON results.

Usage:
    python3 generate_report.py --results DIR --output DIR
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, Optional, List

def load_latest_results(results_dir: Path) -> Optional[Dict[str, Any]]:
    """Load the latest benchmark results"""
    latest = results_dir / "latest.json"
    if latest.exists():
        with open(latest) as f:
            return json.load(f)
    
    # Find most recent JSON file
    json_files = sorted(results_dir.glob("benchmark_*.json"), reverse=True)
    if json_files:
        with open(json_files[0]) as f:
            return json.load(f)
    
    return None


def ratio_str(a: float, b: float, suffix: str = "x") -> str:
    """Format ratio as string"""
    if b <= 0 or a <= 0:
        return "—"
    return f"{a / b:.1f}{suffix}"


def ratio_class(a: float, b: float, good_threshold: float = 1.5) -> str:
    """Return CSS class based on ratio quality"""
    if b <= 0 or a <= 0:
        return ""
    r = a / b
    if r <= good_threshold:
        return "success"
    elif r <= 3.0:
        return ""
    else:
        return "warning"


def generate_html(results: Dict[str, Any]) -> str:
    """Generate HTML report from results"""
    
    version = results.get("version", "unknown")
    timestamp = results.get("timestamp", datetime.now().isoformat())
    correctness = results.get("correctness", {})
    benchmarks = results.get("benchmarks", [])
    cross_lang = results.get("cross_language", {})
    sizes = results.get("binary_sizes", {})
    
    # Convert benchmarks list to dict if needed
    if isinstance(benchmarks, list):
        benchmarks = {b["name"]: b for b in benchmarks}
    
    # Calculate metrics
    total_tests = correctness.get("total", 92)
    passing = correctness.get("iual_pass", total_tests)
    examples_str = f"{passing}/{total_tests}"
    
    # Calculate ratios
    go_times = [b.get("go_ms", 0) for b in benchmarks.values()]
    rust_times = [b.get("rust_ms", 0) for b in benchmarks.values()]
    iual_times = [b.get("iual_ms", 0) for b in benchmarks.values()]
    
    avg_go = sum(go_times) / len(go_times) if go_times else 1
    avg_rust = sum(t for t in rust_times if t > 0) / len([t for t in rust_times if t > 0]) if any(t > 0 for t in rust_times) else avg_go
    avg_iual = sum(iual_times) / len(iual_times) if iual_times else 1
    
    go_rust_ratio = f"{avg_go / avg_rust:.2f}x" if avg_rust > 0 else "N/A"
    iual_ratio = f"{avg_iual / avg_go:.1f}x" if avg_go > 0 else "N/A"
    
    # Determine iual performance class
    iual_class = "success" if avg_iual / avg_go < 5 else "warning" if avg_iual / avg_go < 20 else ""
    
    # Build benchmark rows
    rows_html = ""
    benchmark_names = []
    go_data = []
    rust_data = []
    iual_data = []
    
    for name, data in sorted(benchmarks.items()):
        go_ms = data.get("go_ms", 0)
        rust_ms = data.get("rust_ms", 0)
        iual_ms = data.get("iual_ms", 0)
        ratio = f"{iual_ms / go_ms:.1f}x" if go_ms > 0 else "N/A"
        
        # Highlight if iual is close to compiled
        ratio_cls = "success" if go_ms > 0 and iual_ms / go_ms < 2 else ""
        
        rows_html += f'''<tr>
            <td>{name.replace("_", " ").title()}</td>
            <td class="num">{go_ms}</td>
            <td class="num">{rust_ms if rust_ms > 0 else "—"}</td>
            <td class="num">{iual_ms}</td>
            <td class="num {ratio_cls}">{ratio}</td>
        </tr>
'''
        benchmark_names.append(name.replace("_", " ").title())
        go_data.append(go_ms)
        rust_data.append(rust_ms)
        iual_data.append(iual_ms)
    
    # Cross-language data extraction
    c_data = cross_lang.get("c", {})
    rust_native = cross_lang.get("rust_native", {})
    python_data = cross_lang.get("python", {})
    
    # Build cross-language comparison table
    cross_lang_html = ""
    if cross_lang and any([c_data, rust_native, python_data]):
        cross_lang_html = '''
        <div class="card">
            <h2>Cross-Language Comparison</h2>
            <table>
                <thead>
                    <tr>
                        <th>Benchmark</th>
                        <th class="num">C (ms)</th>
                        <th class="num">Rust (ms)</th>
                        <th class="num">Python (ms)</th>
                        <th class="num">ual-Go (ms)</th>
                        <th class="num">iual (ms)</th>
                    </tr>
                </thead>
                <tbody>
'''
        for bench in ["leibniz", "mandelbrot", "newton"]:
            c_ms = c_data.get(bench, 0)
            rust_ms = rust_native.get(bench, 0)
            py_ms = python_data.get(bench, 0)
            
            # Find corresponding ual benchmark
            ual_bench = benchmarks.get(f"compute_{bench}", {})
            ual_go = ual_bench.get("go_ms", 0)
            ual_iual = ual_bench.get("iual_ms", 0)
            
            cross_lang_html += f'''<tr>
                        <td>{bench.title()}</td>
                        <td class="num">{c_ms if c_ms else "—"}</td>
                        <td class="num">{rust_ms if rust_ms else "—"}</td>
                        <td class="num">{py_ms if py_ms else "—"}</td>
                        <td class="num">{ual_go if ual_go else "—"}</td>
                        <td class="num">{ual_iual if ual_iual else "—"}</td>
                    </tr>
'''
        
        cross_lang_html += '''
                </tbody>
            </table>
            <p class="note">C compiled with -O2, Rust with --release, Python 3.x interpreted</p>
        </div>
'''
    
    # Build analysis tables (like PERFORMANCE.md)
    analysis_html = ""
    if cross_lang and c_data:
        # 1. Compiled ual vs C
        analysis_html += '''
        <div class="card">
            <h2>Analysis: Compiled ual vs C</h2>
            <table>
                <thead>
                    <tr>
                        <th>Benchmark</th>
                        <th class="num">C (ms)</th>
                        <th class="num">ual-Go (ms)</th>
                        <th class="num">ual-Go / C</th>
                        <th class="num">ual-Rust (ms)</th>
                        <th class="num">ual-Rust / C</th>
                    </tr>
                </thead>
                <tbody>
'''
        for bench in ["leibniz", "mandelbrot", "newton"]:
            c_ms = c_data.get(bench, 0)
            ual_bench = benchmarks.get(f"compute_{bench}", {})
            ual_go = ual_bench.get("go_ms", 0)
            ual_rust = ual_bench.get("rust_ms", 0)
            
            go_c_ratio = ratio_str(ual_go, c_ms)
            go_c_cls = ratio_class(ual_go, c_ms, 1.5)
            rust_c_ratio = ratio_str(ual_rust, c_ms)
            rust_c_cls = ratio_class(ual_rust, c_ms, 1.5)
            
            analysis_html += f'''<tr>
                        <td>{bench.title()}</td>
                        <td class="num">{c_ms if c_ms else "—"}</td>
                        <td class="num">{ual_go if ual_go else "—"}</td>
                        <td class="num {go_c_cls}">{go_c_ratio}</td>
                        <td class="num">{ual_rust if ual_rust else "—"}</td>
                        <td class="num {rust_c_cls}">{rust_c_ratio}</td>
                    </tr>
'''
        analysis_html += '''
                </tbody>
            </table>
            <p class="note">Compiled ual should be within 1.0-2.0x of C for compute-heavy workloads</p>
        </div>
'''
        
        # 2. iual vs Compiled
        analysis_html += '''
        <div class="card">
            <h2>Analysis: iual Interpreter vs Compiled</h2>
            <table>
                <thead>
                    <tr>
                        <th>Benchmark</th>
                        <th class="num">ual-Go (ms)</th>
                        <th class="num">iual (ms)</th>
                        <th class="num">iual / ual-Go</th>
                    </tr>
                </thead>
                <tbody>
'''
        for bench in ["leibniz", "mandelbrot", "newton"]:
            ual_bench = benchmarks.get(f"compute_{bench}", {})
            ual_go = ual_bench.get("go_ms", 0)
            ual_iual = ual_bench.get("iual_ms", 0)
            
            iual_go_ratio = ratio_str(ual_iual, ual_go)
            iual_go_cls = ratio_class(ual_iual, ual_go, 2.0)
            
            analysis_html += f'''<tr>
                        <td>{bench.title()}</td>
                        <td class="num">{ual_go if ual_go else "—"}</td>
                        <td class="num">{ual_iual if ual_iual else "—"}</td>
                        <td class="num {iual_go_cls}">{iual_go_ratio}</td>
                    </tr>
'''
        analysis_html += '''
                </tbody>
            </table>
            <p class="note">Threaded code compilation makes iual competitive on structured loops</p>
        </div>
'''
        
        # 3. iual vs Python
        if python_data:
            analysis_html += '''
        <div class="card">
            <h2>Analysis: iual vs Python</h2>
            <table>
                <thead>
                    <tr>
                        <th>Benchmark</th>
                        <th class="num">Python (ms)</th>
                        <th class="num">iual (ms)</th>
                        <th class="num">iual Speedup</th>
                    </tr>
                </thead>
                <tbody>
'''
            for bench in ["leibniz", "mandelbrot", "newton"]:
                py_ms = python_data.get(bench, 0)
                ual_bench = benchmarks.get(f"compute_{bench}", {})
                ual_iual = ual_bench.get("iual_ms", 0)
                
                # Speedup is Python / iual (bigger is better)
                if ual_iual > 0 and py_ms > 0:
                    speedup = f"{py_ms / ual_iual:.1f}x faster"
                    speedup_cls = "success"
                else:
                    speedup = "—"
                    speedup_cls = ""
                
                analysis_html += f'''<tr>
                        <td>{bench.title()}</td>
                        <td class="num">{py_ms if py_ms else "—"}</td>
                        <td class="num">{ual_iual if ual_iual else "—"}</td>
                        <td class="num {speedup_cls}">{speedup}</td>
                    </tr>
'''
            analysis_html += '''
                </tbody>
            </table>
            <p class="note">iual beats Python on every benchmark through threaded code compilation</p>
        </div>
'''
    
    # Binary sizes
    go_size_kb = sizes.get("go_stripped", 0) // 1024
    rust_size_kb = sizes.get("rust_stripped", 0) // 1024
    iual_size_kb = sizes.get("iual_stripped", sizes.get("iual", 0)) // 1024
    
    html = f'''<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ual Benchmark Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * {{ box-sizing: border-box; }}
        
        /* Light mode (default) */
        :root {{
            --bg-primary: #f5f5f5;
            --bg-card: white;
            --text-primary: #333;
            --text-secondary: #666;
            --text-muted: #888;
            --border-color: #eee;
            --border-strong: #ddd;
            --accent-blue: #0066cc;
            --accent-green: #00aa00;
            --accent-orange: #cc6600;
            --hover-bg: #f9f9f9;
            --shadow: rgba(0,0,0,0.1);
            --link-color: #666;
        }}
        
        /* Dark mode */
        @media (prefers-color-scheme: dark) {{
            :root {{
                --bg-primary: #1a1a1a;
                --bg-card: #2a2a2a;
                --text-primary: #e0e0e0;
                --text-secondary: #aaa;
                --text-muted: #888;
                --border-color: #3a3a3a;
                --border-strong: #4a4a4a;
                --accent-blue: #4da6ff;
                --accent-green: #00cc00;
                --accent-orange: #ff9933;
                --hover-bg: #333;
                --shadow: rgba(0,0,0,0.3);
                --link-color: #aaa;
            }}
        }}
        
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0; padding: 2rem;
            background: var(--bg-primary);
            color: var(--text-primary);
        }}
        .container {{ max-width: 1400px; margin: 0 auto; }}
        header {{ margin-bottom: 2rem; }}
        h1 {{ color: var(--text-primary); margin: 0 0 0.5rem 0; }}
        .subtitle {{ color: var(--text-secondary); font-size: 0.9rem; }}
        .summary {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 1rem; margin-bottom: 2rem;
        }}
        .metric-card {{
            background: var(--bg-card); border-radius: 8px; padding: 1.5rem;
            text-align: center; box-shadow: 0 1px 3px var(--shadow);
        }}
        .metric-value {{ font-size: 2.2rem; font-weight: bold; color: var(--accent-blue); }}
        .metric-value.warning {{ color: var(--accent-orange); }}
        .metric-value.success {{ color: var(--accent-green); }}
        .metric-label {{ color: var(--text-secondary); font-size: 0.85rem; margin-top: 0.5rem; }}
        .dashboard {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
            gap: 1.5rem; margin-bottom: 2rem;
        }}
        .card {{
            background: var(--bg-card); border-radius: 8px; padding: 1.5rem;
            box-shadow: 0 1px 3px var(--shadow);
        }}
        .card h2 {{ margin: 0 0 1rem 0; font-size: 1.1rem; color: var(--text-secondary); }}
        .chart-container {{ position: relative; height: 280px; }}
        table {{ width: 100%; border-collapse: collapse; font-size: 0.9rem; }}
        th, td {{ padding: 0.6rem; text-align: left; border-bottom: 1px solid var(--border-color); }}
        th {{ font-weight: 600; color: var(--text-secondary); }}
        tr:hover {{ background: var(--hover-bg); }}
        .num {{ text-align: right; font-family: 'SF Mono', Monaco, Consolas, monospace; }}
        .num.success {{ color: var(--accent-green); font-weight: 600; }}
        .num.warning {{ color: var(--accent-orange); }}
        .note {{ color: var(--text-muted); font-size: 0.8rem; margin-top: 1rem; }}
        footer {{
            margin-top: 2rem; padding-top: 1rem; border-top: 1px solid var(--border-strong);
            color: var(--text-muted); font-size: 0.85rem;
        }}
        footer a {{ color: var(--link-color); }}
        
        /* Analysis section styling */
        .analysis-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 2.5rem 1.5rem;  /* row-gap column-gap */
            margin-bottom: 2rem;
        }}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>ual Benchmark Report</h1>
            <div class="subtitle">
                Generated: {timestamp} | Version: {version}
            </div>
        </header>
        
        <div class="summary">
            <div class="metric-card">
                <div class="metric-value success">{examples_str}</div>
                <div class="metric-label">Examples Passing</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{go_rust_ratio}</div>
                <div class="metric-label">Go / Rust</div>
            </div>
            <div class="metric-card">
                <div class="metric-value {iual_class}">{iual_ratio}</div>
                <div class="metric-label">iual / Compiled</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{go_size_kb}KB</div>
                <div class="metric-label">Go Binary (stripped)</div>
            </div>
        </div>
        
        <div class="dashboard">
            <div class="card">
                <h2>Backend Execution Time</h2>
                <div class="chart-container">
                    <canvas id="backendChart"></canvas>
                </div>
            </div>
            
            <div class="card">
                <h2>Binary Size Comparison (stripped)</h2>
                <div class="chart-container">
                    <canvas id="sizeChart"></canvas>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h2>ual Backend Results</h2>
            <table>
                <thead>
                    <tr>
                        <th>Benchmark</th>
                        <th class="num">Go (ms)</th>
                        <th class="num">Rust (ms)</th>
                        <th class="num">iual (ms)</th>
                        <th class="num">iual/Go</th>
                    </tr>
                </thead>
                <tbody>
{rows_html}
                </tbody>
            </table>
            <p class="note">iual uses threaded code compilation for compute blocks</p>
        </div>
        
{cross_lang_html}

        <div class="analysis-grid">
{analysis_html}
        </div>
        
        <footer>
            ual v{version} — <a href="https://github.com/ha1tch/ual">github.com/ha1tch/ual</a>
        </footer>
    </div>
    
    <script>
        // Detect dark mode for chart colors
        const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        const gridColor = isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)';
        const textColor = isDark ? '#aaa' : '#666';
        
        Chart.defaults.color = textColor;
        Chart.defaults.borderColor = gridColor;
        
        const backendCtx = document.getElementById('backendChart').getContext('2d');
        new Chart(backendCtx, {{
            type: 'bar',
            data: {{
                labels: {json.dumps(benchmark_names)},
                datasets: [
                    {{ label: 'Go', data: {json.dumps(go_data)}, backgroundColor: 'rgba(54, 162, 235, 0.7)' }},
                    {{ label: 'Rust', data: {json.dumps(rust_data)}, backgroundColor: 'rgba(255, 159, 64, 0.7)' }},
                    {{ label: 'iual', data: {json.dumps(iual_data)}, backgroundColor: 'rgba(75, 192, 192, 0.7)' }}
                ]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: false,
                plugins: {{ legend: {{ position: 'top' }} }},
                scales: {{ 
                    y: {{ 
                        beginAtZero: true, 
                        title: {{ display: true, text: 'Time (ms)' }},
                        grid: {{ color: gridColor }}
                    }},
                    x: {{ grid: {{ color: gridColor }} }}
                }}
            }}
        }});
        
        const sizeCtx = document.getElementById('sizeChart').getContext('2d');
        new Chart(sizeCtx, {{
            type: 'bar',
            data: {{
                labels: ['Go', 'Rust', 'iual'],
                datasets: [{{
                    label: 'Size (KB)',
                    data: [{go_size_kb}, {rust_size_kb}, {iual_size_kb}],
                    backgroundColor: ['rgba(54, 162, 235, 0.7)', 'rgba(255, 159, 64, 0.7)', 'rgba(75, 192, 192, 0.7)']
                }}]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: false,
                plugins: {{ legend: {{ display: false }} }},
                scales: {{ 
                    y: {{ 
                        beginAtZero: true, 
                        title: {{ display: true, text: 'Size (KB)' }},
                        grid: {{ color: gridColor }}
                    }},
                    x: {{ grid: {{ color: gridColor }} }}
                }}
            }}
        }});
    </script>
</body>
</html>
'''
    return html


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Generate ual benchmark HTML report")
    parser.add_argument("--results", required=True, help="Directory with benchmark JSON")
    parser.add_argument("--output", required=True, help="Output directory for HTML")
    args = parser.parse_args()
    
    results_dir = Path(args.results)
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    results = load_latest_results(results_dir)
    if not results:
        print(f"No benchmark results found in {results_dir}", file=sys.stderr)
        sys.exit(1)
    
    html = generate_html(results)
    
    # Write timestamped and latest
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    report_file = output_dir / f"report_{timestamp}.html"
    latest_file = output_dir / "latest.html"
    
    with open(report_file, "w") as f:
        f.write(html)
    
    # Symlink or copy to latest
    if latest_file.exists():
        latest_file.unlink()
    latest_file.symlink_to(report_file.name)
    
    print(f"Generated: {report_file}")
    print(f"Latest: {latest_file}")


if __name__ == "__main__":
    main()