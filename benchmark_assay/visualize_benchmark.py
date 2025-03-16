#!/usr/bin/env python3
"""
Visualization tool for UFO benchmark results.
This script generates charts showing performance metrics at different block times.
It can also create comparative visualizations for different configurations.
"""

import os
import sys
import argparse
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path

# Style configuration
plt.style.use('ggplot')
COLORS = {
    'fauxmosis-comet-1': '#ff7f0e',       # Orange
    'fauxmosis-comet-4': '#ff9a41',       # Light Orange
    'fauxmosis-ufo-1': '#d62728',         # Red
    'fauxmosis-ufo-4': '#ff6b6b',         # Light Red
    'osmosis-ufo-bridged-1': '#2ca02c',   # Green
    'osmosis-ufo-bridged-4': '#5fd35f',   # Light Green
    'osmosis-ufo-patched-1': '#9467bd',   # Purple
    'osmosis-ufo-patched-4': '#b589dc',   # Light Purple
    'osmosis-comet-1': '#1f77b4',         # Blue
    'osmosis-comet-4': '#5aa7d4',         # Light Blue
}

# Configuration labels for legends
CONFIG_LABELS = {
    'fauxmosis-comet-1': 'Fauxmosis+CometBFT (1 validator)',
    'fauxmosis-comet-4': 'Fauxmosis+CometBFT (4 validators)',
    'fauxmosis-ufo-1': 'Fauxmosis+UFO (1 validator)',
    'fauxmosis-ufo-4': 'Fauxmosis+UFO (4 validators)',
    'osmosis-ufo-bridged-1': 'Osmosis+UFO Bridged (1 validator)',
    'osmosis-ufo-bridged-4': 'Osmosis+UFO Bridged (4 validators)',
    'osmosis-ufo-patched-1': 'Osmosis+UFO Patched (1 validator)',
    'osmosis-ufo-patched-4': 'Osmosis+UFO Patched (4 validators)',
    'osmosis-comet-1': 'Osmosis+CometBFT (1 validator)',
    'osmosis-comet-4': 'Osmosis+CometBFT (4 validators)',
}

def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description='Visualize benchmark results.')
    parser.add_argument('csv_file', help='Path to the CSV results file')
    parser.add_argument('output_dir', help='Directory for output visualizations')
    parser.add_argument('--comparative', action='store_true', help='Generate comparative visualizations')
    return parser.parse_args()

def read_data(csv_file):
    """Read benchmark data from CSV file."""
    # Check if file exists
    if not os.path.exists(csv_file):
        print(f"Error: CSV file not found: {csv_file}")
        sys.exit(1)
    
    # Read the CSV file
    try:
        df = pd.read_csv(csv_file)
        
        # Check if the dataframe is empty
        if df.empty:
            print("Error: CSV file is empty")
            sys.exit(1)
        
        return df
    except Exception as e:
        print(f"Error reading CSV file: {e}")
        sys.exit(1)

