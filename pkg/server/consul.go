package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
)

type WatchDefinition struct {
	admin.MysqlServer
	ServiceDefinition
}

type ServiceDefinition struct {
	Note        string   `json:"note"`
	Primary     bool     `json:"primary"`
	ServiceName string   `json:"service_name"`
	Tags        []string `json:"tags"`
	RejectTags  []string `json:"reject_tags"`
}

func (w *WatchDefinition) UnmarshalJSON(data []byte) error {
	log.Printf("here!\n")
	err := w.MysqlServer.UnmarshalJSON(data)
	if err != nil && err != admin.ErrUnmarshalNullHostname {
		// admin.MysqlServer has a custom unmarshaler that returns this error if no hostname is provided.
		// The ConsulWatch will be overwritting the hostname any way.
		return err
	}
	return json.Unmarshal(data, &w.ServiceDefinition)
}

type WatchDefinitionSlice []WatchDefinition

// See https://github.com/kelseyhightower/envconfig#custom-decoders
func (c *WatchDefinitionSlice) Decode(value string) error {

	// unmarshal the mysql_server w/ defaults
	err := json.Unmarshal([]byte(value), c)
	if err == admin.ErrUnmarshalNullHostname {
		// admin.MysqlServer has a custom unmarshaler that returns this error if no hostname is provided.
		// The ConsulWatch will be overwritting the hostname any way.
		return nil
	}
	return err
}

type consulServiceWatcher struct {
	def    WatchDefinition
	plans  []*watch.Plan
	ch     chan<- hostgroupUpdate
	addr   string
	errLog consulLogBuffer
}

type hostgroupUpdate struct {
	hostgroupID  int
	primary      bool
	mysqlServers []admin.MysqlServer
}

func newConsulServiceWatcher(consulAddress string, primaryDCs []string, def WatchDefinition, ch chan<- hostgroupUpdate) (*consulServiceWatcher, error) {
	c := &consulServiceWatcher{
		def:  def,
		ch:   ch,
		addr: consulAddress,
	}

	if !def.Primary || len(primaryDCs) == 0 {
		err := c.addPlan("")
		return c, err
	}

	for _, dc := range primaryDCs {
		err := c.addPlan(dc)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *consulServiceWatcher) addPlan(dc string) error {

	m := map[string]interface{}{"type": "service", "service": c.def.ServiceName, "passingonly": true}
	if len(dc) != 0 {
		m["datacenter"] = dc
	}
	plan, err := watch.Parse(m)
	if err != nil {
		return err
	}
	plan.Handler = c.handler
	plan.LogOutput = &c.errLog
	c.plans = append(c.plans, plan)
	return nil
}

func (c *consulServiceWatcher) Start() {
	for _, p := range c.plans {
		go func(plan *watch.Plan) {
			for {
				err := plan.Run(c.addr)
				if err != nil {
					msg := fmt.Sprintf("Error running plan: %s", err.Error())
					c.errLog.Write([]byte(msg))
				}
				time.Sleep(2 * time.Second)
			}
		}(p)
	}
}

func (c *consulServiceWatcher) CheckHealth() error {
	lastErr := c.errLog.String()
	var ret error
	if lastErr != "" {
		ret = fmt.Errorf(lastErr)
	}
	return ret
}

func (c *consulServiceWatcher) handler(_ uint64, raw interface{}) {
	// Clear the log because if we reached this point then consul is reachable
	c.errLog.Reset()
	if raw == nil {
		c.errLog.Write([]byte("Watch returned nil data"))
	}

	switch entries := raw.(type) {
	case []*api.ServiceEntry:
		var mysqlServers []admin.MysqlServer

		for _, e := range entries {
			log.Printf("Address: %s, tags: %+v, name: %s", e.Node.Address, e.Service.Tags, e.Service.Service) // todo use other logger
			activeTags := make(map[string]bool)
			for _, tag := range e.Service.Tags {
				activeTags[tag] = true
			}

			if validServiceTags(activeTags, c.def.Tags, c.def.RejectTags) {
				c.def.MysqlServer.Hostname = &e.Node.Address
				mysqlServers = append(mysqlServers, c.def.MysqlServer)
			}
		}
		c.ch <- hostgroupUpdate{mysqlServers: mysqlServers, hostgroupID: c.def.HostgroupID, primary: c.def.Primary}
	default:
	}

}

type consulLogBuffer struct {
	buf bytes.Buffer
}

func (c *consulLogBuffer) Write(p []byte) (int, error) {
	c.buf.Reset()
	return c.buf.Write(p)
}

func (c *consulLogBuffer) String() string {
	return c.buf.String()
}

func (c *consulLogBuffer) Reset() {
	c.buf.Reset()
}

func validServiceTags(activeTags map[string]bool, matchTags, rejectTags []string) bool {
	for _, tag := range matchTags {
		if !activeTags[tag] {
			return false
		}
	}
	for _, tag := range rejectTags {
		if activeTags[tag] {
			return false
		}
	}

	return true
}
