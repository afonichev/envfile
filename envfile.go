package envfile

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Payload structure.
type Payload struct {

	// line number in file
	Line int

	// export status
	Export bool

	// overload status
	Overload bool

	// key
	Key string

	// value
	Value string
}

var (

	// key name validation
	validation = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

	// unescape special characters
	unescape = regexp.MustCompile(`\\.`)
)

// Load will load files with environment variables for this process.
func Load(filenames ...string) error {

	// file name list is empty
	if len(filenames) == 0 {

		// add the default filename to the list
		filenames = append(filenames, ".envfile")
	}

	// iterating over a list of filenames
	for _, filename := range filenames {

		// parse file
		payloads, err := Parse(filename)
		if err != nil {
			return err
		}

		// iteration over payloads
		for _, payload := range payloads {

			// key is exported or overloaded
			if payload.Export || payload.Overload {

				// key does not exist in environment variables or is overloaded
				if value, ok := os.LookupEnv(payload.Key); !ok || payload.Overload {

					// ignore overload on the same value
					if payload.Value == value {
						continue
					}

					// set key and value to environment variable
					if err := os.Setenv(payload.Key, payload.Value); err != nil {
						return fmt.Errorf("[%s] %s", filename, err)
					}
				}
			}
		}
	}

	return nil
}

