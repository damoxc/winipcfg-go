/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package winipcfg

import (
	"bytes"
	"fmt"
	"golang.zx2c4.com/winipcfg/internal/ole"
	"net"
	"strings"
)

type GatewayCost struct {
	Gateway    net.IP
	CostMetric uint16
}

func (gw *GatewayCost) String() string {
	if gw == nil {
		return "<nil>"
	} else {
		return fmt.Sprintf("%s (Metric: %d)", gw.Gateway.String(), gw.CostMetric)
	}
}

// https://docs.microsoft.com/en-us/windows/desktop/CIMWin32Prov/win32-networkadapterconfiguration
// Based on WMI Win32_NetworkAdapterConfiguration class.
type NetworkAdapterConfiguration struct {
	Caption                      string
	Description                  string
	SettingID                    string
	ArpAlwaysSourceRoute         bool
	ArpUseEtherSNAP              bool
	DatabasePath                 string
	DeadGWDetectEnabled          bool
	DefaultIPGateway             []GatewayCost
	DefaultTOS                   uint8
	DefaultTTL                   uint8
	DHCPEnabled                  bool
	DHCPLeaseExpires             string //time.Time
	DHCPLeaseObtained            string //time.Time
	DHCPServer                   string
	DNSDomain                    string
	DNSDomainSuffixSearchOrder   []string
	DNSEnabledForWINSResolution  bool
	DNSHostName                  string
	DNSServerSearchOrder         []net.IP
	DomainDNSRegistrationEnabled bool
	ForwardBufferMemory          uint32
	FullDNSRegistrationEnabled   bool
	IGMPLevel                    uint8
	Index                        uint32
	InterfaceIndex               uint32
	IPAddress                    []net.IPNet
	IPConnectionMetric           uint32
	IPEnabled                    bool
	IPFilterSecurityEnabled      bool
	IPPortSecurityEnabled        bool
	IPSecPermitIPProtocols       []string
	IPSecPermitTCPPorts          []string
	IPSecPermitUDPPorts          []string
	IPUseZeroBroadcast           bool
	IPXAddress                   string
	IPXEnabled                   bool
	IPXFrameType                 []uint32
	IPXMediaType                 uint32
	IPXNetworkNumber             []string
	IPXVirtualNetNumber          string
	KeepAliveInterval            uint32
	KeepAliveTime                uint32
	MACAddress                   string
	MTU                          uint32
	NumForwardPackets            uint32
	PMTUBHDetectEnabled          bool
	PMTUDiscoveryEnabled         bool
	ServiceName                  string
	TcpipNetbiosOptions          uint32
	TcpMaxConnectRetransmissions uint32
	TcpMaxDataRetransmissions    uint32
	TcpNumConnections            uint32
	TcpUseRFC1122UrgentPointer   bool
	TcpWindowSize                uint16
	WINSEnableLMHostsLookup      bool
	WINSHostLookupFile           string
	WINSPrimaryServer            string
	WINSScopeID                  string
	WINSSecondaryServer          string
}

func getOlePropertyValueArray(item *ole.IDispatch, propertyName string) ([]interface{}, error) {

	arrVal, err := ole.GetProperty(item, propertyName)

	if err != nil {
		return nil, err
	}

	if arrVal == nil {
		return nil, nil
	}

	arr := arrVal.ToArray()

	if arr == nil {
		return nil, nil
	} else {
		return arr.ToValueArray(), nil
	}
}

func getOlePropertyValueStringArray(item *ole.IDispatch, propertyName string) ([]string, error) {

	arr, err := getOlePropertyValueArray(item, propertyName)

	if err != nil {
		return nil, err
	}

	if arr == nil {
		return nil, nil
	}

	length := len(arr)

	strs := make([]string, length, length)

	for idx, val := range arr {
		strs[idx] = val.(string)
	}

	return strs, nil
}

