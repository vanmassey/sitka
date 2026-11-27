package interpreter

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"unicode/utf8"
)

func (runtime *Runtime) installEncodingBuiltins() {
	runtime.Globals.Define("base64_encode", builtin{"base64_encode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("base64_encode", arguments)
		if err != nil {
			return nil, err
		}
		if base64.StdEncoding.EncodedLen(len(text)) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("base64_encode result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		return base64.StdEncoding.EncodeToString([]byte(text)), nil
	}})
	runtime.Globals.Define("base64_decode", builtin{"base64_decode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("base64_decode", arguments)
		if err != nil {
			return nil, err
		}
		if len(text) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("base64_decode input exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		decoded, err := base64.StdEncoding.Strict().DecodeString(text)
		if err != nil {
			return nil, fmt.Errorf("base64_decode: %w", err)
		}
		if !utf8.Valid(decoded) {
			return nil, fmt.Errorf("base64_decode result is not valid UTF-8 text")
		}
		return string(decoded), nil
	}})
	runtime.Globals.Define("hex_encode", builtin{"hex_encode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("hex_encode", arguments)
		if err != nil {
			return nil, err
		}
		if len(text) > maxGeneratedStringBytes/2 {
			return nil, fmt.Errorf("hex_encode result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		return hex.EncodeToString([]byte(text)), nil
	}})
	runtime.Globals.Define("hex_decode", builtin{"hex_decode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("hex_decode", arguments)
		if err != nil {
			return nil, err
		}
		if len(text) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("hex_decode input exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		decoded, err := hex.DecodeString(text)
		if err != nil {
			return nil, fmt.Errorf("hex_decode: %w", err)
		}
		if !utf8.Valid(decoded) {
			return nil, fmt.Errorf("hex_decode result is not valid UTF-8 text")
		}
		return string(decoded), nil
	}})
	runtime.Globals.Define("url_encode", builtin{"url_encode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("url_encode", arguments)
		if err != nil {
			return nil, err
		}
		encoded := url.QueryEscape(text)
		if len(encoded) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("url_encode result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		return encoded, nil
	}})
	runtime.Globals.Define("url_decode", builtin{"url_decode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("url_decode", arguments)
		if err != nil {
			return nil, err
		}
		if len(text) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("url_decode input exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		decoded, err := url.QueryUnescape(text)
		if err != nil {
			return nil, fmt.Errorf("url_decode: %w", err)
		}
		return decoded, nil
	}})
	runtime.Globals.Define("query_encode", builtin{"query_encode", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("query_encode expects 1 map")
		}
		values, ok := arguments[0].(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("query_encode argument must be a map")
		}
		query := make(url.Values, len(values))
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			switch value := values[key].(type) {
			case string:
				query.Add(key, value)
			case []Value:
				for index, item := range value {
					text, ok := item.(string)
					if !ok {
						return nil, fmt.Errorf("query_encode value %q item %d must be a string", key, index+1)
					}
					query.Add(key, text)
				}
			default:
				return nil, fmt.Errorf("query_encode value %q must be a string or list of strings", key)
			}
		}
		return query.Encode(), nil
	}})
	runtime.Globals.Define("query_decode", builtin{"query_decode", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("query_decode", arguments)
		if err != nil {
			return nil, err
		}
		query, err := url.ParseQuery(text)
		if err != nil {
			return nil, fmt.Errorf("query_decode: %w", err)
		}
		if len(query) > maxGeneratedItems {
			return nil, fmt.Errorf("query_decode exceeds maximum result size of %d keys", maxGeneratedItems)
		}
		result := make(map[string]Value, len(query))
		total := 0
		for key, values := range query {
			total += len(values)
			if total > maxGeneratedItems {
				return nil, fmt.Errorf("query_decode exceeds maximum result size of %d values", maxGeneratedItems)
			}
			if len(values) == 1 {
				result[key] = values[0]
				continue
			}
			items := make([]Value, len(values))
			for index, value := range values {
				items[index] = value
			}
			result[key] = items
		}
		return result, nil
	}})
	runtime.Globals.Define("url_parse", builtin{"url_parse", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("url_parse", arguments)
		if err != nil {
			return nil, err
		}
		parsed, err := url.Parse(text)
		if err != nil {
			return nil, fmt.Errorf("url_parse: %w", err)
		}
		result := map[string]Value{
			"scheme":    parsed.Scheme,
			"opaque":    parsed.Opaque,
			"host":      parsed.Host,
			"hostname":  parsed.Hostname(),
			"port":      parsed.Port(),
			"path":      parsed.Path,
			"raw_query": parsed.RawQuery,
			"fragment":  parsed.Fragment,
		}
		if parsed.User != nil {
			result["user"] = parsed.User.Username()
			if password, ok := parsed.User.Password(); ok {
				result["password"] = password
			}
		}
		return result, nil
	}})
	runtime.Globals.Define("url_build", builtin{"url_build", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("url_build expects 1 map")
		}
		fields, ok := arguments[0].(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("url_build argument must be a map")
		}
		allowed := map[string]bool{
			"scheme": true, "opaque": true, "host": true, "path": true,
			"raw_query": true, "fragment": true, "user": true, "password": true,
		}
		parts := make(map[string]string, len(fields))
		for name, value := range fields {
			if !allowed[name] {
				return nil, fmt.Errorf("url_build does not support field %q", name)
			}
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("url_build field %q must be a string", name)
			}
			parts[name] = text
		}
		parsed := url.URL{
			Scheme:   parts["scheme"],
			Opaque:   parts["opaque"],
			Host:     parts["host"],
			Path:     parts["path"],
			RawQuery: parts["raw_query"],
			Fragment: parts["fragment"],
		}
		if password, found := parts["password"]; found {
			user, hasUser := parts["user"]
			if !hasUser {
				return nil, fmt.Errorf("url_build password requires a user field")
			}
			parsed.User = url.UserPassword(user, password)
		} else if user, found := parts["user"]; found {
			parsed.User = url.User(user)
		}
		return parsed.String(), nil
	}})
	runtime.Globals.Define("sha256", builtin{"sha256", func(arguments []Value) (Value, error) {
		text, err := singleStringArgument("sha256", arguments)
		if err != nil {
			return nil, err
		}
		digest := sha256.Sum256([]byte(text))
		return hex.EncodeToString(digest[:]), nil
	}})
	runtime.Globals.Define("sha256_bytes", builtin{"sha256_bytes", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("sha256_bytes expects 1 list of bytes")
		}
		content, err := bytesArgument(arguments[0], "sha256_bytes")
		if err != nil {
			return nil, err
		}
		digest := sha256.Sum256(content)
		return hex.EncodeToString(digest[:]), nil
	}})
}

func singleStringArgument(name string, arguments []Value) (string, error) {
	if len(arguments) != 1 {
		return "", fmt.Errorf("%s expects 1 string", name)
	}
	text, ok := arguments[0].(string)
	if !ok {
		return "", fmt.Errorf("%s argument must be a string", name)
	}
	return text, nil
}
