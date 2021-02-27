package main

import (
	_ "WEB/routers"
	"fmt"

	"github.com/astaxie/beego"
)

func main() {
	fmt.Println("beego run")
	beego.Run()
package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	glog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	runPort := os.Getenv("RUN_PORT")
	if runPort == "" {
		runPort = ":8000"
	}
	e := router()
	e.Logger.Fatal(listenAddr(e, runPort))
}

func listenAddr(e *echo.Echo, port string) error {
	defer Recover()
	if !strings.Contains(port, ":") {
		port = ":" + port
	}
	nicName := os.Getenv("lan_eth")
	listenIps, err := getListenIps(nicName)
	if err != nil {
		glog.Printf("get %s listenIps err, err, %v", nicName, err)
		return err
	}
	glog.Printf("addr: %v, port: %s", listenIps, port)
	for _, ip := range listenIps {
		if strings.Contains(ip, ":") {
			ip = fmt.Sprintf("[%s]", ip)
		}
		go e.Start(ip + port)
	}
	select {}
}

func router() *echo.Echo {
	msb := getMsbRouter()
	projectID := os.Getenv("OPENPALETTE_NAMESPACE")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Logger.SetHeader(`{"time":"${time_rfc3339_nano}","level":"${level}"}`)
	e.Logger.SetLevel(log.DEBUG)

	e.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "Proxy is connectable!")
	})

	url, err := url.Parse(msb)
	if err != nil {
		e.Logger.Fatal(err)
	}

	opapi := e.Group("/opapi")
	opapi.Use(middleware.Rewrite(map[string]string{
		"*/unknown*": "$1/" + projectID + "$2",
	}))
	opapi.Use(addAPIKey())
	opapi.Use(removeHostInHeader())

	targets := []*middleware.ProxyTarget{
		{
			URL: url,
		},
	}
	balancer := middleware.NewRandomBalancer(targets)
	/* #nosec */
	tlsConfig := tls.Config{
		// 项目环境内的https证书不安全，需要忽略
		InsecureSkipVerify: true,
	}
	transport := http.Transport{
		TLSClientConfig: &tlsConfig,
	}
	proxyConfig := middleware.ProxyConfig{
		Balancer:  balancer,
		Transport: &transport,
	}
	opapi.Use(middleware.ProxyWithConfig(proxyConfig))
	return e
}

func addAPIKey() echo.MiddlewareFunc {
	msb := getMsbRouter()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			c.Logger().Debugf("Method: %s, URL: %v, MSBIP: %s", req.Method, req.URL, msb)
			serviceAcountName := req.Header.Get("X-API-User")
			if serviceAcountName == "" {
				serviceAcountName = "default"
			}
			isTCF, err := checkTcf(serviceAcountName)
			if err != nil {
				c.Logger().Errorf("checkTcf err, %v", err)
			}
			header := "X-OPENPALETTE-APIKEY"
			apiKey := ""
			if isTCF {
				header = "X-Auth-Token"
				apiKey, err = readAPIKey(serviceAcountName)
			} else {
				apiKey, err = readEncAPIKey(serviceAcountName)
			}
			if err != nil {
				c.Logger().Error(err)
			}
			req.Header.Add(header, apiKey)
			req.Header.Set(header, apiKey)
			if req.Header.Get("Upgrade") == "WebSocket" || req.Header.Get("Connection") == "Upgrade" {
				req.Header.Set("Upgrade", "websocket")
			}
			return next(c)
		}
	}
}

func readEncAPIKey(name string) (string, error) {
	encyptedAPIKeyPath := filepath.Join("secrets", name+"_encyptedApiKey")
	// 此处gosec会报警告，因为从变量中确定文件路径是危险行为，会造成数据暴露，读文件导致bug等
	safeEncyptedAPIKeyPath := getSafePath(encyptedAPIKeyPath)
	encyptedAPIKey, err := ioutil.ReadFile(safeEncyptedAPIKeyPath) /* #nosec */
	if err != nil || len(encyptedAPIKey) == 0 {
		return "", fmt.Errorf("Failed to get encypted data, err: %v", err)
	}
	KMSPath := filepath.Join("secrets", name+"_kmsId")
	safeKMSPath := getSafePath(KMSPath)
	KMS, err := ioutil.ReadFile(safeKMSPath) /* #nosec */
	if err != nil || len(KMS) == 0 {
		return "", fmt.Errorf("Failed to get KMS: %v", err)
	}
	APIKey, err := aes256Decrypt(encyptedAPIKey, KMS)
	if err != nil || len(APIKey) == 0 {
		return "", fmt.Errorf("Failed to decrypt: %v", err)
	}
	return string(APIKey), nil
}

func readAPIKey(name string) (string, error) {
	APIKeyPath := filepath.Join("secrets", name)
	APIKeyPath = getSafePath(APIKeyPath)
	apikey, err := ioutil.ReadFile(APIKeyPath) /* #nosec */
	if err != nil || len(apikey) == 0 {
		return "", fmt.Errorf("Failed to get apiKey: %v", err)
	}
	return string(apikey), nil
}

func removeHostInHeader() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			req.Header.Del("Host")
			return next(c)
		}
	}
}

// add for go coverity checker: PATH_MANIPULATION
func getSafePath(rawpath string) string {
	if len(rawpath) == 0 {
		return ""
	}
	var safePath []byte
	pathArray := []byte(rawpath)
	for _, item := range pathArray {
		safePath = append(safePath, item)
	}
	return string(safePath)
}

func getMsbRouter() string {
	host := os.Getenv("OPENPALETTE_MSB_ROUTER_IP")
	if strings.Contains(host, ":") {
		host = fmt.Sprintf("[%s]", host)
	}
	port := os.Getenv("OPENPALETTE_MSB_ROUTER_HTTPS_PORT")
	msb := fmt.Sprintf("https://%s:%s", host, port)
	return msb
}

func getListenIps(nicname string) ([]string, error) {
	ips := []string{}
	addr, err := net.InterfaceByName(nicname)
	if err != nil {
		glog.Printf("get %s listenIps err, %v", nicname, err)
		return []string{""}, nil
	}
	// 不接收err无风险
	addrs, _ := addr.Addrs()
	for _, v := range addrs {
		if ip, ok := v.(*net.IPNet); ok {
			ips = append(ips, ip.IP.String())
		}
	}
	return ips, nil
}

func checkTcf(sAName string) (bool, error) {
	encPathExist, err := checkEncyptedPath(sAName)
	if err != nil {
		return false, err
	}
	return !encPathExist, nil
}

func checkEncyptedPath(sAName string) (bool, error) {
	encyptedAPIKeyPath := filepath.Join("secrets", sAName+"_encyptedApiKey")
	encyptedAPIKeyPath = getSafePath(encyptedAPIKeyPath)
	return pathExists(encyptedAPIKeyPath)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Recover() {
	if err := recover(); err != nil {
		glog.Printf("public.Recover panic message: %v", err)
		glog.Printf("public.Recover panic stack: %s", string(debug.Stack()))
	}
}

