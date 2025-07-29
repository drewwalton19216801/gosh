package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BuiltinFunc represents a built-in command function
type BuiltinFunc func(*Shell, *Command) error

// builtins maps command names to their implementation functions
var builtins map[string]BuiltinFunc

// init initializes the builtins map
func init() {
	builtins = map[string]BuiltinFunc{
		"exit":    cmdExit,
		"cd":      cmdCd,
		"pwd":     cmdPwd,
		"echo":    cmdEcho,
		"env":     cmdEnv,
		"export":  cmdExport,
		"local":   cmdLocal,
		"unset":   cmdUnset,
		"alias":   cmdAlias,
		"unalias": cmdUnalias,
		"history": cmdHistory,
		"help":    cmdHelp,
		"which":   cmdWhich,
		"case":    cmdCase,
		"declare": cmdDeclare,
		"test":    cmdTest,
		"[":       cmdTest,
		"return":  cmdReturn,
		"fi":      cmdFi,
		"then":    cmdThen,
		"else":    cmdElse,
		"elif":    cmdElif,
		"esac":    cmdEsac,
		"gosh":    cmdGosh,
		"ls":      cmdLs,
		"dir":     cmdDir,
		"cls":     cmdClear,
		"clear":   cmdClear,
		"cat":     cmdCat,
		"type":    cmdCat,
		"cp":      cmdCopy,
		"copy":    cmdCopy,
		"mv":      cmdMove,
		"move":    cmdMove,
		"rm":      cmdRemove,
		"del":     cmdRemove,
		"mkdir":   cmdMkdir,
		"md":      cmdMkdir,
		"rmdir":   cmdRmdir,
		"rd":      cmdRmdir,
	}
}

// cmdExit implements the exit command
func cmdExit(s *Shell, cmd *Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		if c, err := strconv.Atoi(cmd.Args[0]); err == nil {
			code = c
		}
	}
	s.Exit(code)
	return nil
}

// FileInfo represents file information for listing
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

// cmdLs implements the ls command with Unix-style output
func cmdLs(s *Shell, cmd *Command) error {
	return listFiles(cmd.Args, false)
}

// cmdDir implements the dir command with Windows-style output
func cmdDir(s *Shell, cmd *Command) error {
	return listFiles(cmd.Args, true)
}

// listFiles implements file listing functionality
func listFiles(args []string, windowsStyle bool) error {
	var paths []string
	var showAll bool
	var longFormat bool
	var humanReadable bool

	// Parse arguments
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// Handle flags
			for _, flag := range arg[1:] {
				switch flag {
				case 'a':
					showAll = true
				case 'l':
					longFormat = true
				case 'h':
					humanReadable = true
				}
			}
		} else {
			paths = append(paths, arg)
		}
	}

	// Default to current directory if no paths specified
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for i, path := range paths {
		if i > 0 {
			fmt.Println() // Blank line between multiple directories
		}

		if len(paths) > 1 {
			fmt.Printf("%s:\n", path)
		}

		if err := listDirectory(path, showAll, longFormat, humanReadable, windowsStyle); err != nil {
			fmt.Fprintf(os.Stderr, "ls: %v\n", err)
		}
	}

	return nil
}

// listDirectory lists the contents of a single directory
func listDirectory(path string, showAll, longFormat, humanReadable, windowsStyle bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var files []FileInfo

	// Collect file information
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless -a flag is used
		if !showAll && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    name,
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}

	// Sort files alphabetically
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	if windowsStyle {
		return printWindowsStyle(files, longFormat, humanReadable)
	} else {
		return printUnixStyle(files, longFormat, humanReadable)
	}
}

