/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package winipcfg

// https://docs.microsoft.com/en-us/windows/desktop/api/iptypes/ns-iptypes-_ip_adapter_prefix_xp
// IP_ADAPTER_PREFIX_XP defined in iptypes.h
type wtIpAdapterPrefixXp struct {
	Length       uint32 // Windows type: ULONG
	Flags        uint32 // Windows type: DWORD
	Next         *wtIpAdapterPrefixXp
	Address      wtSocketAddress
	PrefixLength uint32 // Windows type: ULONG
}

func (wt *wtIpAdapterPrefixXp) toIpAdapterPrefix(ifc Interface) (*IpAdapterPrefix, error) {

	if wt == nil {
		return nil, nil
	}

	sainet, err := (&wt.Address).toSockaddrInet()

	if err != nil {
		return nil, err
	}

	ap := IpAdapterPrefix{PrefixLength: wt.PrefixLength}

	ap.InterfaceLuid = ifc.Luid
	ap.InterfaceIndex = ifc.Index
	ap.Length = wt.Length
	ap.Address = *sainet
	ap.Flags = wt.Flags

	return &ap, nil
}
