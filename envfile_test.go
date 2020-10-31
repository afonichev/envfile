package envfile

import (
	"os"
	"testing"
)

// TestLoadDefaultFile tests the default file loading.
func TestLoadDefaultFile(t *testing.T) {

	// load default file
	err := Load()

	// casting to type PathError
	e, ok := err.(*os.PathError)

	// error is not of this type, not when opening a file or when opening a non-default file
	if !ok || (e.Op != "open") || (e.Path != ".envfile") {
		t.Errorf("didn't try and open file .envfile by default")
	}
}

// TestLoadNotFoundFile tests loading a non-existent file.
func TestLoadNotFoundFile(t *testing.T) {

	// load non-existent file
	if err := Load("not_exist.envfile"); err == nil {
		t.Error("file wasn't found but load didn't return an error")
	}
}

// TestParseFile tests file parsing.
func TestParseFile(t *testing.T) {

	// expected key/value pairs
	pairs := map[string]string{
		"KEY_2": "value",
		"KEY_3": "value",
		"KEY_4": "value of another variable",
		"KEY_5": "{ KEY_1 } of another variable",
		"KEY_6": "Title:\n\t1. value\n\t2. value\n\t3. value",
		"KEY_7": "Title:\\n\\t1. value\\n\\t2. value\\n\\t3. value",
	}

	// parse file
	payloads, err := Parse("test.envfile")
	if err != nil {
		t.Fatalf("error parsing env file: %v", err)
	}

	// iterating over expected key/value pairs
	for key, value := range pairs {

		// iteration over payloads
		for _, payload := range payloads {

			// such a key exists in payloads
			if key == payload.Key {

				// value from payload is different from expected
				if payload.Value != value {
					t.Errorf("expected %s to be %s, got %s", key, value, payload.Value)
				}

				// exit loop
				break
			}
		}
	}
}