// printUnixStyle prints files in Unix ls style
func printUnixStyle(files []FileInfo, longFormat, humanReadable bool) error {
	if longFormat {
		// Long format: permissions size date name
		for _, file := range files {
			permissions := formatPermissions(file.Mode)
			size := formatSize(file.Size, humanReadable)
			date := file.ModTime.Format("Jan 02 15:04")

			fmt.Printf("%s %8s %s %s\n", permissions, size, date, file.Name)
		}
	} else {
		// Simple format: just names in columns
		const maxCols = 4
		cols := 0
		for _, file := range files {
			name := file.Name
			if file.IsDir {
				name += "/"
			}
			fmt.Printf("%-20s", name)
			cols++
			if cols >= maxCols {
				fmt.Println()
				cols = 0
			}
		}
		if cols > 0 {
			fmt.Println()
		}
	}
	return nil
}

// printWindowsStyle prints files in Windows dir style
func printWindowsStyle(files []FileInfo, longFormat, humanReadable bool) error {
	if longFormat {
		// Windows dir style header
		fmt.Printf(" Directory of %s\n\n", ".")

		var totalSize int64
		fileCount := 0
		dirCount := 0

		for _, file := range files {
			date := file.ModTime.Format("01/02/2006  03:04 PM")

			if file.IsDir {
				fmt.Printf("%s    <DIR>          %s\n", date, file.Name)
				dirCount++
			} else {
				size := formatSize(file.Size, humanReadable)
				fmt.Printf("%s %13s %s\n", date, size, file.Name)
				totalSize += file.Size
				fileCount++
			}
		}

		fmt.Printf("\n%15d File(s) %s bytes\n", fileCount, formatSize(totalSize, false))
		fmt.Printf("%15d Dir(s)\n", dirCount)
	} else {
		// Simple format similar to Unix
		return printUnixStyle(files, false, humanReadable)
	}
	return nil
}

// formatPermissions converts file mode to Unix-style permission string
func formatPermissions(mode os.FileMode) string {
	perms := make([]byte, 10)

	// File type
	if mode.IsDir() {
		perms[0] = 'd'
	} else if mode&os.ModeSymlink != 0 {
		perms[0] = 'l'
	} else {
		perms[0] = '-'
	}

	// Owner permissions
	if mode&0400 != 0 {
		perms[1] = 'r'
	} else {
		perms[1] = '-'
	}
	if mode&0200 != 0 {
		perms[2] = 'w'
	} else {
		perms[2] = '-'
	}
	if mode&0100 != 0 {
		perms[3] = 'x'
	} else {
		perms[3] = '-'
	}

	// Group permissions
	if mode&0040 != 0 {
		perms[4] = 'r'
	} else {
		perms[4] = '-'
	}
	if mode&0020 != 0 {
		perms[5] = 'w'
	} else {
		perms[5] = '-'
	}
	if mode&0010 != 0 {
		perms[6] = 'x'
	} else {
		perms[6] = '-'
	}

	// Other permissions
	if mode&0004 != 0 {
		perms[7] = 'r'
	} else {
		perms[7] = '-'
	}
	if mode&0002 != 0 {
		perms[8] = 'w'
	} else {
		perms[8] = '-'
	}
	if mode&0001 != 0 {
		perms[9] = 'x'
	} else {
		perms[9] = '-'
	}

	return string(perms)
}

// formatSize formats file size with optional human-readable format
func formatSize(size int64, humanReadable bool) string {
	if !humanReadable {
		return fmt.Sprintf("%d", size)
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"K", "M", "G", "T", "P", "E"}
	return fmt.Sprintf("%.1f%s", float64(size)/float64(div), units[exp])
}

// cmdClear implements both cls and clear commands
func cmdClear(s *Shell, cmd *Command) error {
	// ANSI escape sequence to clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
	return nil
}

// cmdCat implements both cat and type commands
func cmdCat(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("cat: missing file operand")
	}

	for _, filename := range cmd.Args {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			continue
		}

		_, err = io.Copy(os.Stdout, file)
		file.Close()

		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: error reading %s: %v\n", filename, err)
		}
	}

	return nil
}

