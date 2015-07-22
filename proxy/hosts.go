package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

func weaveContainerIPs(container *docker.Container) ([]net.IP, error) {
	stdout, stderr, err := callWeave("ps", container.ID)
	if err != nil || len(stderr) > 0 {
		return nil, errors.New(string(stderr))
	}
	if len(stdout) <= 0 {
		return nil, nil
	}

	fields := strings.Fields(string(stdout))
	if len(fields) <= 2 {
		return nil, nil
	}

	var ips []net.IP
	for _, cidr := range fields[2:] {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func updateHosts(path, hostname string, ips []net.IP) error {
	hosts, err := parseHosts(path)
	if err != nil {
		return err
	}

	// Remove existing ips pointing to our hostname
	toRemove := []string{}
	for ip, addrs := range hosts {
		for _, addr := range addrs {
			if addr == hostname {
				toRemove = append(toRemove, ip)
				break
			}
		}
	}
	for _, ip := range toRemove {
		delete(hosts, ip)
	}

	// Add the weave ip(s)
	for _, ip := range ips {
		ipStr := ip.String()
		hosts[ipStr] = append(hosts[ipStr], hostname)
	}

	return writeHosts(path, hosts)
}

func parseHosts(path string) (map[string][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ips := map[string][]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Remove any comments
		if i := strings.IndexByte(line, '#'); i != -1 {
			line = line[:i]
		}

		fields := strings.Fields(line)
		if len(fields) > 0 {
			ips[fields[0]] = append(ips[fields[0]], fields[1:]...)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return ips, nil
}

func writeHosts(path string, contents map[string][]string) error {
	ips := []string{}
	for ip := range contents {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, "# modified by weave")
	for _, ip := range ips {
		if addrs := contents[ip]; len(addrs) > 0 {
			fmt.Fprintf(buf, "%s\t%s\n", ip, strings.Join(uniqueStrs(addrs), " "))
		}
	}
	return ioutil.WriteFile(path, buf.Bytes(), 644)
}

func uniqueStrs(s []string) []string {
	m := map[string]struct{}{}
	result := []string{}
	for _, str := range s {
		if _, ok := m[str]; !ok {
			m[str] = struct{}{}
			result = append(result, str)
		}
	}
	sort.Strings(result)
	return result
}
