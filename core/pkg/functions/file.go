package functions

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/disintegration/imageorient"
	"github.com/go-basic/uuid"
	"github.com/joho/godotenv"
	xzip "github.com/klauspost/compress/zip"
	"github.com/melbahja/got"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/ulikunitz/xz"

	"skeyevss/core/constants"
	"skeyevss/core/tps"
)

var (
	fileToByteWithCache *bigcache.BigCache
	fileToByteWithCacheExpPool,
	simpleFileBytesPool cmap.ConcurrentMap
)

type noRangeTransport struct {
	base http.RoundTripper
}

func (t *noRangeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 关键：移除 Range 头，告诉服务器我们不支持范围请求
	req.Header.Del("Range")
	req.Header.Del("If-Range")

	// 也可以设置 Accept-Ranges: none
	req.Header.Set("Accept-Ranges", "none")

	return t.base.RoundTrip(req)
}

func init() {
	simpleFileBytesPool = cmap.New()
	fileToByteWithCacheExpPool = cmap.New()
}

type DownloadWithProcessParams struct {
	Url     string
	Dest    string
	Timeout time.Duration
	Output  bool
	Title   string
}

func DownloadWithProcess(params *DownloadWithProcessParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()

	var dl = got.NewWithContext(ctx)
	if params.Output {
		var (
			progress = make(chan string, 1)
			done     = make(chan struct{})
			cancel   = make(chan struct{})
		)

		defer func() {
			cancel <- struct{}{}
		}()

		dl.ProgressFunc = func(d *got.Download) {
			var surplus = d.TotalSize() - d.Size()
			if surplus <= 0 {
				done <- struct{}{}
				return
			}
			progress <- fmt.Sprintf(
				"\r[%s] 已下载: %.2fmb 下载速度: %.2fmb/s, 剩余: %.2fmb, 用时: %s",
				params.Title,
				float64(d.Size())/1024/1024,
				float64(d.AvgSpeed())/1024/1024,
				float64(surplus)/1024/1024,
				FormatDurationPrecise(d.TotalCost(), 2),
			)
		}

		go func() {
			for {
				select {
				case v := <-progress:
					fmt.Printf(v)

				case <-done:
					println(fmt.Sprintf("\n[%s]下载完成", params.Title))
					return

				case <-cancel:
					println(fmt.Sprintf("\n[%s]退出下载", params.Title))
					return
				}
			}
		}()
	}

	return dl.Download(params.Url, params.Dest)
}

type DownloadWithProcessCallItem struct {
	Timeout,
	Done bool

	TotalSize,
	DownloadSize,
	SurplusSize,
	Speed float64

	Dest,
	Url,
	Duration string

	Original *got.Download
}

func DownloadWithProcessCall(params *DownloadWithProcessParams, call func(*DownloadWithProcessCallItem)) (*got.Got, error) {
	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()

	var (
		dl       = got.NewWithContext(ctx)
		progress = make(chan *DownloadWithProcessCallItem, 1)
	)
	dl.ProgressFunc = func(d *got.Download) {
		var surplus = d.TotalSize() - d.Size()
		if surplus <= 0 {
			progress <- &DownloadWithProcessCallItem{
				Url:  params.Title,
				Done: true,
				Dest: params.Dest,
			}
			return
		}

		progress <- &DownloadWithProcessCallItem{
			Dest:         params.Dest,
			Original:     d,
			Url:          params.Title,
			TotalSize:    float64(d.TotalSize()) / 1024 / 1024,
			DownloadSize: float64(d.Size()) / 1024 / 1024,
			SurplusSize:  float64(surplus) / 1024 / 1024,
			Speed:        float64(d.AvgSpeed()) / 1024 / 1024,
			Duration:     FormatDurationPrecise(d.TotalCost(), 2),
		}
	}

	go func() {
		for {
			select {
			case v := <-progress:
				call(v)

				if v.Done {
					return
				}

			case <-ctx.Done():
				progress <- &DownloadWithProcessCallItem{
					Dest:    params.Dest,
					Url:     params.Title,
					Timeout: true,
				}
			}
		}
	}()

	_ = os.RemoveAll(params.Dest)
	_ = os.MkdirAll(path.Dir(params.Dest), 0755)
	// // 开始下载
	// return dl.Do(&got.Download{
	// 	Concurrency: 4,
	// 	URL:         params.Url,
	// 	Dest:        params.Dest,
	// 	Header: []got.GotHeader{
	// 		{Key: "Range", Value: ""},
	// 	},
	// })

	return dl, dl.Download(params.Url, params.Dest)
}

