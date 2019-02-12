/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package winipcfg

import "fmt"

// https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/ns-netioapi-_ip_address_prefix
type IP_ADDRESS_PREFIX struct {
	Prefix wtSockaddrInet
	PrefixLength uint8 // Windows type: UINT8
}

func (addrPfx *IP_ADDRESS_PREFIX) String() string {
	prefix := addrPfx.Prefix.String()
	return fmt.Sprintf("Prefix: [%s]; PrefixLength: %d", prefix, addrPfx.PrefixLength)
}
