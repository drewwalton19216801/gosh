# Tab Completion Test Guide

This guide demonstrates the tab completion functionality in gosh.

## Testing Command Completion

1. Start the shell: `./gosh`
2. Type `ec` and press TAB - should complete to `echo`
3. Type `ex` and press TAB - should complete to `exit` and `export`
4. Type `h` and press TAB - should show `help` and `history`
5. Type `qu` and press TAB - should complete to `quota` (not `ququota`)

## Testing File/Directory Completion

1. Type `cd ` (with space) and press TAB - should show directories in current folder
2. Type `ls *.` and press TAB - should show files starting with dot
3. Type `cat README` and press TAB - should complete to `README.md`
4. Navigate to `examples/` directory and type `./case` then press TAB - should complete to `./case_statement_demo.sh`

## Testing Tilde Expansion Completion

1. Type `ls ~/` and press TAB - should show files and directories in your home directory
2. Type `cd ~/D` and press TAB - should complete to directories starting with 'D' in your home
3. Type `ls ~/Documents/` and press TAB - should show contents of ~/Documents/ directory
4. Type `echo ~` and press TAB - should complete to `echo ~/`
5. Type `cd ~/` and press TAB - should not expand further (no completions shown)
6. Type `cat ~/.*` and press TAB - should show hidden files in home directory

## Testing Glob Pattern Completion

1. Create test files: `touch file1. file2. test.`
2. Type `ls .` and press TAB - should show files ending with dots
3. Create test files: `touch test.txt example.txt sample.go`
4. Type `ls *.txt` and press TAB - should show .txt files
5. Type `ls test.*` and press TAB - should show files starting with "test"
6. Test with directories: `mkdir dir1 dir2` then `ls dir*` + TAB

## Testing PATH Command Completion

1. Type `l` and press TAB - should show commands like `ls`, `ln`, etc.
2. Type `g` and press TAB - should show commands like `git`, `grep`, etc.

## Testing Case-Insensitive Completion

1. Type `EC` and press TAB - should complete to `ECho`
2. Type `EX` and press TAB - should complete to `EXit`
3. Type `exi` and press TAB - should complete to `exit`
4. Type `EXI` and press TAB - should complete to `EXIt` (NOT `EXIexit`)
5. Type `cd ~/PROJ` and press TAB - should complete to `~/PROJects` (if Projects exists)
6. Type `LS` and press TAB - should complete to `LSls`
7. Type `GIT` and press TAB - should show git commands with preserved case

## Testing Windows Path Completion

**Windows Backslash Support:**
1. Type `cd Projects\` and press TAB - should show subdirectories using backslashes
2. Type `cd Projects\g` and press TAB - should complete to `Projects\go\` (preserving backslash style)
3. Type `ls Documents\` and press TAB - should show files in Documents using backslashes
4. Type `cat README.md` then `copy README.md backup\` and press TAB - should complete backup directory path

**Mixed Path Separator Handling:**
1. Type `cd Projects/` and press TAB - should show subdirectories using forward slashes
2. Type `cd Projects\g` and press TAB - should complete using backslashes (preserving user's style)
3. Type `ls ./` and press TAB - should show current directory contents with forward slashes
4. Type `ls .\` and press TAB - should show current directory contents with backslashes

**Cross-Platform Compatibility:**
- Both `cd Projects/go` and `cd Projects\go` work on Windows
- Tab completion preserves your preferred path separator style
- No mixed separators in completion results (e.g., no `Projects/gProjects\go/`)
- Windows paths ending with backslashes (e.g., `cd Projects\`) work correctly without triggering line continuation mode

### Case-Insensitive Completion and Execution Behavior

The shell supports case-insensitive completion and execution for built-in commands, aliases, and user-defined functions.

**Tab Completion Behavior:**
When using case-insensitive completion, the shell preserves your input case and appends the remaining part from the filesystem or command name:

**File/Directory Completions:**
- Input: `proj` → Completion: `projects` (your case + filesystem remainder)
- Input: `PROJ` → Completion: `PROJects` (your case + filesystem remainder)
- Input: `Projects` → Completion: `Projects` (exact match uses filesystem case)

**Command Completions:**
- Input: `exi` → Completion: `exit` (your case + command remainder)
- Input: `EXI` → Completion: `EXIt` (your case + command remainder)
- Input: `EC` → Completion: `ECho` (your case + command remainder)
- Input: `exit` → Completion: `exit` (exact match uses actual command case)

**Command Execution:**
Commands can be executed regardless of case:
- `exit`, `EXIT`, `Exit`, `EXit` all work
- `echo`, `ECHO`, `Echo`, `ECho` all work
- `pwd`, `PWD`, `Pwd`, `PWd` all work
- Aliases and user-defined functions also support case-insensitive execution

This behavior ensures that:
- Tab completion works regardless of case
- Commands execute regardless of case
- Your typing style is respected
- The completion clearly shows it found a case-insensitive match
- No duplication bugs occur (like `PROJProjects` or `EXIexit`)

## Features Implemented

- ✅ Built-in command completion (exit, cd, pwd, echo, etc.)
- ✅ External command completion from PATH
- ✅ File and directory completion
- ✅ File path completion (including ./ and ../ prefixes)
- ✅ Tilde expansion completion (~ and ~/path)
- ✅ Alias completion
- ✅ Context-aware completion (commands vs arguments)
- ✅ Hidden file completion (when prefix starts with dot)
- ✅ Directory trailing slash addition
- ✅ Duplicate removal and sorting
- ✅ Case-insensitive completion for commands, aliases, and files
- ✅ Cross-platform path separator support (both / and \ on Windows)
- ✅ Path separator style preservation (maintains user's preferred style)
- ✅ Windows backslash path completion
- ✅ Mixed path separator handling without corruption