// cmdCopy implements both cp and copy commands
func cmdCopy(s *Shell, cmd *Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("copy: missing destination file operand")
	}

	src := cmd.Args[0]
	dst := cmd.Args[1]

	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("copy: cannot stat '%s': %v", src, err)
	}

	// If destination is a directory, copy into it
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	// Handle directory copying
	if srcInfo.IsDir() {
		return copyDirectory(src, dst)
	}

	// Copy single file
	return copyFile(src, dst)
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// cmdMove implements both mv and move commands
func cmdMove(s *Shell, cmd *Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("move: missing destination file operand")
	}

	src := cmd.Args[0]
	dst := cmd.Args[1]

	// Check if destination is a directory
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	return os.Rename(src, dst)
}

// cmdRemove implements both rm and del commands
func cmdRemove(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("rm: missing file operand")
	}

	var recursive bool
	var force bool
	var files []string

	// Parse arguments
	for _, arg := range cmd.Args {
		if strings.HasPrefix(arg, "-") {
			for _, flag := range arg[1:] {
				switch flag {
				case 'r', 'R':
					recursive = true
				case 'f':
					force = true
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			if !force {
				fmt.Fprintf(os.Stderr, "rm: cannot remove '%s': %v\n", file, err)
			}
			continue
		}

		if info.IsDir() {
			if recursive {
				err = os.RemoveAll(file)
			} else {
				err = fmt.Errorf("is a directory")
			}
		} else {
			err = os.Remove(file)
		}

		if err != nil && !force {
			fmt.Fprintf(os.Stderr, "rm: cannot remove '%s': %v\n", file, err)
		}
	}

	return nil
}

// cmdMkdir implements both mkdir and md commands
func cmdMkdir(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}

	var parents bool
	var dirs []string

	// Parse arguments
	for _, arg := range cmd.Args {
		if strings.HasPrefix(arg, "-") {
			for _, flag := range arg[1:] {
				switch flag {
				case 'p':
					parents = true
				}
			}
		} else {
			dirs = append(dirs, arg)
		}
	}

	for _, dir := range dirs {
		var err error
		if parents {
			err = os.MkdirAll(dir, 0755)
		} else {
			err = os.Mkdir(dir, 0755)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: cannot create directory '%s': %v\n", dir, err)
		}
	}

	return nil
}

// cmdRmdir implements both rmdir and rd commands
func cmdRmdir(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("rmdir: missing operand")
	}

	for _, dir := range cmd.Args {
		if err := os.Remove(dir); err != nil {
			fmt.Fprintf(os.Stderr, "rmdir: failed to remove '%s': %v\n", dir, err)
		}
	}

	return nil
}

// cmdCd implements the cd command
func cmdCd(s *Shell, cmd *Command) error {
	var dir string
	if len(cmd.Args) == 0 {
		// Go to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home directory: %v", err)
		}
		dir = home
	} else {
		dir = cmd.Args[0]
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %v", err)
	}

	return nil
}

// cmdPwd implements the pwd command
func cmdPwd(s *Shell, cmd *Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %v", err)
	}
	fmt.Println(wd)
	return nil
}

// cmdEcho implements the echo command
func cmdEcho(s *Shell, cmd *Command) error {
	// Remove quotes from arguments and process escape sequences
	var processedArgs []string
	for _, arg := range cmd.Args {
		processedArg := arg
		// Remove quotes if present
		if len(arg) >= 2 && ((arg[0] == '"' && arg[len(arg)-1] == '"') || (arg[0] == '\'' && arg[len(arg)-1] == '\'')) {
			processedArg = arg[1 : len(arg)-1]
		}
		// Process escape sequences
		processedArg = processEscapeSequences(processedArg)
		processedArgs = append(processedArgs, processedArg)
	}
	output := strings.Join(processedArgs, " ")
	fmt.Println(output)
	return nil
}

