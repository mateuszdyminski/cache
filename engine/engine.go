package engine

import (
	"fmt"
	"sync"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/sirupsen/logrus"
	"os"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

type Cache struct {
	mx      sync.Mutex
	values  map[string][]byte
	port    int
	peers   []string
	logger  *logrus.Logger
	hClient *http.Client
}

func NewEngine(peers []string, port int) (*Cache, error) {
	// init logger
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	//logger.Formatter = &logrus.JSONFormatter{}

	// check hostname
	host, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("can't check hostname: %v", err)
	}

	// remove our own entry from peers
	toRemove := 0
	for i, v := range peers {
		if v == fmt.Sprintf("%s:%d", host, port) {
			toRemove = i
		}
	}

	peers = append(peers[:toRemove], peers[toRemove+1:]...)

	//logger.Infof("To remove: %+v", toRemove)
	return &Cache{
		peers:   peers,
		values:  make(map[string][]byte),
		port:    port,
		logger:  logger,
		hClient: &http.Client{},
	}, nil
}

func (c *Cache) Start() {

	// start http server
	router := mux.NewRouter()
	l := Middleware{
		Name:   "cache",
		Logger: c.logger,
	}

	router.HandleFunc("/put/{key}", c.put).Methods("PUT")
	router.HandleFunc("/get/{key}", c.get).Methods("GET")
	router.HandleFunc("/all", c.all).Methods("GET")
	router.HandleFunc("/delete/{key}", c.delete).Methods("DELETE")
	router.HandleFunc("/sync/{key}", c.sync).Methods("PUT", "DELETE")

	c.logger.Infof("Cache started on port: %d", c.port)
	c.logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.port), l.Handler(router, "")))
}

func (c *Cache) sync(rw http.ResponseWriter, r *http.Request) {
	c.mx.Lock()
	defer c.mx.Unlock()
	key := mux.Vars(r)["key"]

	if key == "" {
		c.err(rw, http.StatusBadRequest, fmt.Errorf("key is path is required"))
		return
	}

	if r.Method == "PUT" {
		defer r.Body.Close()
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't sync value with key: %s, value: %s", key, string(bytes)))
			return
		}

		c.logger.Debugf("[sync] setting value for key: %s and value: %s", key, string(bytes))
		c.values[key] = bytes
	}

	if r.Method == "DELETE" {
		c.logger.Debugf("[sync] deleting value for key: %s", key)
		delete(c.values, key)
	}
}

func (c *Cache) get(rw http.ResponseWriter, r *http.Request) {
	c.mx.Lock()
	defer c.mx.Unlock()
	key := mux.Vars(r)["key"]

	if key == "" {
		c.err(rw, http.StatusBadRequest, fmt.Errorf("key is path is required"))
		return
	}

	val, ok := c.values[key]
	if ok {
		fmt.Fprint(rw, val)
		return
	}

	c.err(rw, http.StatusNotFound, fmt.Errorf("can't find value with key: %s", key))
	return
}

func (c *Cache) put(rw http.ResponseWriter, r *http.Request) {
	c.mx.Lock()
	defer c.mx.Unlock()
	key := mux.Vars(r)["key"]

	if key == "" {
		c.err(rw, http.StatusBadRequest, fmt.Errorf("key is path is required"))
		return
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't put value with key: %s", key))
		return
	}

	c.values[key] = data

	// sync value with other nodes
	for _, peer := range c.peers {
		adr := fmt.Sprintf("http://%s/sync/%s", peer, key)
		c.logger.Debugf("sync value with address: %s", adr)
		req, err := http.NewRequest("PUT", adr, bytes.NewReader(data))
		if err != nil {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't sync value with key: %s and peer: %s. err: %s", key, peer, err.Error()))
			return
		}

		resp, err := c.hClient.Do(req)
		if err != nil {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't sync value with key: %s and peer: %s. err: %s", key, peer, err.Error()))
			return
		}

		if resp.StatusCode != http.StatusOK {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("wrong http response code: %d", resp.StatusCode))
			return
		}

		c.logger.Debugf("sync key: %s with peer: %s done!", key, peer)
	}
}

func (c *Cache) delete(rw http.ResponseWriter, r *http.Request) {
	c.mx.Lock()
	defer c.mx.Unlock()
	key := mux.Vars(r)["key"]

	if key == "" {
		c.err(rw, http.StatusBadRequest, fmt.Errorf("key is path is required"))
		return
	}

	delete(c.values, key)

	// sync value with other nodes
	for _, peer := range c.peers {
		adr := fmt.Sprintf("http://%s/sync/%s", peer, key)
		c.logger.Debugf("sync value with address: %s", adr)
		req, err := http.NewRequest("DELETE", adr, nil)
		if err != nil {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't sync value with key: %s and peer: %s. err: %s", key, peer, err.Error()))
			return
		}

		resp, err := c.hClient.Do(req)
		if err != nil {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't sync value with key: %s and peer: %s. err: %s", key, peer, err.Error()))
			return
		}

		if resp.StatusCode != http.StatusOK {
			c.err(rw, http.StatusInternalServerError, fmt.Errorf("wrong http response code: %d", resp.StatusCode))
			return
		}
	}
}

func (c *Cache) all(rw http.ResponseWriter, r *http.Request) {
	c.mx.Lock()
	defer c.mx.Unlock()

	bytes, err := json.Marshal(c.values)
	if err != nil {
		c.err(rw, http.StatusInternalServerError, fmt.Errorf("can't marshal map. err: %s", err.Error()))
		return
	}

	fmt.Fprint(rw, bytes)
}

func (c *Cache) err(rw http.ResponseWriter, errCode int, err error) {
	c.logger.Error(err.Error())
	rw.WriteHeader(errCode)
	fmt.Fprint(rw, err.Error())
}
