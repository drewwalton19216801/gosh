#!/usr/local/bin/gosh

test_export() {
    echo "Function called with: $1"
    if [ $# -eq 0 ]; then
        echo "Usage: backup_file <filename>"
        return 1
    fi
    # Use proper variable declarations
    export file="$1"  # Environment variable
    echo "File exported: $file"
    local backup_name="${file}.backup.$(date +%Y%m%d)"  # Local variable
    echo "Backup name: $backup_name"
    cp "$file" "$backup_name"
    echo "Backup created: $backup_name"
}

test_export $1