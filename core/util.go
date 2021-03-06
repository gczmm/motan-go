package core

import (
	"bytes"
	"math/rand"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode"

	"github.com/weibocom/motan-go/log"
)

const (
	defaultServerPort = "9982"
	defaultProtocal   = "motan2"
)

var (
	PanicStatFunc func()

	localIPs = make([]string, 0)
)

func ParseExportInfo(export string) (string, int, error) {
	port := defaultServerPort
	protocol := defaultProtocal
	s := TrimSplit(export, ":")
	if len(s) == 1 && s[0] != "" {
		port = s[0]
	} else if len(s) == 2 {
		if s[0] != "" {
			protocol = s[0]
		}
		port = s[1]
	}
	porti, err := strconv.Atoi(port)
	if err != nil {
		vlog.Errorf("export port not int. port:%s", port)
		return protocol, porti, err
	}
	return protocol, porti, err
}

func InterfaceToString(in interface{}) string {
	rs := ""
	switch in.(type) {
	case int:
		rs = strconv.Itoa(in.(int))
	case float64:
		rs = strconv.FormatFloat(in.(float64), 'f', -1, 64)
	case string:
		rs = in.(string)
	case bool:
		rs = strconv.FormatBool(in.(bool))
	}
	return rs
}

// GetLocalIPs ip from ipnet
func GetLocalIPs() []string {
	if len(localIPs) == 0 {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			vlog.Warningf("get local ip fail. %s", err.Error())
		} else {
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						localIPs = append(localIPs, ipnet.IP.String())
					}
				}
			}
		}
	}
	return localIPs
}

// GetLocalIP falg of localIP > ipnet
func GetLocalIP() string {
	if *LocalIP != "" {
		return *LocalIP
	} else if len(GetLocalIPs()) > 0 {
		return GetLocalIPs()[0]
	}
	return "unknown"
}

func SliceShuffle(slice []string) []string {
	for i := 0; i < len(slice); i++ {
		a := rand.Intn(len(slice))
		b := rand.Intn(len(slice))
		slice[a], slice[b] = slice[b], slice[a]
	}
	return slice
}

func FirstUpper(s string) string {
	r := []rune(s)

	if unicode.IsUpper(r[0]) {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func GetReqInfo(request Request) string {
	if request != nil {
		var buffer bytes.Buffer
		buffer.WriteString("req{")
		buffer.WriteString(strconv.FormatUint(request.GetRequestID(), 10))
		buffer.WriteString(",")
		buffer.WriteString(request.GetServiceName())
		buffer.WriteString(",")
		buffer.WriteString(request.GetMethod())
		buffer.WriteString("}")
		return buffer.String()
	}
	return ""
}

func HandlePanic(f func()) {
	if err := recover(); err != nil {
		vlog.Errorf("recover panic. error:%v, stack: %s", err, debug.Stack())
		if f != nil {
			f()
		}
		if PanicStatFunc != nil {
			PanicStatFunc()
		}
	}
}

// TrimSplit slices s into all substrings separated by sep and
// returns a slice of the substrings between those separators,
// specially trim all substrings.
func TrimSplit(s string, sep string) []string {
	n := strings.Count(s, sep) + 1
	a := make([]string, n)
	i := 0
	if sep == "" {
		return strings.Split(s, sep)
	}
	for {
		m := strings.Index(s, sep)
		if m < 0 {
			s = strings.TrimSpace(s)
			break
		}
		a[i] = strings.TrimSpace(s[:m])
		i++
		s = s[m+len(sep):]
	}
	a[i] = s
	return a[:i+1]
}

// ListenUnixSock try to listen a unix socket address
// this method using by create motan agent server, management server and http proxy server
func ListenUnixSock(unixSockAddr string) (net.Listener, error) {
	if err := os.RemoveAll(unixSockAddr); err != nil {
		vlog.Errorf("listenUnixSock err, remove old unix sock file fail. err: %v", err)
		return nil, err
	}

	listener, err := net.Listen("unix", unixSockAddr)
	if err != nil {
		vlog.Errorf("listenUnixSock err, listen unix sock fail. err:%v", err)
		return nil, err
	}
	return listener, nil
}