func getOlePropertyValueUint16Array(item *ole.IDispatch, propertyName string) ([]uint16, error) {

	arr, err := getOlePropertyValueArray(item, propertyName)

	if err != nil {
		return nil, err
	}

	if arr == nil {
		return nil, nil
	}

	length := len(arr)

	strs := make([]uint16, length, length)

	for idx, val := range arr {
		strs[idx] = uint16(val.(int32))
	}

	return strs, nil
}

func getOlePropertyValueUint32Array(item *ole.IDispatch, propertyName string) ([]uint32, error) {

	arr, err := getOlePropertyValueArray(item, propertyName)

	if err != nil {
		return nil, err
	}

	if arr == nil {
		return nil, nil
	}

	length := len(arr)

	strs := make([]uint32, length, length)

	for idx, val := range arr {
		strs[idx] = uint32(val.(int32))
	}

	return strs, nil
}

func itemRawToNetworkAdaptersConfigurations(itemRaw *ole.VARIANT, settingId string) (*NetworkAdapterConfiguration, error) {

	item := itemRaw.ToIDispatch()
	defer item.Release()

	val, err := ole.GetProperty(item, "SettingID")

	if err != nil {
		return nil, err
	}

	sid := val.ToString()

	if settingId != "" && settingId != strings.ToUpper(strings.TrimSpace(sid)) {
		return nil, nil
	}

	nac := NetworkAdapterConfiguration{SettingID: sid}

	val, err = ole.GetProperty(item, "Caption")

	if err != nil {
		return nil, err
	}

	nac.Caption = val.ToString()

	val, err = ole.GetProperty(item, "Description")

	if err != nil {
		return nil, err
	}

	nac.Description = val.ToString()

	val, err = ole.GetProperty(item, "ArpAlwaysSourceRoute")

	if err != nil {
		return nil, err
	}

	nac.ArpAlwaysSourceRoute = val.Val != 0

	val, err = ole.GetProperty(item, "ArpUseEtherSNAP")

	if err != nil {
		return nil, err
	}

	nac.ArpUseEtherSNAP = val.Val != 0

	val, err = ole.GetProperty(item, "DatabasePath")

	if err != nil {
		return nil, err
	}

	nac.DatabasePath = val.ToString()

	val, err = ole.GetProperty(item, "DeadGWDetectEnabled")

	if err != nil {
		return nil, err
	}

	nac.DeadGWDetectEnabled = val.Val != 0

	stringArr, err := getOlePropertyValueStringArray(item, "DefaultIPGateway")

	if err != nil {
		return nil, err
	}

	uint16Arr, err := getOlePropertyValueUint16Array(item, "GatewayCostMetric")

	if err != nil {
		return nil, err
	}

	if stringArr == nil {
		// TODO: Remove the following check:
		if uint16Arr != nil && len(uint16Arr) != 0 {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - DefaultIPGateway property is nil while GatewayCostMetric contains %d items",
				len(uint16Arr))
		}

		nac.DefaultIPGateway = nil
	} else {

		length := len(stringArr)

		// TODO: Remove the following check:
		if uint16Arr == nil {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - DefaultIPGateway property contains %d items while GatewayCostMetric is nil",
				length)
		}

		// TODO: Remove the following check:
		if uint16Arr == nil || len(uint16Arr) != length {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - DefaultIPGateway property contains %d items while GatewayCostMetric contains %d",
				length, len(uint16Arr))
		}

		nac.DefaultIPGateway = make([]GatewayCost, length, length)

		for idx, addr := range stringArr {
			nac.DefaultIPGateway[idx] = GatewayCost{
				Gateway:    net.ParseIP(addr),
				CostMetric: uint16Arr[idx],
			}
		}
	}

	val, err = ole.GetProperty(item, "DefaultTOS")

	if err != nil {
		return nil, err
	}

	nac.DefaultTOS = uint8(val.Val)

	val, err = ole.GetProperty(item, "DefaultTTL")

	if err != nil {
		return nil, err
	}

	nac.DefaultTTL = uint8(val.Val)

	val, err = ole.GetProperty(item, "DHCPEnabled")

	if err != nil {
		return nil, err
	}

	nac.DHCPEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "DHCPLeaseExpires")

	if err != nil {
		return nil, err
	}

	nac.DHCPLeaseExpires = val.ToString()

	val, err = ole.GetProperty(item, "DHCPLeaseObtained")

	if err != nil {
		return nil, err
	}

	nac.DHCPLeaseObtained = val.ToString()

	val, err = ole.GetProperty(item, "DHCPServer")

	if err != nil {
		return nil, err
	}

	nac.DHCPServer = val.ToString()

	val, err = ole.GetProperty(item, "DNSDomain")

	if err != nil {
		return nil, err
	}

	nac.DNSDomain = val.ToString()

	stringArr, err = getOlePropertyValueStringArray(item, "DNSDomainSuffixSearchOrder")

	if err != nil {
		return nil, err
	}

	nac.DNSDomainSuffixSearchOrder = stringArr

	val, err = ole.GetProperty(item, "DNSEnabledForWINSResolution")

	if err != nil {
		return nil, err
	}

	nac.DNSEnabledForWINSResolution = val.Val != 0

	val, err = ole.GetProperty(item, "DNSHostName")

	if err != nil {
		return nil, err
	}

	nac.DNSHostName = val.ToString()

	stringArr, err = getOlePropertyValueStringArray(item, "DNSServerSearchOrder")

	if err != nil {
		return nil, err
	}

	if stringArr == nil {
		nac.DNSServerSearchOrder = nil
	} else {

		length := len(stringArr)

		nac.DNSServerSearchOrder = make([]net.IP, length, length)

		for idx, addr := range stringArr {
			nac.DNSServerSearchOrder[idx] = net.ParseIP(addr)
		}
	}

	val, err = ole.GetProperty(item, "DomainDNSRegistrationEnabled")

	if err != nil {
		return nil, err
	}

	nac.DomainDNSRegistrationEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "ForwardBufferMemory")

	if err != nil {
		return nil, err
	}

	nac.ForwardBufferMemory = uint32(val.Val)

	val, err = ole.GetProperty(item, "FullDNSRegistrationEnabled")

	if err != nil {
		return nil, err
	}

	nac.FullDNSRegistrationEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "IGMPLevel")

	if err != nil {
		return nil, err
	}

	nac.IGMPLevel = uint8(val.Val)

	val, err = ole.GetProperty(item, "Index")

	if err != nil {
		return nil, err
	}

	nac.Index = uint32(val.Val)

	val, err = ole.GetProperty(item, "InterfaceIndex")

	if err != nil {
		return nil, err
	}

	nac.InterfaceIndex = uint32(val.Val)

	stringArr, err = getOlePropertyValueStringArray(item, "IPAddress")

	if err != nil {
		return nil, err
	}

	subnetArr, err := getOlePropertyValueStringArray(item, "IPSubnet")

	if err != nil {
		return nil, err
	}

	if stringArr == nil {
		// TODO: Remove the following check:
		if subnetArr != nil && len(subnetArr) > 0 {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - IPAddress property is nil while IPSubnet property contains %d items",
				len(subnetArr))
		}

		nac.IPAddress = nil
	} else {

		length := len(stringArr)

		// TODO: Remove the following check:
		if subnetArr == nil {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - IPAddress property contains %d items while IPSubnet is nil",
				length)
		}

		// TODO: Remove the following check:
		if len(subnetArr) != length {
			return nil, fmt.Errorf(
				"itemRawToNetworkAdaptersConfigurations() - IPAddress property contains %d items while IPSubnet contains %d",
				length, len(subnetArr))
		}

		nac.IPAddress = make([]net.IPNet, length, length)

		for idx, addr := range stringArr {
			nac.IPAddress[idx] = net.IPNet{
				IP:   net.ParseIP(addr),
				Mask: net.IPMask(net.ParseIP(subnetArr[idx])),
			}
		}
	}

	val, err = ole.GetProperty(item, "IPConnectionMetric")

	if err != nil {
		return nil, err
	}

	nac.IPConnectionMetric = uint32(val.Val)

	val, err = ole.GetProperty(item, "IPEnabled")

	if err != nil {
		return nil, err
	}

	nac.IPEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "IPFilterSecurityEnabled")

	if err != nil {
		return nil, err
	}

	nac.IPFilterSecurityEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "IPPortSecurityEnabled")

	if err != nil {
		return nil, err
	}

	nac.IPPortSecurityEnabled = val.Val != 0

	stringArr, err = getOlePropertyValueStringArray(item, "IPSecPermitIPProtocols")

	if err != nil {
		return nil, err
	}

	nac.IPSecPermitIPProtocols = stringArr

	stringArr, err = getOlePropertyValueStringArray(item, "IPSecPermitTCPPorts")

	if err != nil {
		return nil, err
	}

	nac.IPSecPermitTCPPorts = stringArr

	stringArr, err = getOlePropertyValueStringArray(item, "IPSecPermitUDPPorts")

	if err != nil {
		return nil, err
	}

	nac.IPSecPermitUDPPorts = stringArr

	val, err = ole.GetProperty(item, "IPUseZeroBroadcast")

	if err != nil {
		return nil, err
	}

	nac.IPUseZeroBroadcast = val.Val != 0

	val, err = ole.GetProperty(item, "IPXAddress")

	if err != nil {
		return nil, err
	}

	nac.IPXAddress = val.ToString()

	val, err = ole.GetProperty(item, "IPXEnabled")

	if err != nil {
		return nil, err
	}

	nac.IPXEnabled = val.Val != 0

	uint32Arr, err := getOlePropertyValueUint32Array(item, "IPXFrameType")

	if err != nil {
		return nil, err
	}

	nac.IPXFrameType = uint32Arr

	val, err = ole.GetProperty(item, "IPXMediaType")

	if err != nil {
		return nil, err
	}

	nac.IPXMediaType = uint32(val.Val)

	stringArr, err = getOlePropertyValueStringArray(item, "IPXNetworkNumber")

	if err != nil {
		return nil, err
	}

	nac.IPXNetworkNumber = stringArr

	val, err = ole.GetProperty(item, "IPXVirtualNetNumber")

	if err != nil {
		return nil, err
	}

	nac.IPXVirtualNetNumber = val.ToString()

	val, err = ole.GetProperty(item, "KeepAliveInterval")

	if err != nil {
		return nil, err
	}

	nac.KeepAliveInterval = uint32(val.Val)

	val, err = ole.GetProperty(item, "KeepAliveTime")

	if err != nil {
		return nil, err
	}

	nac.KeepAliveTime = uint32(val.Val)

	val, err = ole.GetProperty(item, "MACAddress")

	if err != nil {
		return nil, err
	}

	nac.MACAddress = val.ToString()

	val, err = ole.GetProperty(item, "MTU")

	if err != nil {
		return nil, err
	}

	nac.MTU = uint32(val.Val)

	val, err = ole.GetProperty(item, "NumForwardPackets")

	if err != nil {
		return nil, err
	}

	nac.NumForwardPackets = uint32(val.Val)

	val, err = ole.GetProperty(item, "PMTUBHDetectEnabled")

	if err != nil {
		return nil, err
	}

	nac.PMTUBHDetectEnabled = val.Val != 0

	val, err = ole.GetProperty(item, "ServiceName")

	if err != nil {
		return nil, err
	}

	nac.ServiceName = val.ToString()

	val, err = ole.GetProperty(item, "TcpipNetbiosOptions")

	if err != nil {
		return nil, err
	}

	nac.TcpipNetbiosOptions = uint32(val.Val)

	val, err = ole.GetProperty(item, "TcpMaxConnectRetransmissions")

	if err != nil {
		return nil, err
	}

	nac.TcpMaxConnectRetransmissions = uint32(val.Val)

	val, err = ole.GetProperty(item, "TcpMaxDataRetransmissions")

	if err != nil {
		return nil, err
	}

	nac.TcpMaxDataRetransmissions = uint32(val.Val)

	val, err = ole.GetProperty(item, "TcpNumConnections")

	if err != nil {
		return nil, err
	}

	nac.TcpNumConnections = uint32(val.Val)

	val, err = ole.GetProperty(item, "TcpUseRFC1122UrgentPointer")

	if err != nil {
		return nil, err
	}

	nac.TcpUseRFC1122UrgentPointer = val.Val != 0

	val, err = ole.GetProperty(item, "TcpWindowSize")

	if err != nil {
		return nil, err
	}

	nac.TcpWindowSize = uint16(val.Val)

	val, err = ole.GetProperty(item, "WINSEnableLMHostsLookup")

	if err != nil {
		return nil, err
	}

	nac.WINSEnableLMHostsLookup = val.Val != 0

	val, err = ole.GetProperty(item, "WINSHostLookupFile")

	if err != nil {
		return nil, err
	}

	nac.WINSHostLookupFile = val.ToString()

	val, err = ole.GetProperty(item, "WINSPrimaryServer")

	if err != nil {
		return nil, err
	}

	nac.WINSPrimaryServer = val.ToString()

	val, err = ole.GetProperty(item, "WINSScopeID")

	if err != nil {
		return nil, err
	}

	nac.WINSScopeID = val.ToString()

	val, err = ole.GetProperty(item, "WINSSecondaryServer")

	if err != nil {
		return nil, err
	}

	nac.WINSSecondaryServer = val.ToString()

	return &nac, nil
}

