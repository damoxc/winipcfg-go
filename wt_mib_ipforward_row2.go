/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package winipcfg

import (
	"fmt"
	"golang.org/x/sys/windows"
	"net"
	"os"
	"unsafe"
)

// https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/ns-netioapi-_mib_ipforward_row2
// MIB_IPFORWARD_ROW2 defined in netioapi.h
type wtMibIpforwardRow2 struct {
	//
	// Key Structure.
	//
	InterfaceLuid     uint64 // Windows type: NET_LUID
	InterfaceIndex    uint32 // Windows type: NET_IFINDEX
	DestinationPrefix wtIpAddressPrefix
	NextHop           wtSockaddrInet

	//
	// Read-Write Fields.
	//
	SitePrefixLength  uint8  // Windows type: UCHAR
	ValidLifetime     uint32 // Windows type: ULONG
	PreferredLifetime uint32 // Windows type: ULONG
	Metric            uint32 // Windows type: ULONG
	Protocol          NlRouteProtocol

	Loopback             uint8 // Windows type: BOOLEAN
	AutoconfigureAddress uint8 // Windows type: BOOLEAN
	Publish              uint8 // Windows type: BOOLEAN
	Immortal             uint8 // Windows type: BOOLEAN

	//
	// Read-Only Fields.
	//
	Age    uint32 // Windows type: ULONG
	Origin NlRouteOrigin
}

// When interfaceLuid == 0, corresponds to GetIpForwardTable2 function
// (https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/nf-netioapi-getipforwardtable2).
func getWtMibIpforwardRow2s(interfaceLuid uint64, family AddressFamily) ([]wtMibIpforwardRow2, error) {

	var pTable *wtMibIpforwardTable2 = nil

	result := getIpForwardTable2(family, unsafe.Pointer(&pTable))

	if pTable != nil {
		defer freeMibTable(unsafe.Pointer(pTable))
	}

	if result != 0 {
		return nil, os.NewSyscallError("iphlpapi.GetIpForwardTable2", windows.Errno(result))
	}

	var rows []wtMibIpforwardRow2;

	if interfaceLuid == 0 {
		rows = make([]wtMibIpforwardRow2, pTable.NumEntries, pTable.NumEntries)
	}

	pFirstRow := uintptr(unsafe.Pointer(&pTable.Table[0]))
	rowSize := uintptr(wtMibIpforwardRow2_Size) // Should be equal to unsafe.Sizeof(pTable.Table[0])

	for i := uint32(0); i < pTable.NumEntries; i++ {

		row := (*wtMibIpforwardRow2)(unsafe.Pointer(pFirstRow + rowSize*uintptr(i)))

		if interfaceLuid == 0 {
			rows[i] = *row
		} else if row.InterfaceLuid == interfaceLuid {
			rows = append(rows, *row)
		}
	}

	return rows, nil
}

func findWtMibIpforwardRow2(interfaceLuid uint64, destination *net.IPNet) (*wtMibIpforwardRow2, error) {

	rows, err := getWtMibIpforwardRow2s(interfaceLuid, AF_UNSPEC)

	if err != nil {
		return nil, err
	}

	ones, _ := destination.Mask.Size()

	for _, row := range rows {
		if row.DestinationPrefix.PrefixLength == uint8(ones) && row.DestinationPrefix.Prefix.matches(&destination.IP) {
			return &row, nil
		}
	}

	return nil, nil
}

// Uses InitializeIpForwardEntry function
// (https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/nf-netioapi-initializeipforwardentry).
func getInitializedWtMibIpforwardRow2(interfaceLuid uint64) *wtMibIpforwardRow2 {

	row := wtMibIpforwardRow2{InterfaceLuid: interfaceLuid}

	_ = initializeIpForwardEntry(&row)

	row.InterfaceLuid = interfaceLuid

	return &row
}

