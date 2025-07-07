#!/usr/bin/env python3
"""
Script to update Homebrew formula with new version and SHA256 checksums.
Usage: python update_formula.py <formula_file> <version> <checksums_dir>
"""

import sys
import os
import re
import glob
from pathlib import Path

def read_checksums(checksums_dir):
    """Read all SHA256 checksums from the artifacts directory."""
    checksums = {}
    
    # Find all .sha256 files in the artifacts directory
    sha_files = glob.glob(os.path.join(checksums_dir, "**", "*.sha256"), recursive=True)
    
    for sha_file in sha_files:
        with open(sha_file, 'r') as f:
            content = f.read().strip()
            # Extract SHA256 hash (first part before space)
            sha256_hash = content.split()[0]
            # Get binary name from filename (remove .sha256 extension)
            binary_name = os.path.basename(sha_file).replace('.sha256', '')
            checksums[binary_name] = sha256_hash
            print(f"Found checksum for {binary_name}: {sha256_hash}")
    
    return checksums

def update_formula(formula_file, version, checksums):
    """Update the Homebrew formula with new version and checksums."""
    
    with open(formula_file, 'r') as f:
        lines = f.readlines()
    
    print(f"Updating formula version to: {version}")
    
    # Update version line
    for i, line in enumerate(lines):
        if line.strip().startswith('version '):
            lines[i] = f'  version "{version}"\n'
            print(f"✅ Updated version line: {lines[i].strip()}")
            break
    
    # Update SHA256 checksums
    # We'll track which platform we're currently in
    current_platform = None
    platform_map = {
        'macos-arm64': 'execute-my-will-macos-arm64',
        'macos-x64': 'execute-my-will-macos-x64', 
        'linux-arm64': 'execute-my-will-linux-arm64',
        'linux-x64': 'execute-my-will-linux-x64'
    }
    
    i = 0
    while i < len(lines):
        line = lines[i].strip()
        
        # Detect which platform section we're in
        if 'if OS.mac?' in line:
            current_platform = 'mac'
        elif 'elif OS.linux?' in line:
            current_platform = 'linux'
        elif 'if Hardware::CPU.arm?' in line:
            if current_platform == 'mac':
                current_platform = 'macos-arm64'
            elif current_platform == 'linux':
                current_platform = 'linux-arm64'
        elif line.startswith('else') and current_platform:
            if current_platform == 'macos-arm64':
                current_platform = 'macos-x64'
            elif current_platform == 'linux-arm64':
                current_platform = 'linux-x64'
        
        # Update SHA256 if we find one in the current platform
        if line.startswith('sha256 ') and current_platform in platform_map:
            binary_name = platform_map[current_platform]
            if binary_name in checksums:
                new_sha256 = checksums[binary_name]
                lines[i] = f'      sha256 "{new_sha256}"\n'
                print(f"✅ Updated {current_platform} SHA256 to: {new_sha256}")
            else:
                print(f"⚠️  No checksum found for {binary_name}")
        
        i += 1
    
    # Write the updated content back
    with open(formula_file, 'w') as f:
        f.writelines(lines)
    
    print(f"✅ Successfully updated {formula_file}")

def validate_inputs(formula_file, version, checksums_dir):
    """Validate all inputs before processing."""
    errors = []
    
    if not os.path.exists(formula_file):
        errors.append(f"Formula file {formula_file} not found")
    
    if not version or not version.strip():
        errors.append("Version cannot be empty")
    
    if not os.path.exists(checksums_dir):
        errors.append(f"Checksums directory {checksums_dir} not found")
    
    return errors

def main():
    if len(sys.argv) != 4:
        print("Usage: python update_formula.py <formula_file> <version> <checksums_dir>")
        print("Example: python update_formula.py Formula/execute-my-will.rb 1.0.0 ./artifacts")
        sys.exit(1)
    
    formula_file = sys.argv[1]
    version = sys.argv[2]
    checksums_dir = sys.argv[3]
    
    print(f"🔄 Starting formula update...")
    print(f"📁 Formula file: {formula_file}")
    print(f"🏷️  Version: {version}")
    print(f"📂 Checksums directory: {checksums_dir}")
    
    # Validate inputs
    errors = validate_inputs(formula_file, version, checksums_dir)
    if errors:
        print("❌ Validation errors:")
        for error in errors:
            print(f"  - {error}")
        sys.exit(1)
    
    # Read checksums
    print(f"\n📋 Reading checksums from {checksums_dir}...")
    checksums = read_checksums(checksums_dir)
    
    if not checksums:
        print("❌ Error: No checksums found")
        sys.exit(1)
    
    print(f"✅ Found {len(checksums)} checksums")
    
    # Update formula
    print(f"\n🔧 Updating formula...")
    update_formula(formula_file, version, checksums)
    
    print(f"\n🎉 Formula update completed successfully!")

if __name__ == "__main__":
    main()