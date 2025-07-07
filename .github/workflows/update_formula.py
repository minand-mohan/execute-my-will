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
        content = f.read()
    
    print(f"Updating formula version to: {version}")
    
    # Update version line
    content = re.sub(
        r'(\s+version\s+")[^"]*(")',
        f'\\g<1>{version}\\g<2>',
        content
    )
    print(f"✅ Updated version to: {version}")
    
    # Define the mapping of binary names to their expected checksums
    binary_checksums = {
        'execute-my-will-macos-arm64': checksums.get('execute-my-will-macos-arm64', ''),
        'execute-my-will-macos-x64': checksums.get('execute-my-will-macos-x64', ''),
        'execute-my-will-linux-arm64': checksums.get('execute-my-will-linux-arm64', ''),
        'execute-my-will-linux-x64': checksums.get('execute-my-will-linux-x64', '')
    }
    
    # Update each SHA256 checksum using regex patterns that match the specific context
    
    # macOS ARM64 - matches the sha256 line that comes after the macos-arm64 binary assignment
    if binary_checksums['execute-my-will-macos-arm64']:
        pattern = r'(binary_name = "execute-my-will-macos-arm64".*?url.*?sha256\s+")[^"]*(")'
        replacement = f'\\g<1>{binary_checksums["execute-my-will-macos-arm64"]}\\g<2>'
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        print(f"✅ Updated macOS ARM64 SHA256 to: {binary_checksums['execute-my-will-macos-arm64']}")
    
    # macOS x64 - matches the sha256 line that comes after the macos-x64 binary assignment  
    if binary_checksums['execute-my-will-macos-x64']:
        pattern = r'(binary_name = "execute-my-will-macos-x64".*?url.*?sha256\s+")[^"]*(")'
        replacement = f'\\g<1>{binary_checksums["execute-my-will-macos-x64"]}\\g<2>'
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        print(f"✅ Updated macOS x64 SHA256 to: {binary_checksums['execute-my-will-macos-x64']}")
    
    # Linux ARM64 - matches the sha256 line that comes after the linux-arm64 binary assignment
    if binary_checksums['execute-my-will-linux-arm64']:
        pattern = r'(binary_name = "execute-my-will-linux-arm64".*?url.*?sha256\s+")[^"]*(")'
        replacement = f'\\g<1>{binary_checksums["execute-my-will-linux-arm64"]}\\g<2>'
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        print(f"✅ Updated Linux ARM64 SHA256 to: {binary_checksums['execute-my-will-linux-arm64']}")
    
    # Linux x64 - matches the sha256 line that comes after the linux-x64 binary assignment
    if binary_checksums['execute-my-will-linux-x64']:
        pattern = r'(binary_name = "execute-my-will-linux-x64".*?url.*?sha256\s+")[^"]*(")'
        replacement = f'\\g<1>{binary_checksums["execute-my-will-linux-x64"]}\\g<2>'
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        print(f"✅ Updated Linux x64 SHA256 to: {binary_checksums['execute-my-will-linux-x64']}")
    
    # Write the updated content back
    with open(formula_file, 'w') as f:
        f.write(content)
    
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
    
    # Verify we have all required checksums
    required_binaries = [
        'execute-my-will-macos-arm64',
        'execute-my-will-macos-x64',
        'execute-my-will-linux-arm64',
        'execute-my-will-linux-x64'
    ]
    
    missing_checksums = []
    for binary in required_binaries:
        if binary not in checksums:
            missing_checksums.append(binary)
    
    if missing_checksums:
        print(f"⚠️  Warning: Missing checksums for: {', '.join(missing_checksums)}")
        print("Available checksums:")
        for binary, checksum in checksums.items():
            print(f"  - {binary}: {checksum}")
    
    # Update formula
    print(f"\n🔧 Updating formula...")
    update_formula(formula_file, version, checksums)
    
    print(f"\n🎉 Formula update completed successfully!")

if __name__ == "__main__":
    main()