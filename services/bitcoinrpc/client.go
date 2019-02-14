package bitcoinrpc

import (
	"container/list"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ConnConfig describes the connection configuration parameters for the client.
// This
type ConnConfig struct {
	// Host is the IP address and port of the RPC server you want to connect
	// to.
	Host string

	// Endpoint is the websocket endpoint on the RPC server.  This is
	// typically "ws".
	Endpoint string

	// User is the username to use to authenticate to the RPC server.
	User string

	// Pass is the passphrase to use to authenticate to the RPC server.
	Pass string

	// DisableTLS specifies whether transport layer security should be
	// disabled.  It is recommended to always use TLS if the RPC server
	// supports it as otherwise your username and password is sent across
	// the wire in cleartext.
	DisableTLS bool

	// Certificates are the bytes for a PEM-encoded certificate chain used
	// for the TLS connection.  It has no effect if the DisableTLS parameter
	// is true.
	Certificates []byte

	// Proxy specifies to connect through a SOCKS 5 proxy server.  It may
	// be an empty string if a proxy is not required.
	Proxy string

	// ProxyUser is an optional username to use for the proxy server if it
	// requires authentication.  It has no effect if the Proxy parameter
	// is not set.
	ProxyUser string

	// ProxyPass is an optional password to use for the proxy server if it
	// requires authentication.  It has no effect if the Proxy parameter
	// is not set.
	ProxyPass string

	// DisableAutoReconnect specifies the client should not automatically
	// try to reconnect to the server when it has been disconnected.
	DisableAutoReconnect bool

	// DisableConnectOnNew specifies that a websocket client connection
	// should not be tried when creating the client with New.  Instead, the
	// client is created and returned unconnected, and Connect must be
	// called manually.
	DisableConnectOnNew bool

	// HTTPPostMode instructs the client to run using multiple independent
	// connections issuing HTTP POST requests instead of using the default
	// of websockets.  Websockets are generally preferred as some of the
	// features of the client such notifications only work with websockets,
	// however, not all servers support the websocket extensions, so this
	// flag can be set to true to use basic HTTP POST requests instead.
	HTTPPostMode bool

	// EnableBCInfoHacks is an option provided to enable compatibility hacks
	// when connecting to blockchain.info RPC server
	EnableBCInfoHacks bool
}

// Client represents a Bitcoin RPC client which allows easy access to the
// various RPC methods available on a Bitcoin RPC server.  Each of the wrapper
// functions handle the details of converting the passed and return types to and
// from the underlying JSON types which are required for the JSON-RPC
// invocations
//
// The client provides each RPC in both synchronous (blocking) and asynchronous
// (non-blocking) forms.  The asynchronous forms are based on the concept of
// futures where they return an instance of a type that promises to deliver the
// result of the invocation at some future time.  Invoking the Receive method on
// the returned future will block until the result is available if it's not
// already.
type Client struct {
	id uint64 // atomic, so must stay 64-bit aligned

	// config holds the connection configuration assoiated with this client.
	config *ConnConfig

	// wsConn is the underlying websocket connection when not in HTTP POST
	// mode.
	wsConn *websocket.Conn

	// httpClient is the underlying HTTP client to use when running in HTTP
	// POST mode.
	httpClient *http.Client

	// mtx is a mutex to protect access to connection related fields.
	mtx sync.Mutex

	// disconnected indicated whether or not the server is disconnected.
	disconnected bool

	// retryCount holds the number of times the client has tried to
	// reconnect to the RPC server.
	retryCount int64

	// Track command and their response channels by ID.
	requestLock sync.Mutex
	requestMap  map[uint64]*list.Element
	requestList *list.List

	// Notifications.
	ntfnHandlers  *NotificationHandlers
	ntfnStateLock sync.Mutex
	ntfnState     *notificationState

	// Networking infrastructure.
	sendChan        chan []byte
	sendPostChan    chan *sendPostDetails
	connEstablished chan struct{}
	disconnect      chan struct{}
	shutdown        chan struct{}
	wg              sync.WaitGroup
}

// New creates a new RPC client based on the provided connection configuration
// details.  The notification handlers parameter may be nil if you are not
// interested in receiving notifications and will be ignored if the
// configuration is set to run in HTTP POST mode.
func New(config *ConnConfig, ntfnHandlers *NotificationHandlers) (*Client, error) {
	// Either open a websocket connection or create an HTTP client depending
	// on the HTTP POST mode.  Also, set the notification handlers to nil
	// when running in HTTP POST mode.
	var wsConn *websocket.Conn
	var httpClient *http.Client
	connEstablished := make(chan struct{})
	var start bool
	if config.HTTPPostMode {
		ntfnHandlers = nil
		start = true

		var err error
		httpClient, err = newHTTPClient(config)
		if err != nil {
			return nil, err
		}
	} else {
		if !config.DisableConnectOnNew {
			var err error
			wsConn, err = dial(config)
			if err != nil {
				return nil, err
			}
			start = true
		}
	}

	client := &Client{
		config:          config,
		wsConn:          wsConn,
		httpClient:      httpClient,
		requestMap:      make(map[uint64]*list.Element),
		requestList:     list.New(),
		ntfnHandlers:    ntfnHandlers,
		ntfnState:       newNotificationState(),
		sendChan:        make(chan []byte, sendBufferSize),
		sendPostChan:    make(chan *sendPostDetails, sendPostBufferSize),
		connEstablished: connEstablished,
		disconnect:      make(chan struct{}),
		shutdown:        make(chan struct{}),
	}

	if start {
		log.Infof("Established connection to RPC server %s",
			config.Host)
		close(connEstablished)
		client.start()
		if !client.config.HTTPPostMode && !client.config.DisableAutoReconnect {
			client.wg.Add(1)
			go client.wsReconnectHandler()
		}
	}

	return client, nil
}
