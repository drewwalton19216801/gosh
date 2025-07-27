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