# Cross-Platform Features

Gosh is designed to work seamlessly across Windows, macOS, and Linux, providing a consistent shell experience regardless of your operating system. This document details the cross-platform features and compatibility considerations.

## Dual Command Support

One of gosh's key features is support for both Unix-style and Windows-style command names. This means you can use the commands you're familiar with, regardless of your background:

### File Listing
- **Unix-style**: `ls [options] [path...]`
- **Windows-style**: `dir [options] [path...]`

Both commands support the same options and provide identical functionality:
```bash
ls -la          # Unix-style long listing with hidden files
dir -la         # Windows-style equivalent
```

### File Display
- **Unix-style**: `cat [file...]`
- **Windows-style**: `type [file...]`

```bash
cat README.md   # Unix-style file display
type README.md  # Windows-style equivalent
```

### File Operations
- **Copy**: `cp` (Unix) / `copy` (Windows)
- **Move**: `mv` (Unix) / `move` (Windows)
- **Remove**: `rm` (Unix) / `del` (Windows)

```bash
# Unix-style
cp file.txt backup.txt
mv oldname.txt newname.txt
rm unwanted.txt

# Windows-style
copy file.txt backup.txt
move oldname.txt newname.txt
del unwanted.txt
```

### Directory Operations
- **Create**: `mkdir` (Unix) / `md` (Windows)
- **Remove**: `rmdir` (Unix) / `rd` (Windows)

```bash
# Unix-style
mkdir new_directory
rmdir empty_directory

# Windows-style
md new_directory
rd empty_directory
```

### Screen Management
- **Unix-style**: `clear`
- **Windows-style**: `cls`

Both commands clear the terminal screen completely.

## Platform-Specific Adaptations

### Shell Script Execution
Gosh provides seamless shell script execution across all platforms:

#### Windows Shell Script Support
On Windows, gosh automatically detects and executes shell scripts that would normally fail with "not a valid Win32 application" errors:

- **Automatic Detection**: Scripts with `.sh` extension or shell shebang lines are automatically recognized
- **Transparent Execution**: Shell scripts are executed through gosh itself, providing full compatibility
- **Shebang Support**: Scripts with `#!/bin/sh`, `#!/bin/bash`, `#!/usr/bin/env gosh`, etc. work seamlessly
- **Argument Passing**: All script arguments are properly passed through

```bash
# These all work on Windows now:
./script.sh arg1 arg2
gosh script.sh arg1 arg2
gosh -c "script.sh arg1 arg2"

# Even scripts without .sh extension work if they have a shebang:
./my_script  # Works if it starts with #!/usr/bin/env gosh
```

#### Cross-Platform Script Examples
```bash
#!/usr/bin/env gosh
# This script works identically on Windows, macOS, and Linux

# Detect platform in a way that works with gosh
if command -v uname >/dev/null 2>&1; then
    local PLATFORM=$(uname -s)
else
    local PLATFORM="Windows"
fi
echo "Running on: $PLATFORM"
echo "Script arguments: $@"

# Use cross-platform commands
ls -la          # Works on all platforms
echo "Files listed above"

# Environment variables work the same way
export MY_VAR="cross-platform value"
echo "MY_VAR: $MY_VAR"
```

### File Paths
Gosh automatically handles different path conventions:
- **Windows**: Supports both `C:\path\to\file` and `C:/path/to/file`
- **Unix/Linux/macOS**: Standard `/path/to/file` format

### File Permissions
- **Unix/Linux/macOS**: Full permission support (rwxrwxrwx format)
- **Windows**: Simplified permission display that maps Windows attributes to Unix-style format

### History File Location
The command history is stored in a platform-appropriate location:
- **Windows**: `%USERPROFILE%\.gosh_history`
- **Unix/Linux/macOS**: `~/.gosh_history`
- **Fallback**: `.gosh_history` in current directory if home directory is unavailable

### File Size Display
Human-readable file sizes adapt to local conventions:
- Supports KB, MB, GB, TB units
- Consistent formatting across all platforms