// AbsPath 获取入口路径
func AbsPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}

	return dir
}

// 获取系统临时目录，兼容go run
func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

// 检测路径是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func readAllContent(r io.Reader) ([]byte, error) {
	var b = make([]byte, 4096)
	_, err := r.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type Reader struct {
	io.Reader
	Total   int64
	Current int64
}

// DownloadPdf 下载pdf
func DownloadPdf(savePath, url string) (string, error) {
	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return url, errors.New("目录创建失败")
	}

	var extension = path.Ext(url)
	if extension == "" {
		return "", errors.New("非法类型的pdf")
	}

	var (
		fileName = uuid.New() + "." + constants.EXT_PDF // 文件名
		fullPath = savePath + fileName                  // 本地路径
	)

	if err := DownloadFile(url, fullPath); err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// DownloadImage 下载图片
func DownloadImage(savePath, url string) (string, error) {
	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return url, errors.New("目录创建失败")
	}

	var extension = path.Ext(url)
	if extension == "" {
		extension = "." + constants.EXT_JPG
	}

	var (
		fileName = uuid.New() + extension // 文件名
		fullPath = savePath + fileName    // 本地路径
	)

	fullPath = FilePath(fullPath)
	if err := DownloadFile(url, fullPath); err != nil {
		return url, errors.New("图片下载失败")
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fullPath, err
	}

	return absPath, nil
}

/**
 * @Description: 下载文件
 * @param url
 * @param filename
 * @return error
 */
func DownloadFile(url, filename string) error {
	r, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败 code 40048 %s", err)
	}

	defer func() {
		_ = r.Body.Close()
	}()

	var dirName = path.Dir(filename)
	if !IsDir(dirName) {
		if err := MakeDir(dirName); err != nil {
			return fmt.Errorf("目录创建失败 %s", err)
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建失败 code 40049 %s", err)
	}

	defer func() {
		_ = f.Close()
	}()

	reader := &Reader{
		Reader: r.Body,
		Total:  r.ContentLength,
	}

	_, err = io.Copy(f, reader)
	if err != nil {
		return fmt.Errorf("拷贝失败 code 40050 %s", err)
	}

	return nil
}

/**
 * @Description: 创建目录
 * @param path
 * @return error
 */
func MakeDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else {
		if err := os.MkdirAll(path, 0711); err != nil {
			return err
		}
	}

	// check again
	if _, err := os.Stat(path); err == nil {
		return err
	}

	return nil
}

/**
 * @Description: 检测目录是否存在
 * @param path
 * @return bool
 * @return error
 */
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/**
 * @Description: 生成base64 图片 返回路径
 * @param stream
 * @return string 文件名
 * @return string 全路径
 * @return error error
 */
func MakeBase64Image(savePath, stream string) (string, string, error) {
	flag := "data:image/"

	if index := strings.Index(stream, flag); index != 0 {
		return "", "", errors.New("stream 格式错误 error 1")
	}

	stream = strings.TrimLeft(stream, flag)

	index := strings.Index(stream, ",")
	if index <= 0 {
		return "", "", errors.New("stream 格式错误 error 2")
	}

	// 获取文件名和 base64
	arr := SplitWithIndex(stream, index, index+1)
	extension := strings.TrimRight(arr[0], ";base64")

	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return "", "", errors.New("目录创建失败")
	}

	// 写入临时文件
	fileName := uuid.New() + "." + extension
	fullPath := savePath + fileName
	s, _ := base64.StdEncoding.DecodeString(arr[1])
	err := ioutil.WriteFile(fullPath, s, 0666)

	if err != nil {
		// 删除文件
		_ = os.Remove(fullPath)
		return "", "", errors.New("文件写入失败")
	}

	return fileName, fullPath, nil
}