func getNetworkAdaptersConfigurations(adapterName string) (interface{}, error) {

	// init COM, oh yeah
	err := ole.CoInitialize(0)

	if err != nil {
		return nil, err
	}

	defer ole.CoUninitialize()

	unknown, err := ole.CreateObject("WbemScripting.SWbemLocator")

	if err != nil {
		return nil, err
	}

	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return nil, err
	}

	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := ole.CallMethod(wmi, "ConnectServer")

	if err != nil {
		return nil, err
	}

	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, err := ole.CallMethod(service, "ExecQuery",
		"SELECT * FROM Win32_NetworkAdapterConfiguration")

	if err != nil {
		return nil, err
	}

	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, err := ole.GetProperty(result, "Count")

	if err != nil {
		return nil, err
	}

	count := int(countVar.Val)

	var nacs []*NetworkAdapterConfiguration

	if adapterName == "" {
		nacs = make([]*NetworkAdapterConfiguration, count, count)
	} else {
		adapterName = strings.ToUpper(strings.TrimSpace(adapterName))
	}

	for i := 0; i < count; i++ {
		// item is a SWbemObject, but really a Win32_NetworkAdapterConfiguration
		itemRaw, err := ole.CallMethod(result, "ItemIndex", i)

		if err != nil {
			return nil, err
		}

		nac, err := itemRawToNetworkAdaptersConfigurations(itemRaw, adapterName)

		if err != nil {
			return nil, err
		}

		if nac != nil {
			if adapterName == "" {
				nacs[i] = nac
			} else {
				return nac, nil
			}
		}
	}

	if adapterName == "" {
		return nacs, nil
	} else {
		return nil, fmt.Errorf("getNetworkAdaptersConfigurations() - interface not found")
	}
}

