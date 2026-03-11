/**
 * @Author:         yi
 * @Description:    zip
 * @Version:        1.0.0
 * @Date:           2024/5/27 22:03
 */
package functions

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type ZIP struct {
	// 压缩数量
	count int
	// 文件路径
	filepath string
	// 保存路径 分包
	savePath string
	// 包内保存文件名
	innerName string
}

func NewZIP(count int, filepath, savePath string) *ZIP {
	return &ZIP{
		count:     count,
		filepath:  filepath,
		savePath:  savePath,
		innerName: "chunk.data",
	}
}

// ChunkCompress 分包压缩每个包尺寸平均分配大小
func (z *ZIP) ChunkCompress(name string) ([]string, error) {
	// 检测文件是否存在
	if !FileExists(z.filepath) {
		return nil, errors.New("文件不存在")
	}

	if z.savePath != "" {
		if err := MakeDir(z.savePath); err != nil {
			return nil, errors.New("存储路径创建失败")
		}
	}

	inputFile, err := os.Open(z.filepath)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = inputFile.Close()
	}()

	fileInfo, _ := inputFile.Stat()
	var (
		// 文件大小
		filesize = fileInfo.Size()
		// 每个包平均大小
		chunkSize = filesize / int64(z.count)
		// 文件列表
		zipFileList = make([]string, 0, z.count)
	)

	for i := 1; i <= z.count; i++ {
		// 压缩文件名
		var chunkFilename = fmt.Sprintf("%s/%s%d.zip", z.savePath, name, i)
		_ = os.Remove(chunkFilename)
		zipFileList = append(zipFileList, chunkFilename)
		// 创建压缩文件
		zipFile, err := os.Create(chunkFilename)
		if err != nil {
			return nil, err
		}

		// 分块写入
		var zipWriter = zip.NewWriter(zipFile)
		writer, err := zipWriter.Create(z.innerName)
		if err != nil {
			return nil, err
		}

		var bufferSize = chunkSize
		if i == z.count { // 最后一个文件处理剩余部分
			bufferSize = filesize - int64(z.count-1)*chunkSize
		}

		if _, err = io.CopyN(writer, inputFile, bufferSize); err != nil {
			return nil, err
		}

		_ = zipWriter.Close()
		_ = zipFile.Close()
	}

	return zipFileList, nil
}

// CombineDecompression 合并解压
func (z *ZIP) CombineDecompression(name string) (string, []string, error) {
	if z.filepath == "" {
		return "", nil, errors.New("未设置保存文件 filepath")
	}

	var (
		chunkDir = z.savePath
		saveDir  = path.Dir(z.filepath)
	)
	_ = os.Remove(z.filepath)
	_ = os.MkdirAll(saveDir, 0755)

	// 合并后的文件名
	outputFile, err := os.Create(z.filepath)
	if err != nil {
		return "", nil, err
	}

	defer func() {
		_ = outputFile.Close()
	}()

	var chunkFiles []string
	for i := 1; i <= z.count; i++ {
		var chunkFileItem = fmt.Sprintf("%s/%s%d.zip", chunkDir, name, i)
		chunkFiles = append(chunkFiles, chunkFileItem)
		zipFile, err := zip.OpenReader(chunkFileItem)
		if err != nil {
			// 检查是否到达最后一个文件
			if os.IsNotExist(err) {
				break
			}

			return "", nil, err
		}

		for _, file := range zipFile.File {
			reader, err := file.Open()
			if err != nil {
				return "", nil, err
			}

			defer func() {
				_ = reader.Close()
			}()

			_, err = io.Copy(outputFile, reader)
			if err != nil {
				return "", nil, err

			}
		}

		_ = zipFile.Close()
	}

	return z.filepath, chunkFiles, nil
}