// processEscapeSequences processes common escape sequences in strings
func processEscapeSequences(s string) string {
	result := strings.Builder{}
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"':
				result.WriteByte('"')
				i++ // Skip the next character
			case '\'':
				result.WriteByte('\'')
				i++ // Skip the next character
			case '\\':
				result.WriteByte('\\')
				i++ // Skip the next character
			case 'n':
				result.WriteByte('\n')
				i++ // Skip the next character
			case 't':
				result.WriteByte('\t')
				i++ // Skip the next character
			case 'r':
				result.WriteByte('\r')
				i++ // Skip the next character
			default:
				// Unknown escape sequence, keep the backslash
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

// cmdEnv implements the env command
func cmdEnv(s *Shell, cmd *Command) error {
	// Print system environment
	for _, env := range os.Environ() {
		fmt.Println(env)
	}
	// Print shell-specific environment
	for key, value := range s.env {
		fmt.Printf("%s=%s\n", key, value)
	}
	return nil
}

// cmdExport implements the export command
func cmdExport(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("export: missing variable assignment")
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("export: invalid assignment: %s", arg)
		}
		key, value := parts[0], parts[1]
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		s.env[key] = value
		os.Setenv(key, value)
	}

	return nil
}

// cmdLocal implements the local command
func cmdLocal(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("local: missing variable assignment")
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("local: invalid assignment: %s", arg)
		}
		key, value := parts[0], parts[1]

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Handle command substitution if present
		if strings.Contains(value, "$(") {
			expandedValue, err := s.expandToken(value)
			if err != nil {
				return fmt.Errorf("error expanding variable value: %v", err)
			}
			if len(expandedValue) == 1 {
				value = expandedValue[0]
			}
		}

		// Set the variable in the shell environment only (not system environment)
		s.env[key] = value
	}

	return nil
}

// cmdUnset implements the unset command
func cmdUnset(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unset: missing variable or function name")
	}

	// Check for -f flag to unset functions
	if len(cmd.Args) >= 2 && cmd.Args[0] == "-f" {
		for i := 1; i < len(cmd.Args); i++ {
			delete(s.functions, cmd.Args[i])
		}
		return nil
	}

	// Default behavior: unset variables
	for _, arg := range cmd.Args {
		// Check if it's a function first
		if _, exists := s.functions[arg]; exists {
			delete(s.functions, arg)
		} else {
			// Unset as variable
			delete(s.env, arg)
			os.Unsetenv(arg)
		}
	}

	return nil
}

// cmdAlias implements the alias command
func cmdAlias(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		// List all aliases
		for name, value := range s.aliases {
			fmt.Printf("alias %s='%s'\n", name, value)
		}
		return nil
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("alias: invalid assignment: %s", arg)
		}
		name, value := parts[0], parts[1]
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		s.aliases[name] = value
	}

	return nil
}

// cmdUnalias implements the unalias command
func cmdUnalias(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unalias: missing alias name")
	}

	for _, arg := range cmd.Args {
		delete(s.aliases, arg)
	}

	return nil
}

// cmdHistory implements the history command
func cmdHistory(s *Shell, cmd *Command) error {
	for i, line := range s.history {
		fmt.Printf("%4d  %s\n", i+1, line)
	}
	return nil
}

