package net

import (
	"fmt"
	"net"
	"syscall"

	"github.com/vishvananda/netlink/nl"
)

// Wait `wait` seconds for an interface to come up. Pass zero to check once
// and return immediately, or a negative value to wait indefinitely.
func EnsureInterface(ifaceName string, wait int) (iface *net.Interface, err error) {
	s, err := nl.Subscribe(syscall.NETLINK_ROUTE, syscall.RTNLGRP_LINK)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	if iface, err = findInterface(ifaceName); err == nil || wait == 0 {
		return
	}
	waitForIfUp(s, ifaceName)
	iface, err = findInterface(ifaceName)
	return
}

func findInterface(ifaceName string) (iface *net.Interface, err error) {
	if iface, err = net.InterfaceByName(ifaceName); err != nil {
		return iface, fmt.Errorf("Unable to find interface %s", ifaceName)
	}
	if 0 == (net.FlagUp & iface.Flags) {
		return iface, fmt.Errorf("Interface %s is not up", ifaceName)
	}
	return
}

func waitForIfUp(s *nl.NetlinkSocket, ifaceName string) error {
	for {
		msgs, err := s.Receive()
		if err != nil {
			return err
		}
		for _, m := range msgs {
			switch m.Header.Type {
			case syscall.RTM_NEWLINK: // receive this type for link 'up'
				ifmsg := nl.DeserializeIfInfomsg(m.Data)
				attrs, err := syscall.ParseNetlinkRouteAttr(&m)
				if err != nil {
					return err
				}
				name := ""
				for _, attr := range attrs {
					if attr.Attr.Type == syscall.IFA_LABEL {
						name = string(attr.Value[:len(attr.Value)-1])
					}
				}
				up := ifmsg.Flags&syscall.IFF_UP != 0
				if ifaceName == name && up {
					return nil
				}
			}
		}
	}
}