func SaveBase64File(savePath, flag, stream string) error {
	if flag == "" {
		return errors.New("flat 不能为空 error 0")
	}

	if index := strings.Index(stream, flag); index != 0 {
		return errors.New("stream 格式错误 error 1")
	}

	stream = strings.TrimLeft(stream, flag)
	var index = strings.Index(stream, ",")
	if index <= 0 {
		return errors.New("stream 格式错误 error 2")
	}

	// 获取文件名和 base64
	s, _ := base64.StdEncoding.DecodeString(SplitWithIndex(stream, index, index+1)[1])
	if err := ioutil.WriteFile(savePath, s, 0666); err != nil {
		return err
	}

	return nil
}

// BytesBuffer 写文件
func BytesBuffer(savePath string, data *bytes.Buffer, extension string) (string, string, error) {
	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return "", "", errors.New("目录创建失败")
	}

	// 写入临时文件
	fileName := uuid.New() + "." + extension
	fullPath := savePath + fileName
	err := ioutil.WriteFile(fullPath, data.Bytes(), 0666)

	if err != nil {
		// 删除文件
		_ = os.Remove(fullPath)
		return "", "", errors.New("文件写入失败")
	}

	return fileName, fullPath, nil
}

/**
 * @Description: 获取文件拓展名
 * @param path
 * @return string
 */
func FileExtension(path string) string {
	if path == "" {
		return ""
	}

	arr := strings.Split(path, "")
	var final []string
	var start bool
	for _, val := range arr {
		if val == "." {
			start = true
			final = final[len(final):]
		}

		if start {
			final = append(final, val)
		}
	}

	if len(final) < 2 || final == nil {
		return ""
	}
	return strings.Join(final[1:], "")
}

/**
 * @Description: 文件名称
 * @param filename
 * @return string
 */
func FileRename(filename string) string {
	extension := FileExtension(filename)
	if extension == "" {
		extension = "unknown"
	}

	return UniqueId() + "." + extension
}

/**
 * @Description: 获取URL路径
 * @param data
 * @return string
 */
func FilePath(data string) string {
	pathInfo, err := url.Parse(data)
	if err != nil {
		return ""
	}

	return pathInfo.Path
}

// GetImage 获取图片信息  img.Bounds().Dx()
func GetImage(url string) (image.Image, error) {
	resp, err := http.Get(url)

	defer func(*http.Response) {
		if resp != nil {
			resp.Body.Close()
		}
	}(resp)

	if err != nil {
		return nil, err
	}

	img, _, err := imageorient.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	// img, err := tiff.Decode(resp.Body)
	// if err != nil {
	// 	 return nil, err
	// }

	return img, nil
}

func GetImageTmpPath(savePath, imagePath string) string {
	var localDir = strings.Trim(savePath, ".")
	ePath, err := os.Executable()
	if err != nil {
		return ""
	}

	localDir = path.Dir(ePath) + localDir
	// 创建目录
	if err := MakeDir(localDir); err != nil {
		return ""
	}

	localPath := localDir + imagePath
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return ""
	}
	return absPath
}

// 保存文件
func SaveFile(savePath string, stream []byte, extension string) (string, string, error) {
	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return "", "", errors.New("目录创建失败")
	}

	var (
		fileName = uuid.New() + "." + extension
		fullPath = savePath + fileName
	)
	if err := ioutil.WriteFile(fullPath, stream, 0666); err != nil {
		return "", "", errors.New("文件写入失败")
	}

	return fileName, fullPath, nil
}

// 保存文件
func SaveFileCustom(savePath string, filename string, stream []byte) (string, error) {
	// 创建目录
	if err := MakeDir(savePath); err != nil {
		return "", errors.New("目录创建失败")
	}

	var fullPath = strings.TrimRight(savePath, "/") + "/" + filename
	if err := ioutil.WriteFile(fullPath, stream, 0666); err != nil {
		return "", errors.New("文件写入失败")
	}

	return fullPath, nil
}

// 保存文件
func Save(savePath string, stream []byte) error {
	savePath = strings.TrimRight(savePath, "/")
	arr := SplitWithIndex(savePath, strings.LastIndex(savePath, "/"))
	if arr[0] == "" {
		return errors.New("路径错误")
	}

	// 创建目录
	if err := MakeDir(arr[0]); err != nil {
		return errors.New("目录创建失败")
	}

	// 写入临时文件
	err := ioutil.WriteFile(savePath, stream, 0666)

	if err != nil {
		return errors.New("文件写入失败")
	}

	return nil
}