func setDnsesIfMatch(itemRaw *ole.VARIANT, settingId string, dnses []net.IP, add bool) (bool, error) {

	item := itemRaw.ToIDispatch()
	defer item.Release()

	val, err := ole.GetProperty(item, "SettingID")

	if err != nil {
		return false, err
	}

	if settingId != strings.ToUpper(strings.TrimSpace(val.ToString())) {
		return false, nil
	}

	length := 0

	if dnses != nil {
		length = len(dnses)
	}

	strDnses := make([]string, length, length)

	for i := 0; i < length; i++ {
		strDnses[i] = dnses[i].String()
	}

	if add {

		oldDnses, err := getOlePropertyValueStringArray(item, "DNSServerSearchOrder")

		if err != nil {
			return false, err
		}

		if oldDnses != nil && len(oldDnses) > 0 {
			strDnses = append(oldDnses, strDnses...)
		}
	}

	result, err := ole.CallMethod(item, "SetDNSServerSearchOrder", strDnses)

	if err != nil {
		return false, err
	}

	if result.Val == 0 {
		return true, nil
	} else {
		return false, fmt.Errorf("SetDNSServerSearchOrder returned %d", result.Val)
	}
}

func setDnses(ifc *Interface, dnses []net.IP, add bool) error {

	// init COM, oh yeah
	err := ole.CoInitialize(0)

	if err != nil {
		return err
	}

	defer ole.CoUninitialize()

	unknown, err := ole.CreateObject("WbemScripting.SWbemLocator")

	if err != nil {
		return err
	}

	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return err
	}

	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := ole.CallMethod(wmi, "ConnectServer")

	if err != nil {
		return err
	}

	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, err := ole.CallMethod(service, "ExecQuery",
		"SELECT * FROM Win32_NetworkAdapterConfiguration")

	if err != nil {
		return err
	}

	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, err := ole.GetProperty(result, "Count")

	if err != nil {
		return err
	}

	adapterName := strings.ToUpper(strings.TrimSpace(ifc.AdapterName))

	count := int(countVar.Val)

	for i := 0; i < count; i++ {
		// item is a SWbemObject, but really a Win32_NetworkAdapterConfiguration
		itemRaw, err := ole.CallMethod(result, "ItemIndex", i)

		if err != nil {
			return err
		}

		added, err := setDnsesIfMatch(itemRaw, adapterName, dnses, add)

		if err != nil {
			return err
		}

		if added {
			return nil
		}
	}

	return fmt.Errorf("setDnses() - interface not found")
}