def create_single_config_visualizations(df, output_dir):
    """Create visualizations for a single configuration."""
    os.makedirs(output_dir, exist_ok=True)
    
    # Sort by block time for consistent x-axis
    df = df.sort_values('block_time')
    
    # Create visualizations for TPS vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['tps'], 'o-', linewidth=2)
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Transactions Per Second (TPS)', fontsize=12)
    plt.title('TPS vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'tps_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create visualizations for latency vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['latency_ms'], 'o-', linewidth=2, color='green')
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Transaction Latency (ms)', fontsize=12)
    plt.title('Latency vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'latency_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create visualizations for CPU usage vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['cpu_usage'], 'o-', linewidth=2, color='red')
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('CPU Usage (%)', fontsize=12)
    plt.title('CPU Usage vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'cpu_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create visualizations for memory usage vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['memory_usage'], 'o-', linewidth=2, color='purple')
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Memory Usage (%)', fontsize=12)
    plt.title('Memory Usage vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'memory_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create visualizations for blocks produced vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['blocks_produced'], 'o-', linewidth=2, color='brown')
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Blocks Produced', fontsize=12)
    plt.title('Blocks Produced vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'blocks_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create visualizations for transactions per block vs block time
    plt.figure(figsize=(10, 6))
    plt.plot(df['block_time'], df['avg_tx_per_block'], 'o-', linewidth=2, color='orange')
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Average Transactions Per Block', fontsize=12)
    plt.title('Transactions Per Block vs Block Time', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'tx_per_block_vs_blocktime.png'), dpi=300)
    plt.close()
    
    # Create a combined metrics visualization
    fig, axs = plt.subplots(2, 2, figsize=(15, 10))
    
    # TPS vs block time
    axs[0, 0].plot(df['block_time'], df['tps'], 'o-', linewidth=2)
    axs[0, 0].set_xscale('log')
    axs[0, 0].set_xlabel('Block Time (ms)')
    axs[0, 0].set_ylabel('TPS')
    axs[0, 0].set_title('TPS vs Block Time')
    axs[0, 0].grid(True, which="both", ls="-", alpha=0.2)
    
    # Latency vs block time
    axs[0, 1].plot(df['block_time'], df['latency_ms'], 'o-', linewidth=2, color='green')
    axs[0, 1].set_xscale('log')
    axs[0, 1].set_xlabel('Block Time (ms)')
    axs[0, 1].set_ylabel('Latency (ms)')
    axs[0, 1].set_title('Latency vs Block Time')
    axs[0, 1].grid(True, which="both", ls="-", alpha=0.2)
    
    # CPU & Memory usage vs block time
    axs[1, 0].plot(df['block_time'], df['cpu_usage'], 'o-', linewidth=2, color='red', label='CPU')
    axs[1, 0].plot(df['block_time'], df['memory_usage'], 'o-', linewidth=2, color='purple', label='Memory')
    axs[1, 0].set_xscale('log')
    axs[1, 0].set_xlabel('Block Time (ms)')
    axs[1, 0].set_ylabel('Usage (%)')
    axs[1, 0].set_title('Resource Usage vs Block Time')
    axs[1, 0].grid(True, which="both", ls="-", alpha=0.2)
    axs[1, 0].legend()
    
    # Transactions per block vs block time
    axs[1, 1].plot(df['block_time'], df['avg_tx_per_block'], 'o-', linewidth=2, color='orange')
    axs[1, 1].set_xscale('log')
    axs[1, 1].set_xlabel('Block Time (ms)')
    axs[1, 1].set_ylabel('Avg Tx/Block')
    axs[1, 1].set_title('Transactions Per Block vs Block Time')
    axs[1, 1].grid(True, which="both", ls="-", alpha=0.2)
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'combined_metrics.png'), dpi=300)
    plt.close()
    
    # Create a performance dashboard
    fig = plt.figure(figsize=(16, 10))
    
    # Set up grid layout
    gs = fig.add_gridspec(2, 3)
    
    # TPS vs block time (larger plot)
    ax1 = fig.add_subplot(gs[0, :2])
    ax1.plot(df['block_time'], df['tps'], 'o-', linewidth=2)
    ax1.set_xscale('log')
    ax1.set_xlabel('Block Time (ms)', fontsize=12)
    ax1.set_ylabel('TPS', fontsize=12)
    ax1.set_title('Transactions Per Second', fontsize=14)
    ax1.grid(True, which="both", ls="-", alpha=0.2)
    
    # Latency vs block time
    ax2 = fig.add_subplot(gs[0, 2])
    ax2.plot(df['block_time'], df['latency_ms'], 'o-', linewidth=2, color='green')
    ax2.set_xscale('log')
    ax2.set_xlabel('Block Time (ms)', fontsize=12)
    ax2.set_ylabel('Latency (ms)', fontsize=12)
    ax2.set_title('Transaction Latency', fontsize=14)
    ax2.grid(True, which="both", ls="-", alpha=0.2)
    
    # CPU usage vs block time
    ax3 = fig.add_subplot(gs[1, 0])
    ax3.plot(df['block_time'], df['cpu_usage'], 'o-', linewidth=2, color='red')
    ax3.set_xscale('log')
    ax3.set_xlabel('Block Time (ms)', fontsize=12)
    ax3.set_ylabel('CPU Usage (%)', fontsize=12)
    ax3.set_title('CPU Utilization', fontsize=14)
    ax3.grid(True, which="both", ls="-", alpha=0.2)
    
    # Memory usage vs block time
    ax4 = fig.add_subplot(gs[1, 1])
    ax4.plot(df['block_time'], df['memory_usage'], 'o-', linewidth=2, color='purple')
    ax4.set_xscale('log')
    ax4.set_xlabel('Block Time (ms)', fontsize=12)
    ax4.set_ylabel('Memory Usage (%)', fontsize=12)
    ax4.set_title('Memory Utilization', fontsize=14)
    ax4.grid(True, which="both", ls="-", alpha=0.2)
    
    # Transactions per block vs block time
    ax5 = fig.add_subplot(gs[1, 2])
    ax5.plot(df['block_time'], df['avg_tx_per_block'], 'o-', linewidth=2, color='orange')
    ax5.set_xscale('log')
    ax5.set_xlabel('Block Time (ms)', fontsize=12)
    ax5.set_ylabel('Avg Tx/Block', fontsize=12)
    ax5.set_title('Transactions Per Block vs Block Time')
    ax5.grid(True, which="both", ls="-", alpha=0.2)
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'performance_dashboard.png'), dpi=300)
    plt.close()
    
    print(f"Visualizations created in {output_dir}")

