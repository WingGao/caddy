package proxy

import (
	"net/http"
	"fmt"
	scdb "sitecreater/db"
	"strings"
)

var (
	//key: CodeName, Value: DockerContainerPort
	sitesPortsMap map[string]*siteUpstreamItem
	userHost      string
	scUpstream    *staticUpstream
	dockerHost    string
)

type siteUpstreamItem struct {
	ID       uint32
	Port     uint16
	CodeName string
	Upstream *UpstreamHost
}

type SCPolicy struct{}

func (r *SCPolicy) Select(pool HostPool, request *http.Request) *UpstreamHost {
	host := strings.Split(request.Host, ":")[0]
	host = strings.TrimSuffix(host, userHost)
	dhost, ok := sitesPortsMap[host]
	//fmt.Print(dhost, ok, host)
	if ok {
		return dhost.Upstream
	} else {
		site, err := scdb.GetSiteByCodeName(host)
		if err != nil {
			return nil
		}
		addSCSingleItem(*site, true)
		return sitesPortsMap[host].Upstream
	}
}

func addSCSingleItem(site scdb.Site, upstream bool) {
	s := &siteUpstreamItem{
		ID:       site.ID,
		Port:     site.DockerContainerPort,
		CodeName: site.CodeName,
	}
	if upstream {
		s.Upstream, _ = scUpstream.NewHost(fmt.Sprintf("%s:%d", dockerHost, site.DockerContainerPort))

	}
	sitesPortsMap[site.CodeName] = s
}

func initPools(sqlurl string) {
	scdb.Connect(sqlurl, true)
	sites := scdb.GetAllSitePorts()
	sitesPortsMap = make(map[string]*siteUpstreamItem, 10000)
	for _, site := range sites {
		addSCSingleItem(site, false)
	}
}

func makeSiteCreatorUpstreams(upstreams []Upstream) ([]Upstream, error) {
	for _, site := range sitesPortsMap {
		site.Upstream, _ = scUpstream.NewHost(fmt.Sprintf("%s:%d", dockerHost, site.Port))
	}
	return upstreams, nil
}
