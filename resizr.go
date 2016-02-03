package main

import (
	"flag"
	"fmt"
	. "github.com/tj/go-debug"
	"os"
	"runtime"
	d "runtime/debug"
	"strconv"
	"time"
)

var debug = Debug("resizr")

var (
	aAddr            = flag.String("a", "", "bind address")
	aPort            = flag.Int("p", 9000, "port to listen")
	aVers            = flag.Bool("v", false, "Show version")
	aVersl           = flag.Bool("version", false, "Show version")
	aHelp            = flag.Bool("h", false, "Show help")
	aHelpl           = flag.Bool("help", false, "Show help")
	aCors            = flag.Bool("cors", false, "Enable CORS support")
	aGzip            = flag.Bool("gzip", false, "Enable gzip compression")
	aEnableURLSource = flag.Bool("enable-url-source", false, "Enable remote HTTP URL image source processing")
	aKey             = flag.String("key", "", "Define API key for authorization")
	aMount           = flag.String("mount", "", "Mount server local directory")
	aCertFile        = flag.String("certfile", "", "TLS certificate file path")
	aKeyFile         = flag.String("keyfile", "", "TLS private key file path")
	aHttpCacheTtl    = flag.Int("http-cache-ttl", -1, "The TTL in seconds")
	aReadTimeout     = flag.Int("http-read-timeout", 30, "HTTP read timeout in seconds")
	aWriteTimeout    = flag.Int("http-write-timeout", 30, "HTTP write timeout in seconds")
	aConcurrency     = flag.Int("concurrency", 0, "Throttle concurrency limit per second")
	aBurst           = flag.Int("burst", 100, "Throttle burst max cache size")
	aMRelease        = flag.Int("mrelease", 30, "OS memory release inverval in seconds")
	aCpus            = flag.Int("cpus", runtime.GOMAXPROCS(-1), "Number of cpu cores to use")
)

const usage = `resizr %s

Usage:
  resizr -p 80
  resizr -cors

Options:
  -a <addr>                 bind address [default: *]
  -p <port>                 bind port [default: 9000]
  -h, -help                 output help
  -v, -version              output version
  -cors                     Enable CORS support [default: false]
  -gzip                     Enable gzip compression [default: false]
  -key <key>                Define API key for authorization
  -http-cache-ttl <num>     The TTL in seconds. Adds caching headers to locally served files.
  -http-read-timeout <num>  HTTP read timeout in seconds [default: 30]
  -http-write-timeout <num> HTTP write timeout in seconds [default: 30]
  -certfile <path>          TLS certificate file path
  -keyfile <path>           TLS private key file path
  -concurreny <num>         Throttle concurrency limit per second [default: disabled]
  -burst <num>              Throttle burst max cache size [default: 100]
  -mrelease <num>           OS memory release inverval in seconds [default: 30]
  -cpus <num>               Number of used cpu cores.
                            (default for current machine is %d cores)
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, Version, runtime.NumCPU()))
	}
	flag.Parse()

	if *aHelp || *aHelpl {
		showUsage()
	}
	if *aVers || *aVersl {
		showVersion()
	}

	// Only required in Go < 1.5
	runtime.GOMAXPROCS(*aCpus)

	port := getPort(*aPort)
	opts := ServerOptions{
		Port:             port,
		Address:          *aAddr,
		Gzip:             *aGzip,
		CORS:             *aCors,
		ApiKey:           *aKey,
		Concurrency:      *aConcurrency,
		Burst:            *aBurst,
		CertFile:         *aCertFile,
		KeyFile:          *aKeyFile,
		HttpCacheTtl:     *aHttpCacheTtl,
		HttpReadTimeout:  *aReadTimeout,
		HttpWriteTimeout: *aWriteTimeout,
	}

	// Create a memory release goroutine
	if *aMRelease > 0 {
		memoryRelease(*aMRelease)
	}

	// Validate HTTP cache param, if present
	if *aHttpCacheTtl != -1 {
		checkHttpCacheTtl(*aHttpCacheTtl)
	}

	debug("resizr server listening on port %d", port)

	// Start the server
	err := Server(opts)
	if err != nil {
		exitWithError("cannot start the server: %s\n", err)
	}
}

func getPort(port int) int {
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		newPort, _ := strconv.Atoi(portEnv)
		if newPort > 0 {
			port = newPort
		}
	}
	return port
}

func showUsage() {
	flag.Usage()
	os.Exit(1)
}

func showVersion() {
	fmt.Println(Version)
	os.Exit(1)
}

func checkHttpCacheTtl(ttl int) {
	if ttl < -1 || ttl > 31556926 {
		exitWithError("The -http-cache-ttl flag only accepts a value from 0 to 31556926")
	}

	if ttl == 0 {
		debug("Adding HTTP cache control headers set to prevent caching.")
	}
}

func memoryRelease(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for _ = range ticker.C {
			debug("FreeOSMemory()")
			d.FreeOSMemory()
		}
	}()
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args)
	os.Exit(1)
}
