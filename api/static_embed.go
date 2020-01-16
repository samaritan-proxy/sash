// +build embed_front

package api

import (
	"net/http"

	"github.com/rakyll/statik/fs"

	_ "github.com/samaritan-proxy/sash/api/statik"
	"github.com/samaritan-proxy/sash/logger"
)

func staticFileHandler() http.Handler {
	logger.Infof("Embeded the static files of front")
	statikFS, err := fs.New()
	if err != nil {
		logger.Fatal(err)
	}
	return http.FileServer(statikFS)
}
