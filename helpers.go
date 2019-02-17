/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package winipcfg

import (
	"fmt"
	"golang.org/x/sys/windows"
	"strings"
	"unsafe"
)

const (
	expectedStringLength = 1000
)

func wcharToString(wchar *uint16) string {

	buffer := make([]uint16, expectedStringLength)

	const size = uintptr(2) // unsafe.Sizeof(uint16(0))

	start := unsafe.Pointer(wchar)

	for i := 0; ; i++ {

		letter := *(*uint16)(unsafe.Pointer(uintptr(start) + size*uintptr(i)))

		if letter == 0 {

			if i < expectedStringLength {
				buffer[i] = 0
			}

			break
		}

		if i < expectedStringLength {
			buffer[i] = letter
		} else {
			buffer = append(buffer, letter)
		}
	}

	return windows.UTF16ToString(buffer)
}

func charToString(char *uint8) string {

	buffer := make([]byte, expectedStringLength)

	const size = uintptr(1) // unsafe.Sizeof(uint8(0))

	start := unsafe.Pointer(char)

	var i int

	for i = 0; ; i++ {

		letter := *(*uint8)(unsafe.Pointer(uintptr(start) + size*uintptr(i)))

		if letter == 0 {
			break
		}

		if i < expectedStringLength {
			buffer[i] = byte(letter)
		} else {
			buffer = append(buffer, byte(letter))
		}
	}

	if i < len(buffer) {
		buffer = buffer[:i]
	}

	return string(buffer)
}

func guidToString(guid windows.GUID) string {
	return fmt.Sprintf("{%06X-%04X-%04X-%04X-%012X}", guid.Data1, guid.Data2, guid.Data3, guid.Data4[:2],
		guid.Data4[2:])
}

func toIndentedText(text, indent string) string {

	indented := strings.TrimSpace(text)

	indented = strings.Replace(indented, "\n", fmt.Sprintf("\n%s", indent), -1)

	return indent + indented
}

func uint8ToBool(val uint8) bool {
	return val != 0
}

func boolToUint8(val bool) uint8 {
	if val {
		return 1
	} else {
		return 0
	}
}

func allZeroBytes(bytes []byte) bool {

	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}

	return true
}