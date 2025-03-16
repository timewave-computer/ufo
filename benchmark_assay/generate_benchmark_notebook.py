#!/usr/bin/env python3
"""
Generate a Jupyter notebook with all benchmark visualizations and data.
This script creates an interactive notebook for exploring and analyzing benchmark results.
"""

import os
import sys
import argparse
import json
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path
import seaborn as sns

def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description='Generate Jupyter notebook from benchmark results.')
    parser.add_argument('csv_file', help='Path to the CSV results file')
    parser.add_argument('output_dir', help='Directory for output notebook and resources')
    parser.add_argument('--report', help='Path to the benchmark report markdown file (optional)')
    parser.add_argument('--run-name', help='Name of the benchmark run (optional)')
    return parser.parse_args()

def create_notebook_cell(cell_type, content, metadata=None):
    """Create a notebook cell with the specified type and content."""
    cell = {
        "cell_type": cell_type,
        "metadata": metadata or {},
        "source": content
    }
    
    if cell_type == "code":
        cell["execution_count"] = None
        cell["outputs"] = []
    
    return cell

def read_data(csv_file):
    """Read data from CSV file."""
    df = pd.read_csv(csv_file)
    return df

def generate_notebook(df, output_dir, report_path=None, run_name=None):
    """Generate a Jupyter notebook with benchmark visualizations and data."""
    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)
    
    # Create visualizations directory
    vis_dir = os.path.join(output_dir, "notebook_visualizations")
    os.makedirs(vis_dir, exist_ok=True)
    
    # Path for the notebook file
    notebook_path = os.path.join(output_dir, "benchmark_analysis.ipynb")
    
    # Check if the DataFrame has configuration and validators columns
    has_config_columns = all(col in df.columns for col in ['configuration', 'validators'])
    
    # Format run name for display
    display_run_name = run_name or "Benchmark Analysis"
    
    # Create notebook cells
    cells = []
    
    # Title and introduction
    cells.append(create_notebook_cell("markdown", [
        f"# {display_run_name}\n\n",
        "This notebook provides an interactive analysis of UFO performance benchmark results. ",
        "It includes visualizations comparing different configurations and raw data for further analysis."
    ]))
    
    # Import libraries cell
    cells.append(create_notebook_cell("code", [
        "import pandas as pd\n",
        "import matplotlib.pyplot as plt\n",
        "import matplotlib.colors as mcolors\n",
        "import numpy as np\n",
        "import seaborn as sns\n",
        "import os\n",
        "\n",
        "# Set up visualization style\n",
        "plt.style.use('ggplot')\n",
        "sns.set_theme(style=\"whitegrid\")\n",
        "%matplotlib inline\n",
        "plt.rcParams['figure.figsize'] = [12, 8]\n",
        "\n",
        "# Create visualizations directory for saving images\n",
        "vis_dir = 'notebook_visualizations'\n",
        "os.makedirs(vis_dir, exist_ok=True)\n",
        "\n",
        "# Define colors for different configurations\n",
        "CONFIG_COLORS = {\n",
        "    'fauxmosis-comet': '#ff7f0e',\n",
        "    'fauxmosis-ufo': '#d62728',\n",
        "    'osmosis-ufo-bridged': '#2ca02c',\n",
        "    'osmosis-ufo-patched': '#9467bd',\n",
        "    'osmosis-comet': '#1f77b4',\n",
        "}\n",
        "\n",
        "CONFIG_NAMES = {\n",
        "    'fauxmosis-comet': 'Fauxmosis with CometBFT',\n",
        "    'fauxmosis-ufo': 'Fauxmosis with UFO',\n",
        "    'osmosis-ufo-bridged': 'Osmosis with UFO (Bridged)',\n",
        "    'osmosis-ufo-patched': 'Osmosis with UFO (Patched)',\n",
        "    'osmosis-comet': 'Osmosis with CometBFT',\n",
        "}"
    ]))
    
    # Load benchmark data
    cells.append(create_notebook_cell("markdown", ["## Benchmark Data\n\nLet's load and explore the benchmark data."]))
    
    # Path to CSV is relative to the notebook
    csv_rel_path = os.path.relpath(os.path.abspath(df.csv_file), output_dir) if hasattr(df, 'csv_file') else "results.csv"
    
    cells.append(create_notebook_cell("code", [
        f"# Load benchmark data from CSV\n",
        f"df = pd.read_csv('{csv_rel_path}')\n",
        "df.head()"
    ]))
    
    # Data exploration
    cells.append(create_notebook_cell("code", [
        "# Show basic statistics\n",
        "df.describe()"
    ]))
    
    if has_config_columns:
        cells.append(create_notebook_cell("code", [
            "# Create a combined configuration+validators column for easier analysis\n",
            "df['config_type'] = df['configuration'] + '-' + df['validators'].astype(str)\n",
            "\n",
            "# Show data for different configurations\n",
            "for config in df['config_type'].unique():\n",
            "    print(f\"\\nConfiguration: {config}\")\n",
            "    print(df[df['config_type'] == config][['block_time', 'tps', 'latency_ms', 'cpu_usage', 'memory_usage']].describe())\n"
        ]))
    
    # Visualizations
    cells.append(create_notebook_cell("markdown", ["## Visualizations\n\nLet's create visualizations to analyze the benchmark results."]))
    
    # Add TPS comparison visualization
    if has_config_columns:
        cells.append(create_notebook_cell("markdown", ["### TPS Comparison\n\nComparing transactions per second (TPS) across different configurations and block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create TPS comparison visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "\n",
            "# Create a color map for configurations\n",
            "# Use the new recommended way to get colormaps\n",
            "cmap = plt.colormaps['tab10']\n",
            "color_dict = {config: cmap(i % 10) for i, config in enumerate(sorted(df['config_type'].unique()))}\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    plt.plot(config_df['block_time'], config_df['tps'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('Transactions Per Second (TPS)', fontsize=12)\n",
            "plt.title('TPS vs Block Time Comparison', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'tps_comparison.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Add Latency comparison visualization
        cells.append(create_notebook_cell("markdown", ["### Latency Comparison\n\nComparing transaction latency across different configurations and block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create Latency comparison visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    plt.plot(config_df['block_time'], config_df['latency_ms'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('Latency (ms)', fontsize=12)\n",
            "plt.title('Latency vs Block Time Comparison', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'latency_comparison.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Add CPU usage comparison visualization
        cells.append(create_notebook_cell("markdown", ["### CPU Usage Comparison\n\nComparing CPU utilization across different configurations and block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create CPU usage comparison visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    plt.plot(config_df['block_time'], config_df['cpu_usage'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('CPU Usage (%)', fontsize=12)\n",
            "plt.title('CPU Usage vs Block Time Comparison', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'cpu_comparison.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Add Memory usage comparison visualization
        cells.append(create_notebook_cell("markdown", ["### Memory Usage Comparison\n\nComparing memory utilization across different configurations and block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create Memory usage comparison visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    plt.plot(config_df['block_time'], config_df['memory_usage'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('Memory Usage (%)', fontsize=12)\n",
            "plt.title('Memory Usage vs Block Time Comparison', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'memory_comparison.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Add Validator impact analysis
        cells.append(create_notebook_cell("markdown", ["### Validator Count Impact\n\nAnalyzing the impact of validator count on performance metrics."]))
        cells.append(create_notebook_cell("code", [
            "# Analyze Impact of Validator Count\n",
            "# Group by configuration and validators\n",
            "validator_impact = df.groupby(['configuration', 'validators'])['tps'].mean().reset_index()\n",
            "validator_impact = validator_impact.pivot(index='configuration', columns='validators', values='tps')\n",
            "validator_impact['impact_pct'] = ((validator_impact[4] - validator_impact[1]) / validator_impact[1]) * 100\n",
            "\n",
            "print(\"Mean TPS by Configuration and Validator Count:\")\n",
            "print(validator_impact)\n",
            "\n",
            "# Plot the impact\n",
            "plt.figure(figsize=(10, 6))\n",
            "bars = plt.bar(validator_impact.index, validator_impact['impact_pct'])\n",
            "\n",
            "# Color the bars based on positive/negative impact\n",
            "for i, bar in enumerate(bars):\n",
            "    if validator_impact['impact_pct'].iloc[i] < 0:\n",
            "        bar.set_color('firebrick')\n",
            "    else:\n",
            "        bar.set_color('forestgreen')\n",
            "\n",
            "plt.axhline(y=0, color='black', linestyle='-', alpha=0.3)\n",
            "plt.ylabel('% Change in TPS (1 → 4 validators)', fontsize=12)\n",
            "plt.title('Impact of Increasing Validator Count from 1 to 4', fontsize=14)\n",
            "plt.grid(True, axis='y', alpha=0.2)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'validator_impact.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Add Comprehensive dashboard
        cells.append(create_notebook_cell("markdown", ["### Comprehensive Dashboard\n\nA comprehensive view of all performance metrics."]))
        cells.append(create_notebook_cell("code", [
            "# Create a comprehensive dashboard with all metrics\n",
            "fig = plt.figure(figsize=(16, 20))\n",
            "\n",
            "# Use a more flexible GridSpec layout with more space between subplots\n",
            "gs = fig.add_gridspec(4, 2, hspace=0.4, wspace=0.3)\n",
            "\n",
            "# TPS comparison (top left)\n",
            "ax1 = fig.add_subplot(gs[0, 0])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax1.plot(config_df['block_time'], config_df['tps'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "ax1.set_xscale('log')\n",
            "ax1.set_xlabel('Block Time (ms)')\n",
            "ax1.set_ylabel('TPS')\n",
            "ax1.set_title('TPS vs Block Time')\n",
            "ax1.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "ax1.legend(fontsize=8)\n",
            "\n",
            "# Latency comparison (top right)\n",
            "ax2 = fig.add_subplot(gs[0, 1])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax2.plot(config_df['block_time'], config_df['latency_ms'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "ax2.set_xscale('log')\n",
            "ax2.set_xlabel('Block Time (ms)')\n",
            "ax2.set_ylabel('Latency (ms)')\n",
            "ax2.set_title('Latency vs Block Time')\n",
            "ax2.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "ax2.legend(fontsize=8)\n",
            "\n",
            "# CPU usage (middle left)\n",
            "ax3 = fig.add_subplot(gs[1, 0])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax3.plot(config_df['block_time'], config_df['cpu_usage'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "ax3.set_xscale('log')\n",
            "ax3.set_xlabel('Block Time (ms)')\n",
            "ax3.set_ylabel('CPU Usage (%)')\n",
            "ax3.set_title('CPU Usage vs Block Time')\n",
            "ax3.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "ax3.legend(fontsize=8)\n",
            "\n",
            "# Memory usage (middle right)\n",
            "ax4 = fig.add_subplot(gs[1, 1])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax4.plot(config_df['block_time'], config_df['memory_usage'], 'o-', \n",
            "             linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "             color=color_dict[config])\n",
            "ax4.set_xscale('log')\n",
            "ax4.set_xlabel('Block Time (ms)')\n",
            "ax4.set_ylabel('Memory Usage (%)')\n",
            "ax4.set_title('Memory Usage vs Block Time')\n",
            "ax4.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "ax4.legend(fontsize=8)\n",
            "\n",
            "# TPS vs CPU Efficiency (bottom left)\n",
            "ax5 = fig.add_subplot(gs[2, 0])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax5.scatter(config_df['cpu_usage'], config_df['tps'], \n",
            "              label=CONFIG_NAMES.get(config, config),\n",
            "              color=color_dict[config], s=100, alpha=0.7)\n",
            "ax5.set_xlabel('CPU Usage (%)')\n",
            "ax5.set_ylabel('TPS')\n",
            "ax5.set_title('Performance Efficiency (TPS vs CPU)')\n",
            "ax5.grid(True, alpha=0.2)\n",
            "ax5.legend(fontsize=8)\n",
            "\n",
            "# Memory vs TPS (bottom right)\n",
            "ax6 = fig.add_subplot(gs[2, 1])\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax6.scatter(config_df['memory_usage'], config_df['tps'], \n",
            "              label=CONFIG_NAMES.get(config, config),\n",
            "              color=color_dict[config], s=100, alpha=0.7)\n",
            "ax6.set_xlabel('Memory Usage (%)')\n",
            "ax6.set_ylabel('TPS')\n",
            "ax6.set_title('Performance Efficiency (TPS vs Memory)')\n",
            "ax6.grid(True, alpha=0.2)\n",
            "ax6.legend(fontsize=8)\n",
            "\n",
            "# Validator impact (bottom)\n",
            "ax7 = fig.add_subplot(gs[3, :])\n",
            "bars = ax7.bar(validator_impact.index, validator_impact['impact_pct'])\n",
            "for i, bar in enumerate(bars):\n",
            "    if validator_impact['impact_pct'].iloc[i] < 0:\n",
            "        bar.set_color('firebrick')\n",
            "    else:\n",
            "        bar.set_color('forestgreen')\n",
            "ax7.axhline(y=0, color='black', linestyle='-', alpha=0.3)\n",
            "ax7.set_ylabel('% Change in TPS (1 → 4 validators)')\n",
            "ax7.set_title('Impact of Increasing Validator Count from 1 to 4')\n",
            "ax7.grid(True, axis='y', alpha=0.2)\n",
            "\n",
            "# Add a main title with fixed position instead of using tight_layout\n",
            "fig.suptitle('UFO Performance Benchmark - Comprehensive Dashboard', fontsize=16, y=0.99)\n",
            "\n",
            "# Adjust the layout with specific padding to avoid warnings\n",
            "fig.subplots_adjust(top=0.95, bottom=0.05, left=0.1, right=0.95)\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'comparative_dashboard.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Interactive analysis section
        cells.append(create_notebook_cell("markdown", ["## Interactive Analysis\n\nLet's create some additional interactive visualizations."]))
        
        # Performance efficiency analysis
        cells.append(create_notebook_cell("code", [
            "# Performance Efficiency: TPS per CPU%\n",
            "df['tps_per_cpu'] = df['tps'] / df['cpu_usage']\n",
            "\n",
            "fig, ax = plt.subplots(figsize=(14, 8))\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    ax.plot(config_df['block_time'], config_df['tps_per_cpu'], 'o-', \n",
            "            linewidth=2, label=CONFIG_NAMES.get(config, config),\n",
            "            color=color_dict[config])\n",
            "\n",
            "ax.set_xscale('log')\n",
            "ax.set_xlabel('Block Time (ms)', fontsize=12)\n",
            "ax.set_ylabel('TPS per CPU % (Efficiency)', fontsize=12)\n",
            "ax.set_title('Processing Efficiency vs Block Time', fontsize=14)\n",
            "ax.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "ax.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "plt.show()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'efficiency_comparison.png'), dpi=300, bbox_inches='tight')"
        ]))
        
        # Memory Utilization analysis
        cells.append(create_notebook_cell("code", [
            "# Memory Utilization vs TPS\n",
            "plt.figure(figsize=(14, 8))\n",
            "\n",
            "for config in sorted(df['config_type'].unique()):\n",
            "    config_df = df[df['config_type'] == config]\n",
            "    plt.scatter(config_df['tps'], config_df['memory_usage'], \n",
            "                label=CONFIG_NAMES.get(config, config),\n",
            "                color=color_dict[config], s=100, alpha=0.7)\n",
            "\n",
            "plt.xlabel('Transactions Per Second (TPS)', fontsize=12)\n",
            "plt.ylabel('Memory Usage (%)', fontsize=12)\n",
            "plt.title('Memory Usage vs TPS by Configuration', fontsize=14)\n",
            "plt.grid(True, alpha=0.2)\n",
            "plt.legend(fontsize=10)\n",
            "plt.tight_layout()\n",
            "plt.show()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'memory_vs_tps.png'), dpi=300, bbox_inches='tight')"
        ]))
        
        # Block time impact heatmap
        cells.append(create_notebook_cell("code", [
            "# Block Time Impact Heatmap\n",
            "# Pivot the data to create a heatmap\n",
            "heatmap_data = df.pivot_table(index='configuration', columns='block_time', values='tps')\n",
            "\n",
            "# Create the heatmap\n",
            "plt.figure(figsize=(14, 8))\n",
            "sns.heatmap(heatmap_data, annot=True, fmt='.1f', cmap='viridis', linewidths=.5)\n",
            "plt.title('TPS by Configuration and Block Time', fontsize=14)\n",
            "plt.ylabel('Configuration', fontsize=12)\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.tight_layout()\n",
            "plt.show()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'blocktime_heatmap.png'), dpi=300, bbox_inches='tight')"
        ]))
    else:
        # Basic visualizations for single configuration
        cells.append(create_notebook_cell("markdown", ["### TPS vs Block Time\n\nVisualization of transactions per second (TPS) at different block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create TPS vs Block Time visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "plt.plot(df['block_time'], df['tps'], 'o-', linewidth=2, color='tab:blue')\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('Transactions Per Second (TPS)', fontsize=12)\n",
            "plt.title('TPS vs Block Time', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'tps_vs_blocktime.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        cells.append(create_notebook_cell("markdown", ["### Latency vs Block Time\n\nVisualization of transaction latency at different block times."]))
        cells.append(create_notebook_cell("code", [
            "# Create Latency vs Block Time visualization\n",
            "plt.figure(figsize=(14, 8))\n",
            "plt.plot(df['block_time'], df['latency_ms'], 'o-', linewidth=2, color='tab:orange')\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('Latency (ms)', fontsize=12)\n",
            "plt.title('Latency vs Block Time', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.tight_layout()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'latency_vs_blocktime.png'), dpi=300, bbox_inches='tight')\n",
            "plt.show()"
        ]))
        
        # Basic analysis for single configuration
        cells.append(create_notebook_cell("code", [
            "# Calculate performance efficiency (TPS per CPU%)\n",
            "df['tps_per_cpu'] = df['tps'] / df['cpu_usage']\n",
            "\n",
            "plt.figure(figsize=(14, 8))\n",
            "plt.plot(df['block_time'], df['tps_per_cpu'], 'o-', linewidth=2, color='tab:green')\n",
            "plt.xscale('log')\n",
            "plt.xlabel('Block Time (ms)', fontsize=12)\n",
            "plt.ylabel('TPS per CPU % (Efficiency)', fontsize=12)\n",
            "plt.title('Processing Efficiency vs Block Time', fontsize=14)\n",
            "plt.grid(True, which=\"both\", ls=\"-\", alpha=0.2)\n",
            "plt.tight_layout()\n",
            "plt.show()\n",
            "\n",
            "# Save the figure\n",
            "plt.savefig(os.path.join(vis_dir, 'efficiency_vs_blocktime.png'), dpi=300, bbox_inches='tight')"
        ]))
    
    # Add benchmark report if provided
    if report_path and os.path.exists(report_path):
        cells.append(create_notebook_cell("markdown", ["## Benchmark Report\n\nBelow is the full benchmark report."]))
        
        try:
            with open(report_path, 'r') as f:
                report_content = f.read()
            
            cells.append(create_notebook_cell("markdown", [report_content]))
        except Exception as e:
            cells.append(create_notebook_cell("markdown", [f"Error loading benchmark report: {e}"]))
    
    # Conclusion
    cells.append(create_notebook_cell("markdown", [
        "## Conclusion\n\n",
        "This notebook provides an analysis of the UFO benchmark results, comparing performance across different configurations ",
        "and block times. The visualizations show how transaction throughput, latency, and resource usage are affected by ",
        "block time settings and validator counts.\n\n",
        "Key metrics to consider:\n\n",
        "* **TPS (Transactions Per Second)**: Higher is better, indicates throughput\n",
        "* **Latency**: Lower is better, indicates responsiveness\n",
        "* **Resource Usage**: Lower CPU and memory usage for the same TPS indicates better efficiency\n",
        "* **TPS per CPU %**: Higher is better, indicates better performance efficiency\n",
        "\n",
        "All visualizations are saved to the `notebook_visualizations` directory for reference."
    ]))
    
    # Create the notebook
    notebook = {
        "cells": cells,
        "metadata": {
            "kernelspec": {
                "display_name": "Python 3",
                "language": "python",
                "name": "python3"
            },
            "language_info": {
                "codemirror_mode": {
                    "name": "ipython",
                    "version": 3
                },
                "file_extension": ".py",
                "mimetype": "text/x-python",
                "name": "python",
                "nbconvert_exporter": "python",
                "pygments_lexer": "ipython3",
                "version": "3.8.10"
            }
        },
        "nbformat": 4,
        "nbformat_minor": 4
    }
    
    # Write the notebook to a file
    with open(notebook_path, 'w') as f:
        json.dump(notebook, f, indent=2)
    
    print(f"Jupyter notebook created at: {notebook_path}")
    print(f"Visualizations will be saved to: {vis_dir} when the notebook is run")
    
    return notebook_path

def main():
    args = parse_args()
    
    # Read the benchmark data
    df = read_data(args.csv_file)
    df.csv_file = args.csv_file  # Store the CSV path for later use
    
    # Generate the notebook
    notebook_path = generate_notebook(
        df, 
        args.output_dir, 
        report_path=args.report, 
        run_name=args.run_name
    )
    
    print(f"Notebook created successfully. To view it, run:")
    print(f"jupyter notebook {notebook_path}")

if __name__ == "__main__":
    main() 