### Tab Completion Behavior
Gosh adapts its tab completion behavior to match platform conventions:

#### Windows
- **Case-Insensitive**: Tab completion works regardless of case
- **Examples**: `EC[TAB]` completes to `ECho`, `cd proj[TAB]` completes to `cd projects`
- **Behavior**: Preserves your input case and appends the remaining characters

#### Linux/Unix/macOS
- **Case-Sensitive**: Tab completion requires exact case matching
- **Examples**: `EC[TAB]` does not complete to `echo`, `cd proj[TAB]` does not complete to `Projects`
- **Behavior**: Only completes when the case exactly matches available options

This ensures that tab completion feels natural on each platform while maintaining the expected behavior that users are accustomed to.

## Cross-Platform Examples

### Basic File Operations
```bash
# Create a project structure (works on all platforms)
mkdir -p myproject/src/components
mkdir -p myproject/docs

# Create some files
echo "# My Project" > myproject/README.md
echo "console.log('Hello');" > myproject/src/app.js

# List files (choose your preferred style)
ls -la myproject/          # Unix-style
dir -l myproject/          # Windows-style

# Copy files
cp myproject/README.md myproject/docs/
copy myproject/src/app.js myproject/src/backup.js

# View file contents
cat myproject/README.md    # Unix-style
type myproject/README.md   # Windows-style

# Clean up
rm -rf myproject/          # Unix-style recursive removal
```

### Script Compatibility
```bash
#!/usr/bin/env gosh

# This script works identically on all platforms
echo "Cross-platform gosh script"

# Use either command style - both work!
if ls README.md > /dev/null 2>&1; then
    echo "Found README.md using ls"
fi

if dir README.md > /dev/null 2>&1; then
    echo "Found README.md using dir"
fi

# File operations work the same way
mkdir temp_dir
echo "test content" > temp_dir/test.txt
cp temp_dir/test.txt temp_dir/backup.txt
rm -rf temp_dir
```

## Migration from Other Shells

### From Bash/Zsh (Unix/Linux/macOS)
Gosh supports most common bash commands out of the box:
- All standard file operations (`ls`, `cat`, `cp`, `mv`, `rm`, `mkdir`)
- Environment variable handling (`export`, `unset`)
- Command chaining (`;`), pipes (`|`), and redirection (`>`, `>>`, `<`)
- Functions and aliases

### From Command Prompt/PowerShell (Windows)
Windows users can use familiar commands:
- `dir` instead of `ls`
- `type` instead of `cat`
- `copy` instead of `cp`
- `move` instead of `mv`
- `del` instead of `rm`
- `md` instead of `mkdir`
- `cls` instead of `clear`

## Best Practices for Cross-Platform Scripts

1. **Use consistent command style**: Pick either Unix or Windows style and stick with it throughout your script
2. **Handle paths carefully**: Use forward slashes when possible, as they work on all platforms
3. **Test on multiple platforms**: While gosh provides consistency, always test important scripts
4. **Use built-in commands**: Prefer gosh's built-in commands over external utilities for maximum portability
5. **Document platform requirements**: If your script uses external tools, document any platform-specific dependencies

## Limitations

### Pipeline Support
Currently, built-in commands cannot be used in pipelines. This affects both Unix and Windows style commands:
```bash
# This doesn't work yet
cat file.txt | grep "pattern"
type file.txt | findstr "pattern"

# Use this instead
cat file.txt > temp.txt && grep "pattern" temp.txt
```

### External Command Differences
While gosh's built-in commands work consistently, external commands may behave differently:
- `grep` vs `findstr`
- `wc` vs `find /c`
- Different command-line options and syntax

For maximum portability, rely on gosh's built-in commands whenever possible.

## Future Enhancements

Planned improvements for cross-platform support:
- Pipeline support for built-in commands
- Additional Windows-style commands (`attrib`, `xcopy`, etc.)
- Better integration with platform-specific features
- Enhanced tab completion for both command styles with platform-appropriate case sensitivity (case-insensitive on Windows, case-sensitive on Unix/Linux)