// @Title        文件下载
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package sse

import (
	"context"
	"path"
	"strings"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/download"
	"skeyevss/core/pkg/response"
)

type SSEFileDownloadReq struct {
	Type     string `json:"type" form:"type" path:"type" validate:"required"`
	Url      string `json:"url" form:"url" path:"url" validate:"required"`
	Filename string `json:"filename" form:"filename" path:"filename"`
	Cancel   string `json:"cancel" form:"cancel" path:"cancel"`
}

var (
	_ types.SSEHandleLogic[*FileDownloadLogic, *SSEFileDownloadReq] = (*FileDownloadLogic)(nil)

	FileDownloadType = "file_download"

	VFileDownloadLogic = new(FileDownloadLogic)
)

type FileDownloadLogic struct {
	ctx         context.Context
	svcCtx      *types.ServiceContext
	messageChan chan *types.SSEResponse
}

func (l *FileDownloadLogic) New(ctx context.Context, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) *FileDownloadLogic {
	return &FileDownloadLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		messageChan: messageChan,
	}
}

func (l *FileDownloadLogic) GetType() string {
	return FileDownloadType
}

func (l *FileDownloadLogic) downloader(url, filename string, cancel bool) {
	if cancel {
		l.svcCtx.DownloadManager.CancelDownload(url)
		return
	}

	var (
		extension = strings.Trim(path.Ext(filename), ".")
		savePath  = path.Join(l.svcCtx.Config.SavePath.File, extension)
	)
	if extension == "" {
		extension = "." + constants.EXT_KNOWN
		savePath = path.Join(l.svcCtx.Config.SavePath.File, constants.EXT_KNOWN)
	}

	var task = l.svcCtx.DownloadManager.CreateTask(url, functions.UniqueId()+"."+extension, savePath)
	go l.svcCtx.DownloadManager.StartDownload(l.ctx, task)
}

func (l *FileDownloadLogic) DO(req *SSEFileDownloadReq) {
	var cancel = req.Cancel == "1"
	if !l.svcCtx.DownloadManager.CheckExists(req.Url) || cancel {
		l.downloader(req.Url, req.Filename, cancel)
		if cancel {
			l.messageChan <- &types.SSEResponse{
				Data: download.ProgressUpdate{
					Status:  download.StatusCancelled,
					Message: "下载已取消",
				},
				Done: true,
			}
			return
		}
	}

	defer l.svcCtx.DownloadManager.Unsubscribe(req.Url)
	var (
		ch        = l.svcCtx.DownloadManager.Subscribe(req.Url)
		initState = false
	)
	for {
		select {
		case <-l.ctx.Done():
			return

		case v, ok := <-ch:
			if !ok {
				return
			}

			if !initState || functions.NewTimer().NowMilli()%100 == 0 || v.Progress >= 100 {
				initState = true
				l.messageChan <- &types.SSEResponse{
					Data: v,
				}
			}

			if v.Status == download.StatusCompleted {
				l.messageChan <- &types.SSEResponse{
					Data: v,
					Done: true,
				}
				l.svcCtx.DownloadManager.Finished(req.Url)
				return
			}

			if v.Status == download.StatusCancelled {
				l.messageChan <- &types.SSEResponse{
					Done: true,
				}
				l.svcCtx.DownloadManager.Finished(req.Url)
				return
			}

			if v.Status == download.StatusError {
				l.messageChan <- &types.SSEResponse{
					Err:  response.MakeError(response.NewHttpRespMessage().Str(v.Message), localization.M0010),
					Done: true,
				}
				l.svcCtx.DownloadManager.Finished(req.Url)
				return
			}
		}
	}
}
