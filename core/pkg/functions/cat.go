// @Title        functions
// @Description  cat
// @Create       yirl 2025/4/8 15:26

package functions

import (
	"bufio"
	"errors"
	"os"
)

var ErrNoMoreContent = errors.New("no more content to show")

// FileTailViewer 文件尾部查看器结构体
type FileTailViewer struct {
	filePath string
	fileSize int64
	pageSize int
}

// NewFileTailViewer 创建新的文件尾部查看器
func NewFileTailViewer(filePath string, LinesPerPage int) (*FileTailViewer, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, err
	}

	return &FileTailViewer{
		filePath: filePath,
		fileSize: info.Size(),
		pageSize: LinesPerPage,
	}, nil
}

// SetPageSize 设置每页显示的行数
func (f *FileTailViewer) SetPageSize(size int) {
	if size > 0 {
		f.pageSize = size
	}
}

func (f *FileTailViewer) GetFileLines(page int, reverse bool) ([]string, error) {
	if page < 0 {
		return nil, errors.New("invalid page number")
	}

	file, err := os.Open(f.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取总行数
	lineCount, err := countLines(file)
	if err != nil {
		return nil, err
	}

	// 计算需要跳过的行数
	skipLines := page * f.pageSize
	if skipLines >= lineCount {
		return nil, ErrNoMoreContent
	}

	// 重新打开文件读取指定范围
	file, err = os.Open(f.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	if reverse {
		// 倒着读：从文件末尾开始
		startLine := lineCount - skipLines - f.pageSize
		endLine := lineCount - skipLines

		if startLine < 0 {
			startLine = 0
		}
		if endLine > lineCount {
			endLine = lineCount
		}

		scanner := bufio.NewScanner(file)
		currentLine := 0

		for scanner.Scan() {
			if currentLine >= startLine && currentLine < endLine {
				lines = append(lines, scanner.Text())
			}
			currentLine++
			if currentLine >= endLine {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// 倒序结果
		reverseSlice(lines)

	} else {
		// 正着读：从文件开头开始
		startLine := skipLines
		endLine := skipLines + f.pageSize

		if endLine > lineCount {
			endLine = lineCount
		}

		scanner := bufio.NewScanner(file)
		currentLine := 0

		for scanner.Scan() {
			if currentLine >= startLine && currentLine < endLine {
				lines = append(lines, scanner.Text())
			}
			currentLine++
			if currentLine >= endLine {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
		// 正序不需要反转
	}

	return lines, nil
}

func countLines(file *os.File) (int, error) {
	var lineCount int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineCount++
	}

	return lineCount, scanner.Err()
}

func reverseSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