// 获取文件大小kb
func FileSize(path string) uint64 {
	fi, err := os.Stat(path)

	if err == nil {
		return uint64(math.Ceil(float64(fi.Size() / 1024)))
	}

	return 0
}

// ReadFile 读取文件
func ReadFile(p string) ([]byte, error) {
	return os.ReadFile(p)
	// return ioutil.ReadFile(p)
}

// 读取文件
// func ReadFile(filePath string) ([]byte, error) {
//	f, err := os.Open(filePath)
//	if err != nil {
//		return nil, err
//	}
//
//	defer func(f *os.File) {
//		_ = f.Close()
//	}(f)
//
//	res, err := readAllContent(f)
//	if err != nil {
//		return nil, err
//	}
//
//	return res, nil
// }

func FileToByte(p string) []byte {
	val, ok := simpleFileBytesPool.Get(p)
	if ok {
		return val.([]byte)
	}

	data, err := ReadFile(p)
	if err != nil {
		LogError("read file failed err: ", err)
		return nil
	}

	if !ok || val == nil {
		simpleFileBytesPool.Set(p, data)
	}

	return data
}

func FromFile(r *http.Request, name string) (*tps.FormFile, error) {
	if r.MultipartForm == nil {
		// 表示maxMemory,调用ParseMultipart后，
		// 上传的文件存储在maxMemory大小的内存中，
		// 如果大小超过maxMemory，剩下部分存储在系统的临时文件中
		if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB
			return nil, err
		}
	}
	file, fh, err := r.FormFile(name)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	// 打开文件
	openFile, err := fh.Open()
	if err != nil {
		return nil, err
	}

	var data = make([]byte, fh.Size)
	count, err := openFile.Read(data)
	if err != nil {
		return nil, err
	}

	return &tps.FormFile{
		B:        data[:count],
		FileName: fh.Filename,
		Ext:      strings.ToLower(path.Ext(fh.Filename)),
	}, err
}

func SetImageExt(u string) string {
	if u == "" {
		return ""
	}

	if strings.Index(u, ".") < 0 {
		return u + ".jpg"
	}

	return u
}

func Zip(srcDir string, zipFileName string) error {
	if err := os.RemoveAll(zipFileName); err != nil && !os.IsNotExist(err) {
		return err
	}

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}

	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	srcDir = filepath.Clean(srcDir)
	if !strings.HasSuffix(srcDir, string(filepath.Separator)) {
		srcDir += string(filepath.Separator)
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == srcDir {
			return nil
		}

		// 创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 关键修复：统一使用Linux风格的路径分隔符
		relativePath := strings.TrimPrefix(path, srcDir)
		header.Name = filepath.ToSlash(relativePath) // 将\转换为/

		if info.IsDir() {
			header.Name += "/" // 确保目录以/结尾
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})
}

func UnTar(tarFile string, destDir string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("无法打开 tar 文件: %w", err)
	}
	defer file.Close()

	var tarReader = tar.NewReader(file)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 到达文件尾
		}
		if err != nil {
			return fmt.Errorf("读取条目失败: %w", err)
		}

		// 计算输出路径
		fPath := filepath.Join(destDir, header.Name)

		// 判断是否为目录
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(fPath, os.FileMode(header.Mode))
			continue
		}

		// 创建目标文件
		outFile, err := os.Create(fPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer outFile.Close()

		// 将解压内容写入到目标文件
		if _, err := io.Copy(outFile, tarReader); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}

func UnTarGz(tarGzFile string, destDir string) error {
	file, err := os.Open(tarGzFile)
	if err != nil {
		return fmt.Errorf("无法打开 tar.gz 文件: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建 gzip 读取器失败: %w", err)
	}
	defer gzReader.Close()

	_ = os.RemoveAll(destDir)
	var tarReader = tar.NewReader(gzReader)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 到达文件尾
		}
		if err != nil {
			return fmt.Errorf("读取条目失败: %w", err)
		}

		var fPath = filepath.Join(destDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(fPath, os.FileMode(header.Mode))
			continue
		}

		outFile, err := os.Create(fPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, tarReader); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}

