package interpreter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
)

func (runtime *Runtime) installCSVBuiltins() {
	runtime.Globals.Define("csv_decode", builtin{"csv_decode", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("csv_decode expects 1 string")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("csv_decode argument must be a string")
		}
		records, err := decodeCSV(text)
		if err != nil {
			return nil, err
		}
		result := make([]Value, len(records))
		for rowIndex, record := range records {
			row := make([]Value, len(record))
			for columnIndex, field := range record {
				row[columnIndex] = field
			}
			result[rowIndex] = row
		}
		return result, nil
	}})
	runtime.Globals.Define("csv_encode", builtin{"csv_encode", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("csv_encode expects 1 list")
		}
		rows, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("csv_encode argument must be a list of rows")
		}
		records, err := csvRecords(rows)
		if err != nil {
			return nil, err
		}
		return encodeCSV(records)
	}})
	runtime.Globals.Define("csv_decode_maps", builtin{"csv_decode_maps", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("csv_decode_maps expects 1 string")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("csv_decode_maps argument must be a string")
		}
		records, err := decodeCSV(text)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return []Value{}, nil
		}
		headers := records[0]
		if err := validateCSVHeaders(headers); err != nil {
			return nil, err
		}
		result := make([]Value, 0, len(records)-1)
		for rowIndex, record := range records[1:] {
			if len(record) != len(headers) {
				return nil, fmt.Errorf("CSV row %d has %d fields; header has %d", rowIndex+2, len(record), len(headers))
			}
			row := make(map[string]Value, len(headers))
			for columnIndex, header := range headers {
				row[header] = record[columnIndex]
			}
			result = append(result, row)
		}
		return result, nil
	}})
	runtime.Globals.Define("csv_encode_maps", builtin{"csv_encode_maps", func(arguments []Value) (Value, error) {
		if len(arguments) < 1 || len(arguments) > 2 {
			return nil, fmt.Errorf("csv_encode_maps expects records and optional headers")
		}
		rows, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("csv_encode_maps first argument must be a list of maps")
		}
		headers, err := csvMapHeaders(rows)
		if err != nil {
			return nil, err
		}
		if len(arguments) == 2 {
			headers, err = csvHeadersArgument(arguments[1])
			if err != nil {
				return nil, err
			}
		}
		if err := validateCSVHeaders(headers); err != nil {
			return nil, err
		}
		records := make([][]string, 0, len(rows)+1)
		records = append(records, headers)
		for rowIndex, value := range rows {
			row, ok := value.(map[string]Value)
			if !ok {
				return nil, fmt.Errorf("CSV row %d must be a map", rowIndex+1)
			}
			record := make([]string, len(headers))
			for columnIndex, header := range headers {
				field, found := row[header]
				if found {
					record[columnIndex], err = csvFieldText(field)
					if err != nil {
						return nil, fmt.Errorf("CSV row %d field %q: %w", rowIndex+1, header, err)
					}
				}
			}
			for key := range row {
				if !containsString(headers, key) {
					return nil, fmt.Errorf("CSV row %d contains undeclared field %q", rowIndex+1, key)
				}
			}
			records = append(records, record)
		}
		return encodeCSV(records)
	}})
}

func decodeCSV(text string) ([][]string, error) {
	if len(text) > maxGeneratedStringBytes {
		return nil, fmt.Errorf("CSV input exceeds maximum size of %d bytes", maxGeneratedStringBytes)
	}
	reader := csv.NewReader(bytes.NewBufferString(text))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}
	if len(records) > maxGeneratedItems {
		return nil, fmt.Errorf("CSV has more than %d rows", maxGeneratedItems)
	}
	totalFields := 0
	for _, record := range records {
		totalFields += len(record)
		if totalFields > maxGeneratedItems {
			return nil, fmt.Errorf("CSV has more than %d fields", maxGeneratedItems)
		}
	}
	return records, nil
}

func encodeCSV(records [][]string) (string, error) {
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	for rowIndex, record := range records {
		if len(record) == 1 && record[0] == "" {
			writer.Flush()
			if err := writer.Error(); err != nil {
				return "", fmt.Errorf("encode CSV row %d: %w", rowIndex+1, err)
			}
			output.WriteString("\"\"\n")
			continue
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("encode CSV row %d: %w", rowIndex+1, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("encode CSV: %w", err)
	}
	if output.Len() > maxGeneratedStringBytes {
		return "", fmt.Errorf("CSV output exceeds maximum size of %d bytes", maxGeneratedStringBytes)
	}
	return output.String(), nil
}

func csvRecords(rows []Value) ([][]string, error) {
	if len(rows) > maxGeneratedItems {
		return nil, fmt.Errorf("CSV has more than %d rows", maxGeneratedItems)
	}
	records := make([][]string, len(rows))
	fields := 0
	for rowIndex, value := range rows {
		row, ok := value.([]Value)
		if !ok {
			return nil, fmt.Errorf("CSV row %d must be a list", rowIndex+1)
		}
		fields += len(row)
		if fields > maxGeneratedItems {
			return nil, fmt.Errorf("CSV has more than %d fields", maxGeneratedItems)
		}
		record := make([]string, len(row))
		for columnIndex, field := range row {
			text, err := csvFieldText(field)
			if err != nil {
				return nil, fmt.Errorf("CSV row %d column %d: %w", rowIndex+1, columnIndex+1, err)
			}
			record[columnIndex] = text
		}
		records[rowIndex] = record
	}
	return records, nil
}

func csvFieldText(value Value) (string, error) {
	switch value.(type) {
	case nil:
		return "", nil
	case string, bool, int64, float64:
		return Format(value), nil
	default:
		return "", fmt.Errorf("cannot encode %s as a CSV field", TypeName(value))
	}
}

func csvHeadersArgument(value Value) ([]string, error) {
	values, ok := value.([]Value)
	if !ok {
		return nil, fmt.Errorf("CSV headers must be a list of strings")
	}
	headers := make([]string, len(values))
	for index, value := range values {
		header, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("CSV header %d must be a string", index+1)
		}
		headers[index] = header
	}
	return headers, nil
}

func csvMapHeaders(rows []Value) ([]string, error) {
	seen := make(map[string]bool)
	headers := make([]string, 0)
	for rowIndex, value := range rows {
		row, ok := value.(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("CSV row %d must be a map", rowIndex+1)
		}
		for key := range row {
			if !seen[key] {
				seen[key] = true
				headers = append(headers, key)
			}
		}
	}
	sort.Strings(headers)
	return headers, nil
}

func validateCSVHeaders(headers []string) error {
	seen := make(map[string]bool, len(headers))
	for _, header := range headers {
		if seen[header] {
			return fmt.Errorf("duplicate CSV header %q", header)
		}
		seen[header] = true
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
