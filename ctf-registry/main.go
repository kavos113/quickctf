package main

import (
	"log"
	"net/http"

	"github.com/kavos113/quickctf/ctf-registry/handler"
	"github.com/kavos113/quickctf/ctf-registry/storage/filesystem"
	"github.com/kavos113/quickctf/ctf-registry/store/boltstore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogMethod: true,
		LogURI:    true,
		LogError:  true,
		LogHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Content-Range",
			"Docker-Content-Digest",
			"Location",
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("Status: %d, Method: %s, URI: %s, Headers: %+v, Error: %+v", v.Status, v.Method, v.URI, v.Headers, v.Error)
			return nil
		},
	}))

	fs := filesystem.NewStorage()
	ss := boltstore.NewStore()

	bh := handler.NewBlobHandler(fs)
	buh := handler.NewBlobUploadHandler(fs)
	mh := handler.NewManifestHandler(fs, ss)
	th := handler.NewTagHandler(ss)

	e.GET("/v2/", baseHandler)
	e.GET("/v2/:name/blobs/:digest", bh.GetBlobs)
	e.HEAD("/v2/:name/blobs/:digest", bh.GetBlobs)
	e.DELETE("/v2/:name/blobs/:digest", bh.DeleteBlob)
	e.POST("/v2/:name/blobs/uploads/", buh.PostBlobUploads)
	e.GET("/v2/:name/blobs/uploads/:reference", buh.GetBlobUploads)
	e.PUT("/v2/:name/blobs/uploads/:reference", buh.PutBlobUpload)
	e.PATCH("/v2/:name/blobs/uploads/:reference", buh.PatchBlobUpload)

	e.PUT("/v2/:name/manifests/:reference", mh.PutManifests)
	e.GET("/v2/:name/manifests/:reference", mh.GetManifests)
	e.HEAD("/v2/:name/manifests/:reference", mh.GetManifests)
	e.DELETE("/v2/:name/manifests/:digest", mh.DeleteManifests)
	e.GET("/v2/:name/referrers/:digest", mh.GetReferrers)

	e.GET("/v2/:name/tags/list", th.GetTags)

	e.Logger.Fatal(e.Start(":8080"))
}

func baseHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSON(http.StatusOK, map[string]string{})
}