func UnTarXz(tarXzFile string, destDir string) error {
	file, err := os.Open(tarXzFile)
	if err != nil {
		return fmt.Errorf("无法打开 tar.xz 文件: %w", err)
	}
	defer file.Close()

	xzReader, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建 xz 读取器失败: %w", err)
	}

	tarReader := tar.NewReader(xzReader)

	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 遍历 tar 文件中的每个条目
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 到达文件尾
		}
		if err != nil {
			return fmt.Errorf("读取条目失败: %w", err)
		}

		// 计算输出路径
		fPath := filepath.Join(destDir, header.Name)

		// 判断是否为目录
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(fPath, os.FileMode(header.Mode))
			continue
		}

		// 创建目标文件
		outFile, err := os.Create(fPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer outFile.Close()

		// 将解压内容写入到目标文件
		if _, err := io.Copy(outFile, tarReader); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}

func UnZip(zipFile, dest string) error {
	r, err := xzip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, f := range r.File {
		wg.Add(1)
		go func(f *xzip.File) {
			defer wg.Done()

			var filePath = filepath.Join(dest, f.Name)

			if f.FileInfo().IsDir() {
				os.MkdirAll(filePath, os.ModePerm)
				return
			}

			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			rc, err := f.Open()
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
			defer rc.Close()

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
			defer outFile.Close()

			if _, err = io.Copy(outFile, rc); err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(f)
	}

	wg.Wait()
	close(errChan)

	return <-errChan
}

func ArchiveUnZip(filename string, destDir string) error {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}

	defer r.Close()
	os.MkdirAll(destDir, 0755)

	for _, file := range r.File {
		var fPath = filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(fPath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		outFile, err := os.Create(fPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer outFile.Close()

		// 解压文件内容
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开 zip 文件条目失败: %w", err)
		}
		defer rc.Close()

		// 将内容写入到目标文件
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}

func Mv(old, newName string) error {
	return os.Rename(old, newName)
}

func WriteToFile(filePath string, input string) error {
	if !FileExists(filePath) {
		// 确保文件路径存在
		if err := os.MkdirAll(filepath.Dir(filePath), 0o777); err != nil {
			return err
		}
		// f, err := os.Create(filePath)
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o777)
		defer f.Close()
		if err != nil {
			return err
		}

		_, err = f.Write([]byte(input))
		return err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0600)

	defer func() {
		_ = f.Close()
	}()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(input)
	if err != nil {
		return err
	}

	_ = writer.Flush()

	return nil
}

func WriteToFileWithLineEnding(filePath string, input string, lineEnding string) error {
	var normalized string
	switch lineEnding {
	case "crlf":
		normalized = strings.ReplaceAll(input, "\r\n", "\n")      // 先统一为 LF
		normalized = strings.ReplaceAll(normalized, "\n", "\r\n") // 再转换为 CRLF
	case "lf":
		normalized = strings.ReplaceAll(input, "\r\n", "\n")
	default:
		normalized = input // 保持原样
	}

	if !FileExists(filePath) {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(normalized))
	return err
}

func AppendContent(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	var write = bufio.NewWriter(file)
	if _, err := write.WriteString(content); err != nil {
		return err
	}

	_ = write.Flush()
	return nil
}

func GetAllFiles(dirPth string, filter string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			tmp, err := GetAllFiles(dirPth+PthSep+fi.Name(), filter)
			if err != nil {
				return nil, err
			}

			files = append(files, tmp...)
		} else {
			// 过滤指定格式
			if strings.HasSuffix(fi.Name(), ".go") || filter == "" {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table, filter)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

func ReadFileToWords(wordFilepath string) ([]string, error) {
	inputFile, err := os.Open(wordFilepath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	var (
		contentMap = make(map[string]bool)
		scanner    = bufio.NewScanner(inputFile)
	)
	scanner.Buffer(make([]byte, 3*1024*1024), 3*1024*1024)

	for scanner.Scan() {
		var (
			line  = scanner.Text()
			parts = strings.Split(line, "#@@")
		)

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				contentMap[part] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var words []string
	for content := range contentMap {
		words = append(words, content)
	}

	return words, nil
}

// CopyDir 复制文件夹到指定目录
func CopyDir(src string, dst string) error {
	// 获取源文件信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法获取源文件信息: %w", err)
	}

	// 创建目标文件夹
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标文件夹失败: %w", err)
	}

	// 遍历源文件夹
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源文件夹失败: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 如果是目录，递归调用 CopyDir
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 如果是文件，复制文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile 复制单个文件
func CopyFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件 %s: %w", src, err)
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件 %s: %w", dst, err)
	}
	defer output.Close()

	// 复制文件内容
	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 复制文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法获取源文件信息: %w", err)
	}
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("设置目标文件权限失败: %w", err)
	}

	return nil
}

