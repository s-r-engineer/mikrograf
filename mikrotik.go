package main

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	lineProtocol "github.com/s-r-engineer/library/lineProtocol"
	librarySync "github.com/s-r-engineer/library/sync"
	"go.uber.org/multierr"
)

const defaultTimeout = time.Second * 10

type Mikrotik struct {
	Address                string
	IgnoreCert             bool
	IgnoreComments         []string
	IncludeModules         []string
	IgnoreDisabled         bool
	Username               string
	Password               string
	Tags                   map[string]string
	SystemTagsURL          []string
	URLS                   []mikrotikEndpoint
	client                 *http.Client
	accumulator            *lineProtocol.Accumulator
	ignoreCommentsFunction func(commonData) bool
}

func newMikrotik(node string, accumulator *lineProtocol.Accumulator) (m Mikrotik, err error) {
	if node == "" {
		return m, fmt.Errorf("mikrotik init -> empty configuration")
	}

	nodeURLObject, err := url.Parse(node)
	if err != nil {
		return m, err
	}

	if nodeURLObject.Host == "" {
		return m, fmt.Errorf("mikrotik init -> wrong configuration: %s", node)
	}

	m.Address = fmt.Sprintf("%s://%s", nodeURLObject.Scheme, nodeURLObject.Host)

	queryParams := nodeURLObject.Query()

	var (
		ignoreCertificate bool = false
		modulesArray      []string
		modulesOk         bool
	)

	if ignoreCertificatesArray, ok := queryParams["ignoreCertificate"]; ok && len(ignoreCertificatesArray) == 1 {
		ignoreCertificate = ignoreCertificatesArray[0] == "true"
	}

	if ignoreComments, ok := queryParams["ignoreComments"]; ok {
		m.IgnoreComments = strings.Split(ignoreComments[0], ",")
	}

	m.IgnoreDisabled = true
	if idis, ok := queryParams["ignoreDisabled"]; ok && idis[0] == "false" {
		m.IgnoreDisabled = false
	}

	modulesArrayRaw, modulesPresent := queryParams["modules"]
	if modulesPresent {
		modulesArray = strings.Split(modulesArrayRaw[0], ",")
		modulesOk = true
	}

	if nodeURLObject.User != nil {
		m.Username = nodeURLObject.User.Username()
		m.Password, _ = nodeURLObject.User.Password()
	}

	mainPropList, systemResourcesPropList, systemRouterBoardPropList := createPropLists()

	if modulesOk {
		if len(modulesArray) == 1 && modulesArray[0] == "all" {
			for selectedModule := range modules {
				m.URLS = append(m.URLS, mikrotikEndpoint{
					name: selectedModule,
					url:  fmt.Sprintf("%s%s?%s", m.Address, modules[selectedModule], mainPropList),
				})
			}
		} else {
			for _, selectedModule := range modulesArray {
				if _, ok := modules[selectedModule]; !ok {
					return m, fmt.Errorf("mikrotik init -> module %s does not exist or has a typo. Correct modules are: %s", selectedModule, getModuleNames())
				}
				m.URLS = append(m.URLS, mikrotikEndpoint{
					name: selectedModule,
					url:  fmt.Sprintf("%s%s?%s", m.Address, modules[selectedModule], mainPropList),
				})
			}
		}
	} else {
		m.URLS = append(m.URLS, mikrotikEndpoint{
			name: "system_resourses",
			url:  fmt.Sprintf("%s%s?%s", m.Address, modules["system_resourses"], mainPropList),
		})
	}

	m.SystemTagsURL = []string{
		m.Address + "/rest/system/resource?" + systemResourcesPropList,
		m.Address + "/rest/system/routerboard?" + systemRouterBoardPropList,
	}

	m.ignoreCommentsFunction = basicCommentAndDisableFilter(m.IgnoreComments, m.IgnoreDisabled)

	m.getClient(ignoreCertificate)
	m.accumulator = accumulator

	return m, m.getSystemTags()
}

func (h *Mikrotik) Run() (errorToReturn error) {
	add, done, wait := librarySync.GetWait()
	lock, unlock := librarySync.GetMutex()
	for _, u := range h.URLS {
		add()
		go func(url mikrotikEndpoint) {
			defer done()
			if err := h.gatherURL(url); err != nil {
				lock()
				defer unlock()
				errorToReturn = multierr.Append(errorToReturn, err)
			}
		}(u)
	}
	wait()
	return errorToReturn
}

func (h *Mikrotik) getClient(ignoreCertificate bool) {
	h.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCertificate},
		},
		Timeout: time.Duration(defaultTimeout),
	}
}

func (h *Mikrotik) getSystemTags() error {
	h.Tags = make(map[string]string)
	for _, tagURL := range h.SystemTagsURL {

		request, err := http.NewRequest("GET", tagURL, nil)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		err = h.setRequestAuth(request)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		binaryData, err := h.queryData(request)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		err = json.Unmarshal(binaryData, &h.Tags)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}
	}
	return nil
}

func (h *Mikrotik) gatherURL(endpoint mikrotikEndpoint) error {
	request, err := http.NewRequest("GET", endpoint.url, nil)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	err = h.setRequestAuth(request)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	binaryData, err := h.queryData(request)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	timestamp := time.Now()

	result, err := binToCommon(binaryData)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	parsedData, err := parse(result, h.ignoreCommentsFunction)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	for _, point := range parsedData {
		point.Tags["source-module"] = endpoint.name
		for n, v := range h.Tags {

			point.Tags[n] = v
		}

		h.accumulator.AddLine("mikrotik", point.Fields, point.Tags, timestamp)
	}
	return nil
}

func (h *Mikrotik) setRequestAuth(request *http.Request) error {
	request.SetBasicAuth(h.Username, h.Password)
	return nil
}

func (h *Mikrotik) queryData(request *http.Request) (data []byte, err error) {
	resp, err := h.client.Do(request)
	if err != nil {
		return data, fmt.Errorf("queryData -> %w", err)
	}

	defer resp.Body.Close()
	defer h.client.CloseIdleConnections()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("queryData -> received status code %d (%s), expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("queryData -> %w", err)
	}

	return data, err
}
