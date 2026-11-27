package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
)

func (runtime *Runtime) installPathBuiltins() {
	runtime.Globals.Define("path_separator", string(os.PathSeparator))
	runtime.Globals.Define("path_join", builtin{"path_join", func(arguments []Value) (Value, error) {
		if len(arguments) == 0 {
			return nil, fmt.Errorf("path_join expects at least 1 path")
		}
		parts := make([]string, len(arguments))
		for index, argument := range arguments {
			path, ok := argument.(string)
			if !ok {
				return nil, fmt.Errorf("path_join argument %d must be a string", index+1)
			}
			parts[index] = path
		}
		return filepath.Clean(filepath.Join(parts...)), nil
	}})
	runtime.Globals.Define("path_clean", builtin{"path_clean", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_clean", arguments)
		if err != nil {
			return nil, err
		}
		return filepath.Clean(path), nil
	}})
	runtime.Globals.Define("path_base", builtin{"path_base", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_base", arguments)
		if err != nil {
			return nil, err
		}
		return filepath.Base(path), nil
	}})
	runtime.Globals.Define("path_dir", builtin{"path_dir", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_dir", arguments)
		if err != nil {
			return nil, err
		}
		return filepath.Dir(path), nil
	}})
	runtime.Globals.Define("path_ext", builtin{"path_ext", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_ext", arguments)
		if err != nil {
			return nil, err
		}
		return filepath.Ext(path), nil
	}})
	runtime.Globals.Define("path_is_abs", builtin{"path_is_abs", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_is_abs", arguments)
		if err != nil {
			return nil, err
		}
		return filepath.IsAbs(path), nil
	}})
	runtime.Globals.Define("path_abs", builtin{"path_abs", func(arguments []Value) (Value, error) {
		path, err := singlePathArgument("path_abs", arguments)
		if err != nil {
			return nil, err
		}
		return runtime.pathArgument(path, "path_abs")
	}})
	runtime.Globals.Define("path_rel", builtin{"path_rel", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("path_rel expects base and target paths")
		}
		base, err := runtime.pathArgument(arguments[0], "path_rel")
		if err != nil {
			return nil, err
		}
		target, err := runtime.pathArgument(arguments[1], "path_rel")
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(base, target)
		if err != nil {
			return nil, fmt.Errorf("path_rel: %w", err)
		}
		return relative, nil
	}})
	runtime.Globals.Define("cwd", builtin{"cwd", func(arguments []Value) (Value, error) {
		if len(arguments) != 0 {
			return nil, fmt.Errorf("cwd expects no arguments")
		}
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("cwd: %w", err)
		}
		return workingDirectory, nil
	}})
}

func singlePathArgument(name string, arguments []Value) (string, error) {
	if len(arguments) != 1 {
		return "", fmt.Errorf("%s expects 1 path", name)
	}
	path, ok := arguments[0].(string)
	if !ok {
		return "", fmt.Errorf("%s path must be a string", name)
	}
	return path, nil
}
