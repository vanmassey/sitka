package interpreter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func (runtime *Runtime) installFilesystemBuiltins() {
	runtime.Globals.Define("read_text", builtin{"read_text", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("read_text expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "read_text")
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read_text %q: %w", path, err)
		}
		return string(content), nil
	}})
	runtime.Globals.Define("write_text", builtin{"write_text", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("write_text expects a path and string")
		}
		path, err := runtime.pathArgument(arguments[0], "write_text")
		if err != nil {
			return nil, err
		}
		content, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("write_text content must be a string")
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("write_text %q: %w", path, err)
		}
		return int64(len(content)), nil
	}})
	runtime.Globals.Define("append_text", builtin{"append_text", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("append_text expects a path and string")
		}
		path, err := runtime.pathArgument(arguments[0], "append_text")
		if err != nil {
			return nil, err
		}
		content, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("append_text content must be a string")
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("append_text %q: %w", path, err)
		}
		written, writeErr := file.WriteString(content)
		closeErr := file.Close()
		if writeErr != nil {
			return nil, fmt.Errorf("append_text %q: %w", path, writeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("append_text %q: %w", path, closeErr)
		}
		return int64(written), nil
	}})
	runtime.Globals.Define("file_exists", builtin{"file_exists", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("file_exists expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "file_exists")
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(path)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return nil, fmt.Errorf("file_exists %q: %w", path, err)
	}})
	runtime.Globals.Define("make_dir", builtin{"make_dir", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("make_dir expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "make_dir")
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, fmt.Errorf("make_dir %q: %w", path, err)
		}
		return path, nil
	}})
	runtime.Globals.Define("list_dir", builtin{"list_dir", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("list_dir expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "list_dir")
		if err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("list_dir %q: %w", path, err)
		}
		if len(entries) > maxGeneratedItems {
			return nil, fmt.Errorf("directory contains more than %d entries", maxGeneratedItems)
		}
		names := make([]Value, len(entries))
		for index, entry := range entries {
			names[index] = entry.Name()
		}
		return names, nil
	}})
	runtime.Globals.Define("read_lines", builtin{"read_lines", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("read_lines expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "read_lines")
		if err != nil {
			return nil, err
		}
		content, err := readLimitedFile(path, maxGeneratedStringBytes)
		if err != nil {
			return nil, fmt.Errorf("read_lines %q: %w", path, err)
		}
		if !utf8.Valid(content) {
			return nil, fmt.Errorf("read_lines %q: file is not valid UTF-8 text", path)
		}
		lines, err := splitTextLines(string(content))
		if err != nil {
			return nil, fmt.Errorf("read_lines %q: %w", path, err)
		}
		return lines, nil
	}})
	runtime.Globals.Define("write_lines", builtin{"write_lines", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("write_lines expects a path and list of strings")
		}
		path, err := runtime.pathArgument(arguments[0], "write_lines")
		if err != nil {
			return nil, err
		}
		lines, ok := arguments[1].([]Value)
		if !ok {
			return nil, fmt.Errorf("write_lines lines must be a list")
		}
		var output strings.Builder
		for index, value := range lines {
			line, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("write_lines item %d must be a string", index+1)
			}
			if strings.ContainsAny(line, "\r\n") {
				return nil, fmt.Errorf("write_lines item %d cannot contain a newline", index+1)
			}
			separatorBytes := 0
			if index > 0 {
				separatorBytes = 1
			}
			if index == len(lines)-1 {
				separatorBytes++
			}
			if len(line)+separatorBytes > maxGeneratedStringBytes-output.Len() {
				return nil, fmt.Errorf("write_lines content exceeds maximum size of %d bytes", maxGeneratedStringBytes)
			}
			if index > 0 {
				output.WriteByte('\n')
			}
			output.WriteString(line)
		}
		if len(lines) > 0 {
			output.WriteByte('\n')
		}
		content := output.String()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("write_lines %q: %w", path, err)
		}
		return int64(len(lines)), nil
	}})
	runtime.Globals.Define("read_bytes", builtin{"read_bytes", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("read_bytes expects 1 path")
		}
		path, err := runtime.pathArgument(arguments[0], "read_bytes")
		if err != nil {
			return nil, err
		}
		content, err := readLimitedFile(path, maxGeneratedItems)
		if err != nil {
			return nil, fmt.Errorf("read_bytes %q: %w", path, err)
		}
		result := make([]Value, len(content))
		for index, value := range content {
			result[index] = int64(value)
		}
		return result, nil
	}})
	runtime.Globals.Define("write_bytes", builtin{"write_bytes", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("write_bytes expects a path and list of bytes")
		}
		path, err := runtime.pathArgument(arguments[0], "write_bytes")
		if err != nil {
			return nil, err
		}
		content, err := bytesArgument(arguments[1], "write_bytes")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return nil, fmt.Errorf("write_bytes %q: %w", path, err)
		}
		return int64(len(content)), nil
	}})
	runtime.Globals.Define("append_bytes", builtin{"append_bytes", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("append_bytes expects a path and list of bytes")
		}
		path, err := runtime.pathArgument(arguments[0], "append_bytes")
		if err != nil {
			return nil, err
		}
		content, err := bytesArgument(arguments[1], "append_bytes")
		if err != nil {
			return nil, err
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("append_bytes %q: %w", path, err)
		}
		written, writeErr := file.Write(content)
		closeErr := file.Close()
		if writeErr != nil {
			return nil, fmt.Errorf("append_bytes %q: %w", path, writeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("append_bytes %q: %w", path, closeErr)
		}
		return int64(written), nil
	}})
}