def create_comparative_visualizations(df, output_dir):
    """Create visualizations comparing different configurations."""
    os.makedirs(output_dir, exist_ok=True)
    
    # Add a combined configuration+validators column for easier plotting
    df['config_type'] = df['configuration'] + '-' + df['validators'].astype(str)
    
    # Get unique configurations for plotting
    configs = df['config_type'].unique()
    
    # Sort by block time within each configuration
    df = df.sort_values(['config_type', 'block_time'])
    
    # Create TPS comparison visualization
    plt.figure(figsize=(12, 8))
    for config in configs:
        config_df = df[df['config_type'] == config]
        plt.plot(config_df['block_time'], config_df['tps'], 'o-', 
                 linewidth=2, label=CONFIG_LABELS.get(config, config),
                 color=COLORS.get(config, None))
    
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Transactions Per Second (TPS)', fontsize=12)
    plt.title('TPS Comparison Across Configurations', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.legend(fontsize=10)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'tps_comparison.png'), dpi=300)
    plt.close()
    
    # Create Latency comparison visualization
    plt.figure(figsize=(12, 8))
    for config in configs:
        config_df = df[df['config_type'] == config]
        plt.plot(config_df['block_time'], config_df['latency_ms'], 'o-', 
                 linewidth=2, label=CONFIG_LABELS.get(config, config),
                 color=COLORS.get(config, None))
    
    plt.xscale('log')
    plt.yscale('log')  # Using log scale for latency to better show differences
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Transaction Latency (ms)', fontsize=12)
    plt.title('Latency Comparison Across Configurations', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.legend(fontsize=10)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'latency_comparison.png'), dpi=300)
    plt.close()
    
    # Create CPU usage comparison visualization
    plt.figure(figsize=(12, 8))
    for config in configs:
        config_df = df[df['config_type'] == config]
        plt.plot(config_df['block_time'], config_df['cpu_usage'], 'o-', 
                 linewidth=2, label=CONFIG_LABELS.get(config, config),
                 color=COLORS.get(config, None))
    
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('CPU Usage (%)', fontsize=12)
    plt.title('CPU Usage Comparison Across Configurations', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.legend(fontsize=10)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'cpu_comparison.png'), dpi=300)
    plt.close()
    
    # Create Memory usage comparison visualization
    plt.figure(figsize=(12, 8))
    for config in configs:
        config_df = df[df['config_type'] == config]
        plt.plot(config_df['block_time'], config_df['memory_usage'], 'o-', 
                 linewidth=2, label=CONFIG_LABELS.get(config, config),
                 color=COLORS.get(config, None))
    
    plt.xscale('log')
    plt.xlabel('Block Time (ms)', fontsize=12)
    plt.ylabel('Memory Usage (%)', fontsize=12)
    plt.title('Memory Usage Comparison Across Configurations', fontsize=14)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    plt.legend(fontsize=10)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'memory_comparison.png'), dpi=300)
    plt.close()
    
    # Create Validator Count Impact visualization
    # For this, we'll compare 1 vs 4 validators for each type at specific block times
    # First, let's choose representative block times
    representative_block_times = [1000, 100, 10]
    
    # Get the distinct configuration types (without validator count)
    config_types = df['configuration'].unique()
    
    # Create a plot for each representative block time
    fig, axs = plt.subplots(1, len(representative_block_times), figsize=(18, 6))
    
    for i, block_time in enumerate(representative_block_times):
        # Filter for the block time of interest (with a small tolerance)
        block_df = df[df['block_time'].between(block_time*0.9, block_time*1.1)]
        
        # Set up a bar position counter
        bar_positions = np.arange(len(config_types))
        bar_width = 0.35
        
        # Plot bars for 1 validator
        vals_1 = [block_df[(block_df['configuration'] == config) & 
                           (block_df['validators'] == 1)]['tps'].values[0] 
                 if not block_df[(block_df['configuration'] == config) & 
                                (block_df['validators'] == 1)].empty else 0 
                 for config in config_types]
        
        # Plot bars for 4 validators
        vals_4 = [block_df[(block_df['configuration'] == config) & 
                           (block_df['validators'] == 4)]['tps'].values[0] 
                 if not block_df[(block_df['configuration'] == config) & 
                                (block_df['validators'] == 4)].empty else 0 
                 for config in config_types]
        
        axs[i].bar(bar_positions - bar_width/2, vals_1, bar_width, 
                  label='1 Validator', alpha=0.7)
        axs[i].bar(bar_positions + bar_width/2, vals_4, bar_width, 
                  label='4 Validators', alpha=0.7)
        
        axs[i].set_title(f'Block Time: {block_time}ms')
        axs[i].set_xticks(bar_positions)
        axs[i].set_xticklabels([c.capitalize() for c in config_types], rotation=45)
        axs[i].set_ylabel('TPS')
        
        # Only add legend to the first subplot
        if i == 0:
            axs[i].legend()
    
    plt.suptitle('Validator Count Impact on TPS', fontsize=16)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'validator_impact.png'), dpi=300)
    plt.close()
    
    # Create a comprehensive comparative dashboard
    fig = plt.figure(figsize=(18, 12))
    gs = fig.add_gridspec(3, 2)
    
    # TPS comparison (larger plot)
    ax1 = fig.add_subplot(gs[0, :])
    for config in configs:
        config_df = df[df['config_type'] == config]
        ax1.plot(config_df['block_time'], config_df['tps'], 'o-', 
                linewidth=2, label=CONFIG_LABELS.get(config, config),
                color=COLORS.get(config, None))
    
    ax1.set_xscale('log')
    ax1.set_xlabel('Block Time (ms)', fontsize=12)
    ax1.set_ylabel('TPS', fontsize=12)
    ax1.set_title('TPS Comparison', fontsize=14)
    ax1.grid(True, which="both", ls="-", alpha=0.2)
    ax1.legend(fontsize=10, loc='upper left')
    
    # Latency comparison
    ax2 = fig.add_subplot(gs[1, 0])
    for config in configs:
        config_df = df[df['config_type'] == config]
        ax2.plot(config_df['block_time'], config_df['latency_ms'], 'o-', 
                linewidth=2, label=CONFIG_LABELS.get(config, config),
                color=COLORS.get(config, None))
    
    ax2.set_xscale('log')
    ax2.set_yscale('log')
    ax2.set_xlabel('Block Time (ms)', fontsize=12)
    ax2.set_ylabel('Latency (ms)', fontsize=12)
    ax2.set_title('Latency Comparison', fontsize=14)
    ax2.grid(True, which="both", ls="-", alpha=0.2)
    
    # CPU & Memory usage comparison
    ax3 = fig.add_subplot(gs[1, 1])
    
    # Set up a bar position counter for average CPU usage across all block times
    bar_positions = np.arange(len(configs))
    bar_width = 0.35
    
    # Calculate average CPU and memory usage for each configuration
    avg_cpu = [df[df['config_type'] == config]['cpu_usage'].mean() for config in configs]
    avg_mem = [df[df['config_type'] == config]['memory_usage'].mean() for config in configs]
    
    ax3.bar(bar_positions - bar_width/2, avg_cpu, bar_width, label='CPU Usage', color='red', alpha=0.7)
    ax3.bar(bar_positions + bar_width/2, avg_mem, bar_width, label='Memory Usage', color='purple', alpha=0.7)
    
    ax3.set_xticks(bar_positions)
    ax3.set_xticklabels([CONFIG_LABELS.get(c, c) for c in configs], rotation=45, ha='right')
    ax3.set_ylabel('Average Usage (%)')
    ax3.set_title('Resource Usage Comparison', fontsize=14)
    ax3.grid(True, which="both", ls="-", alpha=0.2)
    ax3.legend()
    
    # Validator impact for each configuration type
    ax4 = fig.add_subplot(gs[2, :])
    
    # Get the distinct configuration types (without validator count)
    config_types = df['configuration'].unique()
    
    # Choose a representative block time to show validator impact
    rep_block_time = 100
    
    # Filter for the block time of interest (with a small tolerance)
    block_df = df[df['block_time'].between(rep_block_time*0.9, rep_block_time*1.1)]
    
    # Set up bar positions
    bar_positions = np.arange(len(config_types))
    bar_width = 0.35
    
    # Plot bars for 1 validator
    vals_1 = [block_df[(block_df['configuration'] == config) & 
                       (block_df['validators'] == 1)]['tps'].values[0] 
             if not block_df[(block_df['configuration'] == config) & 
                            (block_df['validators'] == 1)].empty else 0 
             for config in config_types]
    
    # Plot bars for 4 validators
    vals_4 = [block_df[(block_df['configuration'] == config) & 
                       (block_df['validators'] == 4)]['tps'].values[0] 
             if not block_df[(block_df['configuration'] == config) & 
                            (block_df['validators'] == 4)].empty else 0 
             for config in config_types]
    
    ax4.bar(bar_positions - bar_width/2, vals_1, bar_width, 
            label='1 Validator', alpha=0.7, color='green')
    ax4.bar(bar_positions + bar_width/2, vals_4, bar_width, 
            label='4 Validators', alpha=0.7, color='blue')
    
    ax4.set_title(f'Validator Count Impact on TPS at {rep_block_time}ms Block Time', fontsize=14)
    ax4.set_xticks(bar_positions)
    ax4.set_xticklabels([c.capitalize() for c in config_types])
    ax4.set_ylabel('TPS')
    ax4.grid(True, which="both", ls="-", alpha=0.2)
    ax4.legend()
    
    plt.suptitle('Comprehensive Performance Comparison', fontsize=16)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'comparative_dashboard.png'), dpi=300)
    plt.close()
    
    print(f"Comparative visualizations created in {output_dir}")

def main():
    args = parse_args()
    
    # Read the benchmark data
    df = read_data(args.csv_file)
    
    # Ensure output directory exists
    os.makedirs(args.output_dir, exist_ok=True)
    
    if args.comparative:
        # Create comparative visualizations
        create_comparative_visualizations(df, args.output_dir)
    else:
        # Create standard single-configuration visualizations
        create_single_config_visualizations(df, args.output_dir)

if __name__ == "__main__":
    main() 