func GetNetworkAdaptersConfigurations() ([]*NetworkAdapterConfiguration, error) {

	nacs, err := getNetworkAdaptersConfigurations("")

	if err == nil {
		return nacs.([]*NetworkAdapterConfiguration), nil
	} else {
		return nil, err
	}
}

func (nac *NetworkAdapterConfiguration) String() string {

	if nac == nil {
		return "<nil>"
	}

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf(`Caption: %s
Description: %s
SettingID: %s
ArpAlwaysSourceRoute: %v
ArpUseEtherSNAP: %v
DatabasePath: %s
DeadGWDetectEnabled: %v
DefaultIPGateway:`, nac.Caption, nac.Description, nac.SettingID, nac.ArpAlwaysSourceRoute, nac.ArpUseEtherSNAP,
		nac.DatabasePath, nac.DeadGWDetectEnabled))

	if nac.DefaultIPGateway == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.DefaultIPGateway) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.DefaultIPGateway {
			buffer.WriteString(fmt.Sprintf("    %s\n", item.String()))
		}
	}

	buffer.WriteString(fmt.Sprintf(`DefaultTOS: %d
DefaultTTL: %d
DHCPEnabled: %v
DHCPLeaseExpires: %v
DHCPLeaseObtained: %v
DHCPServer: %s
DNSDomain: %s
DNSDomainSuffixSearchOrder:`, nac.DefaultTOS, nac.DefaultTTL, nac.DHCPEnabled, nac.DHCPLeaseExpires,
		nac.DHCPLeaseObtained, nac.DHCPServer, nac.DNSDomain))

	if nac.DNSDomainSuffixSearchOrder == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.DNSDomainSuffixSearchOrder) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.DNSDomainSuffixSearchOrder {
			buffer.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}

	buffer.WriteString(fmt.Sprintf(`DNSEnabledForWINSResolution: %v
DNSHostName: %s
DNSServerSearchOrder:`, nac.DNSEnabledForWINSResolution, nac.DNSHostName))

	if nac.DNSServerSearchOrder == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.DNSServerSearchOrder) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.DNSServerSearchOrder {
			buffer.WriteString(fmt.Sprintf("    %s\n", item.String()))
		}
	}

	buffer.WriteString(fmt.Sprintf(`DomainDNSRegistrationEnabled: %v
ForwardBufferMemory: %d
FullDNSRegistrationEnabled: %v
IGMPLevel: %d
Index: %d
InterfaceIndex: %d
IPAddress:`, nac.DomainDNSRegistrationEnabled, nac.ForwardBufferMemory, nac.FullDNSRegistrationEnabled, nac.IGMPLevel,
		nac.Index, nac.InterfaceIndex))

	if nac.IPAddress == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPAddress) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPAddress {
			buffer.WriteString(fmt.Sprintf("    %s\n", item.String()))
		}
	}

	buffer.WriteString(fmt.Sprintf(`IPConnectionMetric: %d
IPEnabled: %v
IPFilterSecurityEnabled: %v
IPPortSecurityEnabled: %v
IPSecPermitIPProtocols:`, nac.IPConnectionMetric, nac.IPEnabled, nac.IPFilterSecurityEnabled, nac.IPPortSecurityEnabled))

	if nac.IPSecPermitIPProtocols == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPSecPermitIPProtocols) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPSecPermitIPProtocols {
			buffer.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}

	buffer.WriteString("IPSecPermitTCPPorts:")

	if nac.IPSecPermitTCPPorts == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPSecPermitTCPPorts) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPSecPermitTCPPorts {
			buffer.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}

	buffer.WriteString("IPSecPermitUDPPorts:")

	if nac.IPSecPermitUDPPorts == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPSecPermitUDPPorts) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPSecPermitUDPPorts {
			buffer.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}

	buffer.WriteString(fmt.Sprintf(`IPUseZeroBroadcast: %v
IPXAddress: %s
IPXEnabled: %v
IPXFrameType:`, nac.IPUseZeroBroadcast, nac.IPXAddress, nac.IPXEnabled))

	if nac.IPXFrameType == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPXFrameType) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPXFrameType {
			buffer.WriteString(fmt.Sprintf("    %d\n", item))
		}
	}

	buffer.WriteString(fmt.Sprintf("IPXMediaType: %d\nIPXNetworkNumber:", nac.IPXMediaType))

	if nac.IPXNetworkNumber == nil {
		buffer.WriteString(" <nil>\n")
	} else if len(nac.IPXNetworkNumber) < 1 {
		buffer.WriteString(" <empty>\n")
	} else {

		buffer.WriteString("\n")

		for _, item := range nac.IPXNetworkNumber {
			buffer.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}

	buffer.WriteString(fmt.Sprintf(`IPXVirtualNetNumber: %s
KeepAliveInterval: %d
KeepAliveTime: %d
MACAddress: %s
MTU: %d
NumForwardPackets: %d
PMTUBHDetectEnabled: %v
PMTUDiscoveryEnabled: %v
ServiceName: %s
TcpipNetbiosOptions: %d
TcpMaxConnectRetransmissions: %d
TcpMaxDataRetransmissions: %d
TcpNumConnections: %d
TcpUseRFC1122UrgentPointer: %v
TcpWindowSize: %d
WINSEnableLMHostsLookup: %v
WINSHostLookupFile: %s
WINSPrimaryServer: %s
WINSScopeID: %s
WINSSecondaryServer: %s
`, nac.IPXVirtualNetNumber, nac.KeepAliveInterval, nac.KeepAliveTime, nac.MACAddress, nac.MTU, nac.NumForwardPackets,
		nac.PMTUBHDetectEnabled, nac.PMTUDiscoveryEnabled, nac.ServiceName, nac.TcpipNetbiosOptions,
		nac.TcpMaxConnectRetransmissions, nac.TcpMaxDataRetransmissions, nac.TcpNumConnections, nac.TcpUseRFC1122UrgentPointer,
		nac.TcpWindowSize, nac.WINSEnableLMHostsLookup, nac.WINSHostLookupFile, nac.WINSPrimaryServer, nac.WINSScopeID,
		nac.WINSSecondaryServer))

	return buffer.String()
}