func ConvertPath(path string) string {
	switch runtime.GOOS {
	case "windows":
		return filepath.FromSlash(path)

	case "linux", "darwin":
		return filepath.ToSlash(path)

	default:
		return path
	}
}

func Cat(filePath string, numLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	var (
		scanner   = bufio.NewScanner(file)
		lineCount = 0
		content   = ""
	)

	for scanner.Scan() {
		lineCount++
		if lineCount > numLines {
			break
		}

		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取文件时出错: %w", err)
	}

	return content, nil
}

func LoadEnvFile(envFile string) {
	if err := godotenv.Load(envFile); err != nil {
		panic(".env.local file not found, input [" + envFile + "]")
	}

	PrintStyle("green", "env file:", envFile)
}

func OverloadEnvFile(envFile string) {
	if err := godotenv.Overload(envFile); err != nil {
		panic(".env.local file not found, input [" + envFile + "]")
	}

	PrintStyle("green", "env file:", envFile)
}

type FileTreeNode struct {
	Name     string          `json:"name"`
	IsDir    bool            `json:"isDir"`
	Level    int             `json:"level"`
	Children []*FileTreeNode `json:"children"`
}

func FileTrees(rootPath string, level int) (*FileTreeNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	node := &FileTreeNode{
		Name:  info.Name(),
		IsDir: info.IsDir(),
		Level: level,
	}

	if !info.IsDir() {
		return node, nil
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(rootPath, entry.Name())
		childNode, err := FileTrees(childPath, level+1)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// ReplaceInFiles 递归替换指定目录下所有文件中的内容
// rootDir: 要处理的根目录
// oldStr: 要被替换的字符串
// newStr: 替换后的新字符串
// extensions: 指定要处理的文件扩展名(如 []string{".txt", ".go"})，空切片表示处理所有文件
// skipDirs: 要跳过的目录名
func ReplaceInFiles(rootDir, oldStr, newStr string, extensions []string, skipDirs []string) error {
	// 检查目录是否存在
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", rootDir)
	}

	skipMap := make(map[string]bool)
	for _, dir := range skipDirs {
		skipMap[dir] = true
	}

	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否在跳过列表中
			if skipMap[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名
		if len(extensions) > 0 {
			ext := filepath.Ext(path)
			found := false
			for _, e := range extensions {
				if strings.EqualFold(ext, e) {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件失败: %s, 错误: %v", path, err)
		}

		// 替换内容
		newContent := strings.ReplaceAll(string(content), oldStr, newStr)

		if newContent != string(content) {
			err = ioutil.WriteFile(path, []byte(newContent), info.Mode())
			if err != nil {
				return fmt.Errorf("写入文件失败: %s, 错误: %v", path, err)
			}
		}

		return nil
	})
}

func ReplaceFileLineStream(configFilePath string, patternMaps map[string]string) error {
	if len(patternMaps) <= 0 {
		return nil
	}

	var tmpFile = configFilePath + ".tmp"
	input, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer output.Close()

	var scanner = bufio.NewScanner(input)
	for scanner.Scan() {
		var (
			line  = scanner.Text()
			state = false
		)
		for key, item := range patternMaps {
			if !regexp.MustCompile(key).MatchString(line) {
				continue
			}

			// 替换内容
			if _, err = output.WriteString(item + "\n"); err != nil {
				return err
			}

			state = true
		}

		if !state {
			if _, err = output.WriteString(line + "\n"); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	_ = output.Close()
	_ = input.Close()

	return os.Rename(tmpFile, configFilePath)
}

func GetAllFilesInFolder(folderPath string) ([]string, error) {
	var filePaths []string
	var err = filepath.WalkDir(folderPath, func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			filePaths = append(filePaths, fp)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

func ToUnixPath(ipt string) string {
	if len(ipt) >= 2 && ipt[1] == ':' {
		var drive = strings.ToLower(string(ipt[0]))
		if len(ipt) == 2 {
			ipt = "/" + drive
		}
		ipt = "/" + drive + ipt[2:]
	}

	return filepath.ToSlash(ipt)
}
