// +build !embed_front

package api

import (
	"net/http"

	"github.com/samaritan-proxy/sash/logger"
)

func staticFileHandler() http.Handler {
	logger.Infof("use external mode")
	return http.FileServer(http.Dir("./dist"))
}