// Parse parses file with environment variables.
func Parse(filename string) ([]Payload, error) {

	// open file with environment variables
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// deferred file close
	defer file.Close()

	// line number
	var line int

	// payload list
	var payloads []Payload

	// line by line file reading
	scanner := bufio.NewScanner(file)

	// iterate through the lines of the file
	for scanner.Scan() {

		// increase line number
		line++

		// current line
		current := strings.TrimSpace(scanner.Text())

		// ignore blank lines and comments
		if len(current) == 0 || strings.HasPrefix(current, "#") {
			continue
		}

		// split current line with equal sign
		pair := strings.SplitN(current, "=", 2)

		// could not split current line
		if len(pair) != 2 {
			return nil, fmt.Errorf("[%s] line %d: can't split line into key and value", filename, line)
		}

		// payload
		var payload Payload

		// set line number
		payload.Line = line

		// set key name
		payload.Key = strings.TrimSpace(pair[0])

		// export directive
		if strings.HasPrefix(strings.ToLower(payload.Key), "export") {

			// update key name
			payload.Key = strings.TrimSpace(payload.Key[6:])

			// set export status
			payload.Export = true
		}

		// overload directive
		if strings.HasPrefix(strings.ToLower(payload.Key), "overload") {

			// update key name
			payload.Key = strings.TrimSpace(payload.Key[8:])

			// set overload status
			payload.Overload = true
		}

		// empty key name
		if len(payload.Key) == 0 {
			return nil, fmt.Errorf("[%s] line %d: key name is empty", filename, line)
		}

		// invalid key name
		if !validation.MatchString(payload.Key) {
			return nil, fmt.Errorf("[%s] line %d: invalid key name '%s'", filename, line, payload.Key)
		}

		// iterating over a list of payloads
		for _, pld := range payloads {

			// key already exists in the payload list
			if pld.Key == payload.Key {
				return nil, fmt.Errorf("[%s] line %d: duplicate key '%s'", filename, line, payload.Key)
			}
		}

		// set value
		payload.Value = strings.TrimSpace(pair[1])

		// add payload to list
		payloads = append(payloads, payload)
	}

	// cycle of changing variables to their values
	for {

		// temporary storage of variable names with their positions
		temp := make(map[string][][2]int)

		// iterating over a list of payloads
		for _, payload := range payloads {

			// previous character
			var previous rune

			// parts list
			var parts []string

			// character list
			var chars []rune

			// iteration over value
			for _, current := range payload.Value {

				// start of variable
				if previous != '{' && current == '{' {

					// list of characters is not empty
					if len(chars) > 0 {

						// combine characters and add to parts list
						parts = append(parts, string(chars))

						// clear the list of characters
						chars = nil
					}
				}

				// end of variable
				if previous == '}' && current != '}' {

					// list of characters is not empty
					if len(chars) > 0 {

						// combine characters and add to parts list
						parts = append(parts, string(chars))

						// clear the list of characters
						chars = nil
					}
				}

				// add the current character to the character list
				chars = append(chars, current)

				// change the previous character to the current one
				previous = current
			}

			// list of characters is not empty
			if len(chars) > 0 {

				// combine characters and add to parts list
				parts = append(parts, string(chars))
			}

			// offset to get the position of the variable in the original value
			var offset int

			// iteration by parts
			for i, part := range parts {

				// number of opening curly braces
				opening := strings.Count(part, "{")

				// number of closing curly braces
				closing := strings.Count(part, "}")

				// an even number of opening curly braces status
				evenOpening := (opening % 2) == 0

				// an even number of closing curly braces status
				evenClosing := (closing % 2) == 0

				// only open curly braces were found and their number is odd
				if !evenOpening && evenClosing {

					// there are more opening curly braces than closing curly braces
					if opening > closing {

						// current part is the last and is equal to the opening curly brace
						if (i == len(parts)-1) && (part == "{") {
							return nil, fmt.Errorf("[%s] line %d: excess opening curly brace '{' in at the end",
								filename, payload.Line)
						}

						return nil, fmt.Errorf("[%s] line %d: can't find the closing curly brace '}'",
							filename, payload.Line)
					}

					// there are fewer opening curly braces than closing curly braces
					if opening < closing {
						return nil, fmt.Errorf("[%s] line %d: excess closing curly brace '}'", filename, payload.Line)
					}
				}

				// only close curly braces were found and their number is odd
				if evenOpening && !evenClosing {

					// there are more opening curly braces than closing curly braces
					if opening > closing {
						return nil, fmt.Errorf("[%s] line %d: excess opening curly brace '{'", filename, payload.Line)
					}

					// there are fewer opening curly braces than closing curly braces
					if opening < closing {

						// current part is the first and is equal to the closing curly brace
						if (i == 0) && (part == "}") {
							return nil, fmt.Errorf("[%s] line %d: excess closing curly brace '}' at the beginning",
								filename, payload.Line)
						}

						return nil, fmt.Errorf("[%s] line %d: can't find the opening curly brace '{'",
							filename, payload.Line)
					}
				}

				// found open and close curly braces and their number is odd
				if !evenOpening && !evenClosing {

					// start of variable
					start := offset + opening

					// end of variable
					end := offset + (len(part) - closing)

					// variable
					variable := strings.TrimSpace(payload.Value[start:end])

					// empty variable name
					if len(variable) == 0 {
						return nil, fmt.Errorf("[%s] line %d: variable name is empty", filename, payload.Line)
					}

					// variable name is the same as the name of the current key
					if payload.Key == variable {
						return nil, fmt.Errorf("[%s] line %d: key '%s' is used recursively",
							filename, payload.Line, payload.Key)
					}

					// add a variable and its position to temporary storage
					temp[payload.Key] = append(temp[payload.Key], [2]int{

						// start of variable
						start,

						// end of variable
						end,
					})
				}

				// increase offset by part length
				offset += len(part)
			}
		}

		// temporary storage is not empty
		if len(temp) > 0 {

			// iterating over temporary storage
			for variable, positions := range temp {

				// iterating over a list of payloads
				for i, payload := range payloads {

					// key exists in the list of payloads
					if payload.Key == variable {

						// current line number
						line := payload.Line

						// iterating over variable positions
						for i := len(positions) - 1; i >= 0; i-- {

							// variable position
							position := positions[i]

							// start of variable
							start := position[0]

							// end of variable
							end := position[1]

							// variable
							variable := strings.TrimSpace(payload.Value[start:end])

							// character list
							var chars []rune

							// add everything before the variable to the character list
							chars = append(chars, []rune(payload.Value[:start-1])...)

							// variable value
							var value *string

							// iterating over a list of payloads
							for _, payload := range payloads {

								// variable exists in the list of payloads
								if payload.Key == variable {

									// update variable value
									value = &payload.Value

									// exit loop
									break
								}
							}

							// variable value is missing
							if value == nil {

								// variable value from environment variables
								value, ok := os.LookupEnv(variable)

								// variable does not exist
								if !ok {
									return nil, fmt.Errorf("[%s] line %d: variable '%s' does not exist",
										filename, line, variable)
								}

								// add variable value to character list
								chars = append(chars, []rune(value)...)

							} else {

								// add variable value to character list
								chars = append(chars, []rune(*value)...)
							}

							// add everything after the variable to the character list
							chars = append(chars, []rune(payload.Value[end+1:])...)

							// combine characters and update value
							payload.Value = string(chars)
						}
					}

					// update payload
					payloads[i] = payload
				}
			}

		} else {

			// exit loop
			break
		}
	}

	// iterating over a list of payloads
	for i, payload := range payloads {

		// unescape the open curly brace
		payload.Value = strings.ReplaceAll(payload.Value, "{{", "{")

		// unescape the close curly brace
		payload.Value = strings.ReplaceAll(payload.Value, "}}", "}")

		// unescape the special characters
		payload.Value = unescape.ReplaceAllStringFunc(payload.Value, func(match string) string {

			switch strings.TrimPrefix(match, "\\") {

			// new line
			case "n":
				return "\n"

			// horizontal tab
			case "t":
				return "\t"

			// backslash
			case "\\":
				return "\\"

			// any
			default:
				return match
			}
		})

		// update payload
		payloads[i] = payload
	}

	return payloads, nil
}