// cmdHelp implements the help command
func cmdHelp(s *Shell, cmd *Command) error {
	fmt.Println("gosh - A simple shell written in Go")
	fmt.Println("")
	fmt.Println("Built-in commands:")
	fmt.Println("  exit [code]     - Exit the shell with optional exit code")
	fmt.Println("  cd [dir]        - Change directory (defaults to home)")
	fmt.Println("  pwd             - Print working directory")
	fmt.Println("  echo [args...]  - Print arguments")
	fmt.Println("  env             - Show environment variables")
	fmt.Println("  export VAR=val  - Set environment variable")
	fmt.Println("  local VAR=val   - Set local variable (shell only)")
	fmt.Println("  unset VAR       - Unset environment variable or function")
	fmt.Println("  unset -f FUNC   - Unset function")
	fmt.Println("  alias name=cmd  - Create command alias")
	fmt.Println("  unalias name    - Remove alias")
	fmt.Println("  declare         - List all defined functions")
	fmt.Println("  declare -f      - List function names only")
	fmt.Println("  history         - Show command history")
	fmt.Println("  which cmd       - Show path to command")
	fmt.Println("  help            - Show this help")
	fmt.Println("")
	fmt.Println("Features:")
	fmt.Println("  - Tab completion: Press TAB to complete commands and file paths")
	fmt.Println("  - Command chaining: cmd1; cmd2; cmd3")
	fmt.Println("  - Unix pipes: cmd1 | cmd2 | cmd3")
	fmt.Println("  - Input/output redirection: cmd < input.txt > output.txt")
	fmt.Println("  - Background execution: cmd &")
	fmt.Println("  - Script execution: gosh script.sh")
	fmt.Println("  - Case statements: case $var in pattern) commands ;; esac")
	fmt.Println("  - User-defined functions: function_name() { commands; }")
	fmt.Println("  - Function parameters: $1, $2, ..., $#, $0")
	fmt.Println("  - Environment variables and aliases")
	fmt.Println("  - Variable expansion: $VAR, ${VAR}")
	fmt.Println("  - Command substitution: $(cmd), `cmd`")
	fmt.Println("  - Tilde expansion: ~, ~/path, ~user/path")
	fmt.Println("  - Glob patterns: *.txt, file?.log, [abc]*")
	return nil
}

// cmdWhich implements the which command
func cmdWhich(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("which: missing command name")
	}

	for _, cmdName := range cmd.Args {
		// Check if it's a user-defined function
		if _, exists := s.functions[cmdName]; exists {
			fmt.Printf("%s: user-defined function\n", cmdName)
			continue
		}

		// Check if it's a built-in
		if _, exists := builtins[cmdName]; exists {
			fmt.Printf("%s: shell builtin\n", cmdName)
			continue
		}

		// Check if it's an alias
		if alias, exists := s.aliases[cmdName]; exists {
			fmt.Printf("%s: aliased to %s\n", cmdName, alias)
			continue
		}

		// Look for external command
		path, err := s.resolvePath(cmdName)
		if err != nil {
			fmt.Printf("%s: not found\n", cmdName)
		} else {
			// Get absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Println(path)
			} else {
				fmt.Println(absPath)
			}
		}
	}

	return nil
}

// cmdCase implements the case command (for interactive use)
func cmdCase(s *Shell, cmd *Command) error {
	return fmt.Errorf("case statements are only supported in scripts, not interactive mode")
}

// cmdDeclare implements the declare command for listing functions
func cmdDeclare(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		// List all functions
		if len(s.functions) == 0 {
			fmt.Println("No functions defined")
			return nil
		}

		fmt.Println("Defined functions:")
		for name, function := range s.functions {
			fmt.Printf("%s() {\n", name)
			for _, line := range function.Body {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println("}")
			fmt.Println()
		}
		return nil
	}

	// Check for -f flag to list only function names
	if len(cmd.Args) == 1 && cmd.Args[0] == "-f" {
		for name := range s.functions {
			fmt.Println(name)
		}
		return nil
	}

	return fmt.Errorf("declare: unsupported option")
}

// TestFailureError represents a test command failure (exit code 1)
type TestFailureError struct{}

func (e TestFailureError) Error() string {
	return "test failed"
}

