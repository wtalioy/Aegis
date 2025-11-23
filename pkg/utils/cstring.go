package utils

import "bytes"

// extract a null-terminated C string from a byte array
func ExtractCString(data []byte) string {
	if idx := bytes.IndexByte(data, 0); idx != -1 {
		return string(data[:idx])
	}
	return string(data)
}
