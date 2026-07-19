package drill

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
)

// drillNet drills into network connections and host information.
func drillNet(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// Default: show local network info
		return drillNetLocal(stdout, stderr)
	}

	host := args[0]

	// Check if it's host:port format
	if strings.Contains(host, ":") {
		return drillNetConnect(host, stdout, stderr)
	}

	// DNS lookup
	return drillNetDNS(host, stdout, stderr)
}

// drillNetLocal shows local network information.
func drillNetLocal(stdout, stderr io.Writer) int {
	ctx := cueutil.NewContext()
	data := make(map[string]any)

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		data["hostname"] = hostname
	}

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		var ifaces []map[string]any
		for _, iface := range interfaces {
			addrs, _ := iface.Addrs()
			addrStrs := make([]string, 0, len(addrs))
			for _, addr := range addrs {
				addrStrs = append(addrStrs, addr.String())
			}
			ifaces = append(ifaces, map[string]any{
				"name":         iface.Name,
				"mtu":          iface.MTU,
				"hardwareAddr": iface.HardwareAddr.String(),
				"flags":        iface.Flags.String(),
				"addresses":    addrStrs,
			})
		}
		data["interfaces"] = ifaces
	}

	v := ctx.Encode(data)
	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "drill net: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}

// drillNetConnect connects to a host:port and shows TLS/connection info.
func drillNetConnect(host string, stdout, stderr io.Writer) int {
	hostname, portStr, err := net.SplitHostPort(host)
	if err != nil {
		fmt.Fprintf(stderr, "drill net: invalid host:port: %s\n", host)
		return 1
	}

	port, _ := strconv.Atoi(portStr)

	ctx := cueutil.NewContext()
	data := map[string]any{
		"host": hostname,
		"port": port,
	}

	// TCP connection attempt
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err == nil {
		defer conn.Close()
		data["reachable"] = true
		data["localAddr"] = conn.LocalAddr().String()
		data["remoteAddr"] = conn.RemoteAddr().String()

		// HTTP check
		if port == 80 || port == 8080 {
			fmt.Fprintf(conn, "GET / HTTP/1.0\r\nHost: %s\r\n\r\n", hostname)
			resp := make([]byte, 512)
			n, _ := conn.Read(resp)
			data["httpResponse"] = string(resp[:n])
		}
	} else {
		data["reachable"] = false
		data["error"] = err.Error()
	}

	v := ctx.Encode(data)
	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "drill net: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}

// drillNetDNS performs a DNS lookup.
func drillNetDNS(host string, stdout, stderr io.Writer) int {
	ctx := cueutil.NewContext()
	data := make(map[string]any)
	data["host"] = host

	ips, err := net.LookupHost(host)
	if err == nil {
		data["addresses"] = ips
	} else {
		data["error"] = err.Error()
	}

	v := ctx.Encode(data)
	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "drill net: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, str)
	return 0
}