// cmdTest implements the test command (and [ command)
func cmdTest(s *Shell, cmd *Command) error {
	args := cmd.Args

	// Handle [ command - remove trailing ] if present
	if cmd.Name == "[" {
		if len(args) == 0 || args[len(args)-1] != "]" {
			return fmt.Errorf("[: missing closing ']'")
		}
		args = args[:len(args)-1]
	}

	// No arguments means false
	if len(args) == 0 {
		return TestFailureError{}
	}

	// Single argument means test if non-empty
	if len(args) == 1 {
		if args[0] != "" {
			return nil // success
		} else {
			return TestFailureError{}
		}
	}

	// Two arguments with unary operator
	if len(args) == 2 {
		operator := args[0]
		operand := args[1]

		switch operator {
		case "-z":
			// String is empty
			if operand == "" {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-n":
			// String is non-empty
			if operand != "" {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-f":
			// File exists and is regular file
			if info, err := os.Stat(operand); err == nil && info.Mode().IsRegular() {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-d":
			// Directory exists
			if info, err := os.Stat(operand); err == nil && info.IsDir() {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-e":
			// File or directory exists
			if _, err := os.Stat(operand); err == nil {
				return nil
			} else {
				return TestFailureError{}
			}
		default:
			return fmt.Errorf("test: unknown unary operator: %s", operator)
		}
	}

	// Three arguments with binary operator
	if len(args) == 3 {
		left := args[0]
		operator := args[1]
		right := args[2]

		switch operator {
		case "=", "==":
			// String equality
			if left == right {
				return nil
			} else {
				return TestFailureError{}
			}
		case "!=":
			// String inequality
			if left != right {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-eq":
			// Numeric equality
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum == rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-ne":
			// Numeric inequality
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum != rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-lt":
			// Numeric less than
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum < rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-le":
			// Numeric less than or equal
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum <= rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-gt":
			// Numeric greater than
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum > rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-ge":
			// Numeric greater than or equal
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum >= rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		default:
			return fmt.Errorf("test: unknown binary operator: %s", operator)
		}
	}

	return fmt.Errorf("test: too many arguments")
}

// ReturnError represents a function return with exit code
type ReturnError struct {
	Code int
}

func (e ReturnError) Error() string {
	return fmt.Sprintf("return %d", e.Code)
}

// cmdReturn implements the return command for functions
func cmdReturn(s *Shell, cmd *Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		if c, err := strconv.Atoi(cmd.Args[0]); err == nil {
			code = c
		}
	}
	return ReturnError{Code: code}
}

// cmdFi implements the fi command (should only be used to close if statements)
func cmdFi(s *Shell, cmd *Command) error {
	return fmt.Errorf("fi: unexpected token (not inside an if statement)")
}

// cmdThen implements the then command (should only be used in if statements)
func cmdThen(s *Shell, cmd *Command) error {
	return fmt.Errorf("then: unexpected token (not inside an if statement)")
}

// cmdElse implements the else command (should only be used in if statements)
func cmdElse(s *Shell, cmd *Command) error {
	return fmt.Errorf("else: unexpected token (not inside an if statement)")
}

// cmdElif implements the elif command (should only be used in if statements)
func cmdElif(s *Shell, cmd *Command) error {
	return fmt.Errorf("elif: unexpected token (not inside an if statement)")
}

// cmdEsac implements the esac command (should only be used to close case statements)
func cmdEsac(s *Shell, cmd *Command) error {
	return fmt.Errorf("esac: unexpected token (not inside a case statement)")
}

// cmdGosh implements a fun easter egg command
func cmdGosh(s *Shell, cmd *Command) error {
	jokes := []string{
		"Gosh! You found the secret command!",
		"Oh my gosh! This shell is shell-arious!",
		"Gosh darn it, you're good at exploring!",
		"Holy gosh! You must be shell-shocked by this discovery!",
		"Gosh golly! This shell really knows how to shell out the fun!",
		"By gosh! You've struck shell gold!",
		"Gosh almighty! You're really shelling it today!",
		"Well I'll be gosh-darned! You found the fun zone!",
		"Gosh! This is what happens when developers get shell-fish with their humor!",
		"Oh gosh! You've entered the shell-ter of bad puns!",
	}

	// Pick a random joke based on the current working directory hash
	// This gives a pseudo-random but deterministic selection
	pwd, _ := os.Getwd()
	hash := 0
	for _, c := range pwd {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}

	selectedJoke := jokes[hash%len(jokes)]
	fmt.Println(selectedJoke)

	// Add some extra flair if they pass arguments
	if len(cmd.Args) > 0 {
		fmt.Printf("Gosh %s, you're really going all out!\n", strings.Join(cmd.Args, " "))
	}

	return nil
}
