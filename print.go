package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

// Print encodes the value using the given encoding and then prints it to the
// standard output. Accepted encodings are "json", "yml", "yaml", "table" and
// "raw". If encoding is the empty string this function defaults to "table"
// encoding.
//
// Usually the encoding is controlled via command line flags of your application
// so the user can select in what format the output should be returned.
//
// Accepted encodings
//
// "table": value is printed via a tab writer (see below)
// "json":  value is printed as indented JSON
// "yaml":  value is printed as YAML
// "raw":   value is printed via fmt.Println
//
// Table encoding
//
// If the "table" encoding is used, the reflection API is used to print all
// exported fields of the value via a tab writer. The columns will be the
// UPPERCASE field names or whatever you set in the "table" tag of the
// corresponding field. Field names with a "table" tag set to "-" are omitted.
// When the "table" encoding is used the value must either be a struct, pointer
// to a struct, a slice or an array.
func Print(encoding string, value interface{}) error {
	return PrintWriter(encoding, value, os.Stdout)
}

// PrintWriter is like Print but lets the caller inject an io.Writer.
func PrintWriter(encoding string, value interface{}, w io.Writer) error {
	switch strings.ToLower(encoding) {
	case "json":
		return printJSON(value, w)
	case "yml", "yaml":
		return printYAML(value, w)
	case "table", "":
		return printTable(value, w)
	case "raw":
		return printRaw(value, w)
	default:
		return fmt.Errorf("unknown encoding %q", encoding)
	}
}

// MustPrint is exactly like Print but panics if an error occurs.
func MustPrint(encoding string, i interface{}) {
	err := Print(encoding, i)
	if err != nil {
		panic(err)
	}
}

func printRaw(i interface{}, w io.Writer) error {
	_, err := fmt.Fprintln(w, i)
	return err
}

func printJSON(i interface{}, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(i)
}

func printYAML(i interface{}, w io.Writer) error {
	out, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(out))
	return err
}

func printTable(v interface{}, w io.Writer) error {
	val := reflect.ValueOf(v)
	t := val.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var isArray bool
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		isArray = true
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		if isArray {
			for i := 0; i < val.Len(); i++ {
				_, err := fmt.Fprintln(w, val.Index(i))
				if err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("cannot print type %T as table (kind %v)", v, t.Kind())
	}

	type field struct {
		Name  string
		Index int
	}

	var fields []field
	var header string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := strings.ToUpper(f.Name)

		if t := f.Tag.Get("table"); t != "" {
			if t == "-" {
				continue
			}
			name = t
		}

		fields = append(fields, field{Name: name, Index: i})
		header += name + "\t"
	}

	records := []map[string]string{}
	if isArray {
		for i := 0; i < val.Len(); i++ {
			rr := map[string]string{}
			for _, f := range fields {
				rr[f.Name] = fmt.Sprint(val.Index(i).Field(f.Index).Interface())
			}
			records = append(records, rr)
		}
	} else {
		rr := map[string]string{}
		for _, f := range fields {
			rr[f.Name] = fmt.Sprint(val.Field(f.Index).Interface())
		}
		records = append(records, rr)
	}

	tw := tabwriter.NewWriter(w, 8, 8, 2, ' ', 0)
	header = strings.TrimSpace(header) + "\n"
	_, err := fmt.Fprint(tw, header)
	if err != nil {
		return err
	}

	for _, record := range records {
		for _, f := range fields {
			_, err = fmt.Fprint(tw, record[f.Name]+"\t")
			if err != nil {
				return err
			}
		}
		fmt.Fprint(tw, "\n")
	}

	return tw.Flush()
}
