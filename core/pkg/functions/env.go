// @Title        env
// @Description  main
// @Create       yiyiyi 2025/11/11 17:40

package functions

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 读取环境变量文件并解析为map
func parseEnvFile(filename string) (map[string]string, error) {
	envVars := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("警告: 第 %d 行格式不正确: %s\n", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			fmt.Printf("警告: 第 %d 行键名为空\n", lineNum)
			continue
		}

		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件 %s 时出错: %v", filename, err)
	}

	return envVars, nil
}

// 替换环境变量值
func ReplaceEnvValues(sourceFile, targetFile string, keepVariables []string) error {
	// 读取源文件
	oldVars, err := parseEnvFile(sourceFile)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %v", err)
	}

	// 读取目标文件
	targetVars, err := parseEnvFile(targetFile)
	if err != nil {
		return fmt.Errorf("读取目标文件失败: %v", err)
	}

	// 替换值
	for key, oldValue := range oldVars {
		if _, exists := targetVars[key]; exists {
			if Contains(key, keepVariables) {
				continue
			}
			targetVars[key] = oldValue
		}
	}

	// 重新写入目标文件
	return writeEnvFile(targetFile, targetVars)
}

// 将环境变量map写入文件
func writeEnvFile(filename string, envVars map[string]string) error {
	// 读取原始文件内容以保留注释和格式
	originalContent, err := readFileWithStructure(filename)
	if err != nil {
		return fmt.Errorf("读取原始文件结构失败: %v", err)
	}

	// 创建新文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %v", filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入内容，保留原始结构但替换值
	for _, line := range originalContent {
		trimmedLine := strings.TrimSpace(line)

		// 如果是空行或注释，直接写入
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			fmt.Fprintln(writer, line)
			continue
		}

		// 如果是环境变量行，替换值
		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			if newValue, exists := envVars[key]; exists {
				// 替换值，保留原始行的格式（比如空格）
				prefix := strings.Split(line, "=")[0]
				fmt.Fprintf(writer, "%s=%s\n", prefix, newValue)
			} else {
				fmt.Fprintln(writer, line)
			}
		} else {
			fmt.Fprintln(writer, line)
		}
	}

	return writer.Flush()
}

// 读取文件并保留原始结构（包括空格、注释等）
func readFileWithStructure(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
