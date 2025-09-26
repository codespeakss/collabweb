// file: private_ips.go
package main

import (
	"fmt"
	"net"
	"sort"
)

// isPrivateIPv4 检查 IPv4 是否在 RFC1918 私有网段
func isPrivateIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	b0 := ip4[0]
	b1 := ip4[1]

	switch {
	case b0 == 10:
		return true // 10.0.0.0/8
	case b0 == 172 && b1 >= 16 && b1 <= 31:
		return true // 172.16.0.0/12
	case b0 == 192 && b1 == 168:
		return true // 192.168.0.0/16
	default:
		return false
	}
}

// isPrivateIPv6 检查 IPv6 是否为 ULA (fc00::/7)
func isPrivateIPv6(ip net.IP) bool {
	ip16 := ip.To16()
	if ip16 == nil || ip.To4() != nil {
		return false
	}
	// ULA: first 7 bits == 0b1111110 -> 0xfc or 0xfd as first byte (fc00::/7)
	return (ip16[0] & 0xfe) == 0xfc
}

// GetPrivateIPs 枚举并返回所有私有/内网地址（IPv4 + IPv6 ULA）
func GetPrivateIPs() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{})

	for _, iface := range ifaces {
		// 跳过未启用或回环接口
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			// 某些接口可能无法读取地址，忽略错误继续
			continue
		}

		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// 过滤 link-local 地址（可选，这里我们排除 IPv4 link-local 169.254.x.x 和 IPv6 fe80::/10）
			if ip4 := ip.To4(); ip4 != nil {
				if ip4[0] == 169 && ip4[1] == 254 {
					continue
				}
				if isPrivateIPv4(ip4) {
					set[ip.String()] = struct{}{}
				}
			} else {
				// IPv6
				// 排除链路本地 fe80::/10
				if (ip[0] == 0xfe) && ((ip[1]&0xc0) == 0x80) {
					continue
				}
				if isPrivateIPv6(ip) {
					set[ip.String()] = struct{}{}
				}
			}
		}
	}

	// 将集合转为排序后的切片以便可预测输出
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}

func main() {
	ips, err := GetPrivateIPs()
	if err != nil {
		fmt.Printf("获取内网地址失败: %v\n", err)
		return
	}
	if len(ips) == 0 {
		fmt.Println("未发现内网私有地址")
		return
	}
	fmt.Println("本机内网私有地址：")
	for _, ip := range ips {
		fmt.Println(" -", ip)
	}
}

