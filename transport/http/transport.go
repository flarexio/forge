package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"

	"github.com/flarexio/forge"
)

func ImageDescriptionHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" || namespace == "_" {
			namespace = "library"
		}

		image := c.Param("image")
		if image == "" {
			err := errors.New("image is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		image = namespace + "/" + image

		ctx := c.Request.Context()
		resp, err := endpoint(ctx, image)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		desc, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, desc)
	}
}

func ListImagesHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.Query("page")
		pageSizeStr := c.Query("page_size")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			page = 1
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			pageSize = 20
		}

		req := forge.ListImagesRequest{
			Page:     page,
			PageSize: pageSize,
		}

		ctx := c.Request.Context()
		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func PullImageHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" || namespace == "_" {
			namespace = "library"
		}

		image := c.Param("image")
		if image == "" {
			err := errors.New("image is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		image = namespace + "/" + image

		ctx := c.Request.Context()
		_, err := endpoint(ctx, image)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "ok")
	}
}

func ListContainersHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, nil)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func RunContainerOnceHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		var req forge.RunContainerRequest
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, output)
	}
}

func RunContainerHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		var req forge.RunContainerRequest
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, containerID)
	}
}

func SendToContainerHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		var req forge.SendToContainerRequest
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID := c.Param("container")
		if containerID == "" {
			err := errors.New("container id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req.ContainerID = containerID

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		_, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "ok")
	}
}

func LogsContainerHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID := c.Param("container")
		if containerID == "" {
			err := errors.New("container id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req := forge.LogsContainerRequest{
			ContainerID: containerID,
		}

		if sinceStr := c.Query("since"); sinceStr != "" {
			since, err := time.ParseInLocation(time.RFC3339, sinceStr, time.Local)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				c.Error(err)
				c.Abort()
				return
			}

			req.Since = since
		}

		if tailStr := c.Query("tail"); tailStr != "" {
			tail, err := strconv.Atoi(tailStr)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				c.Error(err)
				c.Abort()
				return
			}

			req.Tail = tail
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		logs, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, logs)
	}
}

func ExecCommandHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		var req forge.ExecCommandRequest
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID := c.Param("container")
		if containerID == "" {
			err := errors.New("container id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req.ContainerID = containerID

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, output)
	}
}

func WaitHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		timeoutStr := c.Query("timeout")
		if timeoutStr == "" {
			err := errors.New("timeout is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req := forge.WaitRequest{
			Timeout: timeout,
		}

		ctx := c.Request.Context()
		endpoint(ctx, req)

		c.String(http.StatusOK, "ok")
	}
}

func SendAndReadHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		var req forge.SendAndReadRequest
		if err := c.ShouldBind(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID := c.Param("container")
		if containerID == "" {
			err := errors.New("container id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req.ContainerID = containerID

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, output)
	}
}

func RemoveContainerHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		containerID := c.Param("container")
		if containerID == "" {
			err := errors.New("container id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		req := forge.RemoveContainerRequest{
			ContainerID: containerID,
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		_, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "ok")
	}
}

func RemoveAllContainersHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspace")
		if workspaceID == "" {
			err := errors.New("workspace id is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, forge.WorkspaceID, workspaceID)

		_, err := endpoint(ctx, nil)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "ok")
	}
}