func createAndAddWtMibIpforwardRow2(interfaceLuid uint64, routeData *RouteData) error {

	wtdest, err := createWtIpAddressPrefix(&routeData.Destination)

	if err != nil {
		return err
	}

	wtsaNextHop, err := createWtSockaddrInet(&routeData.NextHop, 0)

	if err != nil {
		return err
	}

	row := getInitializedWtMibIpforwardRow2(interfaceLuid)

	row.DestinationPrefix = *wtdest
	row.NextHop = *wtsaNextHop
	row.Metric = routeData.Metric

	return row.add()
}

// Uses CreateIpForwardEntry2 function
// (https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/nf-netioapi-createipforwardentry2).
func (r *wtMibIpforwardRow2) add() error {

	result := createIpForwardEntry2(r)

	if result == 0 {
		return nil
	} else {
		return os.NewSyscallError("iphlpapi.CreateIpForwardEntry2", windows.Errno(result))
	}
}

// Corresponds to DeleteIpForwardEntry2 function
// (https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/nf-netioapi-deleteipforwardentry2).
func (r *wtMibIpforwardRow2) delete() error {

	result := deleteIpForwardEntry2(r)

	if result == 0 {
		return nil
	} else {
		return os.NewSyscallError("iphlpapi.DeleteIpForwardEntry2", windows.Errno(result))
	}
}

func (r *wtMibIpforwardRow2) toRoute() (*Route, error) {

	if r == nil {
		return nil, nil
	}

	iap, err := r.DestinationPrefix.toIpAddressPrefix()

	if err != nil {
		return nil, err
	}

	sainet, err := r.NextHop.toSockaddrInet()

	if err != nil {
		return nil, err
	}

	return &Route{
		InterfaceLuid:        r.InterfaceLuid,
		InterfaceIndex:       r.InterfaceIndex,
		DestinationPrefix:    *iap,
		NextHop:              *sainet,
		SitePrefixLength:     r.SitePrefixLength,
		ValidLifetime:        r.ValidLifetime,
		PreferredLifetime:    r.PreferredLifetime,
		Metric:               r.Metric,
		Protocol:             r.Protocol,
		Loopback:             uint8ToBool(r.Loopback),
		AutoconfigureAddress: uint8ToBool(r.AutoconfigureAddress),
		Publish:              uint8ToBool(r.Publish),
		Immortal:             uint8ToBool(r.Immortal),
		Age:                  r.Age,
		Origin:               r.Origin,
	}, nil
}

func (r *wtMibIpforwardRow2) extractRouteData() (*RouteData, error) {

	if r == nil {
		return nil, nil
	}

	iap, err := r.DestinationPrefix.toIpAddressPrefix()

	if err != nil {
		return nil, err
	}

	destination, err := iap.toNetIpNet()

	if err != nil {
		return nil, err
	}

	sainet, err := r.NextHop.toSockaddrInet()

	if err != nil {
		return nil, err
	}

	return &RouteData{
		Destination: *destination,
		NextHop:     sainet.Address,
		Metric:      r.Metric,
	}, nil
}

func (r *wtMibIpforwardRow2) String() string {

	if r == nil {
		return "<nil>"
	}

	return fmt.Sprintf(`InterfaceLuid: %d
InterfaceIndex: %d
DestinationPrefix: %s
NextHop: %s
SitePrefixLength: %d
ValidLifetime: %d
PreferredLifetime: %d
Metric: %d
Protocol: %s
Loopback: %d
AutoconfigureAddress: %d
Publish: %d
Immortal: %d
Age: %d
Origin: %s
`, r.InterfaceLuid, r.InterfaceIndex, r.DestinationPrefix.String(), r.NextHop.String(), r.SitePrefixLength,
		r.ValidLifetime, r.PreferredLifetime, r.Metric, r.Protocol.String(), r.Loopback, r.AutoconfigureAddress,
		r.Publish, r.Immortal, r.Age, r.Origin.String())
}
