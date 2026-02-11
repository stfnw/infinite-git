// SPDX-FileCopyrightText: 2026 Stefan Walter (stfnw)
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
)

func pktline(data []byte) []byte {
	prefix := fmt.Appendf(nil, "%04x", len(data)+4)
	return append(prefix, data...)
}

func sideband(channel SidebandChannel, msg []byte) []byte {
	data := append([]byte{byte(channel)}, msg...)
	return pktline(data)
}

func flush() []byte {
	return []byte("0000")
}

// func delim() []byte {
// 	return []byte("0001")
// }

// func responseEnd() []byte {
// 	return []byte("0002")
// }

type SidebandChannel byte

const (
	SidebandPackfile SidebandChannel = 1
	SidebandProgress SidebandChannel = 2
	SidebandError    SidebandChannel = 3
)
