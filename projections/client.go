package projections

import (
	"encoding/json"
	"fmt"
	cli "github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type client struct {
	operationTimeout time.Duration
}

func newClient(operationTimeout time.Duration) *client {
	return &client{
		operationTimeout: operationTimeout,
	}
}

func (c *client) Enable(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendPost(toHttpUrl(addr, "/projection/%s/command/enable", name), "", userCredentials, http.StatusOK)
}

func (c *client) Disable(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendPost(toHttpUrl(addr, "/projection/%s/command/disable", name), "", userCredentials, http.StatusOK)
}

func (c *client) Abort(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendPost(toHttpUrl(addr, "/projection/%s/command/abort", name), "", userCredentials, http.StatusOK)
}

func (c *client) CreateOneTime(addr *net.TCPAddr, query string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendPost(toHttpUrl(addr, "/projections/oneTime?type=JS"), query, userCredentials, http.StatusCreated)
}

func (c *client) CreateTransient(
	addr *net.TCPAddr,
	name string,
	query string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendPost(toHttpUrl(addr, "/projections/transient?name=%s&type=JS", name), query, userCredentials,
		http.StatusCreated)
}

func (c *client) CreateContinuous(
	addr *net.TCPAddr,
	name string,
	query string,
	trackEmitted bool,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendPost(
		toHttpUrl(addr, "/projections/continuous?name=%s&type=JS&emit=1&trackemittedstreams=%t", name, trackEmitted),
		query, userCredentials, http.StatusCreated)
}

func (c *client) ListAll(addr *net.TCPAddr, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projections/any"), userCredentials, http.StatusOK).
		ContinueWith(getProjectionDetails)
}

func (c *client) ListOneTime(addr *net.TCPAddr, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projections/onetime"), userCredentials, http.StatusOK).
		ContinueWith(getProjectionDetails)
}

func (c *client) ListContinuous(addr *net.TCPAddr, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projections/continuous"), userCredentials, http.StatusOK).
		ContinueWith(getProjectionDetails)
}

func (c *client) GetStatus(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s", name), userCredentials, http.StatusOK)
}

func (c *client) GetState(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/state", name), userCredentials, http.StatusOK)
}

func (c *client) GetPartitionStateAsync(
	addr *net.TCPAddr,
	name string,
	partition string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/state?partition=%s", name, partition), userCredentials,
		http.StatusOK)
}

func (c *client) GetResult(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/result", name), userCredentials, http.StatusOK)
}

func (c *client) GetPartitionResultAsync(
	addr *net.TCPAddr,
	name string,
	partition string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/result?partition=%s", name, partition), userCredentials,
		http.StatusOK)
}

func (c *client) GetStatistics(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/statistics", name), userCredentials, http.StatusOK)
}

func (c *client) GetQuery(addr *net.TCPAddr, name string, userCredentials *cli.UserCredentials) *tasks.Task {
	return c.sendGet(toHttpUrl(addr, "/projection/%s/query", name), userCredentials, http.StatusOK)
}

func (c *client) UpdateQuery(
	addr *net.TCPAddr,
	name string,
	query string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendPut(toHttpUrl(addr, "/projection/%s/query?type=JS", name), query, userCredentials, http.StatusOK)
}

func (c *client) Delete(
	addr *net.TCPAddr,
	name string,
	deleteEmittedStreams bool,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	return c.sendDelete(toHttpUrl(addr, "/projection/%s?deleteEmittedStreams=%t", name, deleteEmittedStreams),
		userCredentials, http.StatusOK)
}

func (c *client) sendGet(
	reqUrl *url.URL,
	userCredentials *cli.UserCredentials,
	expectedCode int,
) *tasks.Task {
	return tasks.New(func() (interface{}, error) {
		req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
		if err != nil {
			return nil, err
		}

		if userCredentials != nil {
			req.SetBasicAuth(userCredentials.Username(), userCredentials.Password())
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == expectedCode {
			body, _ := ioutil.ReadAll(res.Body)
			return string(body), nil
		} else {
			return nil, fmt.Errorf("projection command failed. server returned %d (%s) for GET on %s",
				res.StatusCode, res.Status, reqUrl.String())
		}
	})
}

func (c *client) sendDelete(
	reqUrl *url.URL,
	userCredentials *cli.UserCredentials,
	expectedCode int,
) *tasks.Task {
	return tasks.New(func() (interface{}, error) {
		req, err := http.NewRequest(http.MethodDelete, reqUrl.String(), nil)
		if err != nil {
			return nil, err
		}

		if userCredentials != nil {
			req.SetBasicAuth(userCredentials.Username(), userCredentials.Password())
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == expectedCode {
			return nil, nil
		} else {
			return nil, fmt.Errorf("projection command failed. server returned %d (%s) for GET on %s",
				res.StatusCode, res.Status, reqUrl.String())
		}
	})
}

func (c *client) sendPost(
	reqUrl *url.URL,
	body string,
	userCredentials *cli.UserCredentials,
	expectedCode int,
) *tasks.Task {
	return tasks.New(func() (interface{}, error) {
		req, err := http.NewRequest(http.MethodPost, reqUrl.String(), strings.NewReader(body))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/json")
		if userCredentials != nil {
			req.SetBasicAuth(userCredentials.Username(), userCredentials.Password())
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == expectedCode {
			return nil, nil
		} else if res.StatusCode == http.StatusConflict {
			return nil, fmt.Errorf("projection command conflict: %s", res.Status)
		} else {
			return nil, fmt.Errorf("projection command failed. server returned %d (%s) for POST on %s",
				res.StatusCode, res.Status, reqUrl.String())
		}
	})
}

func (c *client) sendPut(
	reqUrl *url.URL,
	body string,
	userCredentials *cli.UserCredentials,
	expectedCode int,
) *tasks.Task {
	return tasks.New(func() (interface{}, error) {
		req, err := http.NewRequest(http.MethodPut, reqUrl.String(), strings.NewReader(body))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/json")
		if userCredentials != nil {
			req.SetBasicAuth(userCredentials.Username(), userCredentials.Password())
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == expectedCode {
			return nil, nil
		} else {
			return nil, fmt.Errorf("projection command failed. server returned %d (%s) for PUT on %s",
				res.StatusCode, res.Status, reqUrl.String())
		}
	})
}

func toHttpUrl(addr *net.TCPAddr, pathFormat string, args ...interface{}) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   addr.String(),
		Path:   fmt.Sprintf(pathFormat, args...),
	}
}

func getProjectionDetails(t *tasks.Task) (interface{}, error) {
	if t.IsFaulted() {
		return nil, t.Error()
	}

	body := t.Result().(string)
	data := ListResult{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, err
	}

	return data.Projections, nil
}
