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
    # Update version
    content = re.sub(
        r'(  version\s+")[^"]*(")',
        f'\\1{version}\\2',
        content
    )
    
    # Update SHA256 checksums for each platform
    platforms = [
        ('execute-my-will-macos-arm64', 'macOS ARM64'),
        ('execute-my-will-macos-x64', 'macOS x64'),
        ('execute-my-will-linux-arm64', 'Linux ARM64'),
        ('execute-my-will-linux-x64', 'Linux x64')
    ]
    
    for binary_name, platform_name in platforms:
        if binary_name in checksums:
            sha256 = checksums[binary_name]
            print(f"Updating {platform_name} SHA256 to: {sha256}")
            
            # Find and replace the sha256 line for this binary
            # Look for the pattern: binary_name = "execute-my-will-..." followed by sha256
            pattern = rf'(binary_name = "{binary_name}".*?sha256\s+")[^"]*(")'
            
            if re.search(pattern, content, re.DOTALL):
                content = re.sub(pattern, f'\\1{sha256}\\2', content, flags=re.DOTALL)
                print(f"✅ Successfully updated {platform_name} checksum")
            else:
                print(f"⚠️  Warning: Could not find pattern for {platform_name}")
                # Try alternative pattern matching
                alt_pattern = rf'(url.*{binary_name}.*?sha256\s+")[^"]*(")'
                if re.search(alt_pattern, content, re.DOTALL):
                    content = re.sub(alt_pattern, f'\\1{sha256}\\2', content, flags=re.DOTALL)
                    print(f"✅ Successfully updated {platform_name} checksum (alternative pattern)")
                else:
                    print(f"❌ Failed to update {platform_name} checksum")
    
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
    
    # Update formula
    print(f"\n🔧 Updating formula...")
    update_formula(formula_file, version, checksums)
    
    print(f"\n🎉 Formula update completed successfully!")

if __name__ == "__main__":
    main()