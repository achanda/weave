package proxy

import (
	"net/http"
)

type startContainerInterceptor struct{ proxy *Proxy }

func (i *startContainerInterceptor) InterceptRequest(r *http.Request) error {
	container, err := inspectContainerInPath(i.proxy.client, r.URL.Path)
	if err == nil && containerShouldAttach(container) {
		i.proxy.createWait(container.ID)
	}
	return err
}

func (i *startContainerInterceptor) InterceptResponse(r *http.Response) error {
	if r.StatusCode != 201 && r.StatusCode != 204 { // Docker didn't do the start
		return nil
	}
	container, err := inspectContainerInPath(i.proxy.client, r.Request.URL.Path)
	if err == nil {
		i.proxy.waitForStart(container.ID)
	}
	return err
}