func splitTextLines(text string) ([]Value, error) {
	if len(text) > maxGeneratedStringBytes {
		return nil, fmt.Errorf("text exceeds maximum size of %d bytes", maxGeneratedStringBytes)
	}
	breaks := 0
	for index := 0; index < len(text); index++ {
		if text[index] == '\n' {
			breaks++
		} else if text[index] == '\r' && (index+1 == len(text) || text[index+1] != '\n') {
			breaks++
		}
		if breaks > maxGeneratedItems {
			return nil, fmt.Errorf("text exceeds maximum line count of %d", maxGeneratedItems)
		}
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if text == "" {
		return []Value{}, nil
	}
	text = strings.TrimSuffix(text, "\n")
	parts := strings.Split(text, "\n")
	lines := make([]Value, len(parts))
	for index, line := range parts {
		lines[index] = line
	}
	return lines, nil
}

func bytesArgument(value Value, builtinName string) ([]byte, error) {
	values, ok := value.([]Value)
	if !ok {
		return nil, fmt.Errorf("%s bytes must be a list", builtinName)
	}
	if len(values) > maxGeneratedItems {
		return nil, fmt.Errorf("%s input exceeds maximum size of %d bytes", builtinName, maxGeneratedItems)
	}
	content := make([]byte, len(values))
	for index, value := range values {
		number, ok := value.(int64)
		if !ok || number < 0 || number > 255 {
			return nil, fmt.Errorf("%s item %d must be an integer from 0 to 255", builtinName, index+1)
		}
		content[index] = byte(number)
	}
	return content, nil
}

func readLimitedFile(path string, maximumBytes int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	content, readErr := io.ReadAll(io.LimitReader(file, int64(maximumBytes)+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(content) > maximumBytes {
		return nil, fmt.Errorf("file exceeds maximum size of %d bytes", maximumBytes)
	}
	return content, nil
}

func (runtime *Runtime) pathArgument(value Value, builtinName string) (string, error) {
	path, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s path must be a string", builtinName)
	}
	if !filepath.IsAbs(path) {
		base := "."
		if runtime.currentPath != "" {
			base = filepath.Dir(runtime.currentPath)
		}
		path = filepath.Join(base, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("%s path %q: %w", builtinName, path, err)
	}
	return absolute, nil
}
