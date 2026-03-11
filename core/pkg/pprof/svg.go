// @Title        svg
// @Description  main
// @Create       yiyiyi 2025/9/11 10:50

package pprof

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PProfSVGGenerator PProf SVG生成器
type PProfSVGGenerator struct {
	pprofPath   string
	outputDir   string
	profileType string
}

// NewPProfSVGGenerator 创建新的SVG生成器
func NewPProfSVGGenerator(pprofPath string) *PProfSVGGenerator {
	return &PProfSVGGenerator{
		pprofPath: pprofPath,
		outputDir: "./pprof_svg",
	}
}

// WithOutputDir 设置输出目录
func (g *PProfSVGGenerator) WithOutputDir(dir string) *PProfSVGGenerator {
	g.outputDir = dir
	return g
}

// WithProfileType 设置profile类型
func (g *PProfSVGGenerator) WithProfileType(profileType string) *PProfSVGGenerator {
	g.profileType = profileType
	return g
}

// GenerateSVG 生成SVG图片
func (g *PProfSVGGenerator) GenerateSVG() (string, error) {
	// 验证文件是否存在
	if err := g.validateFile(); err != nil {
		return "", err
	}

	// 创建输出目录
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成输出文件名
	outputFile := g.generateOutputFilename()

	// 执行go tool pprof命令生成SVG
	cmd := exec.Command("go", "tool", "pprof", "-svg", "-output", outputFile, g.pprofPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("生成SVG失败: %v\n输出: %s", err, output)
	}

	return outputFile, nil
}

// GenerateMultipleFormats 生成多种格式
func (g *PProfSVGGenerator) GenerateMultipleFormats() (map[string]string, error) {
	formats := map[string]string{
		"svg": "-svg",
		"png": "-png",
		"pdf": "-pdf",
		"txt": "-text",
	}

	results := make(map[string]string)

	for format, flag := range formats {
		outputFile := g.generateOutputFilenameWithExt("." + format)

		cmd := exec.Command("go", "tool", "pprof", flag, "-output", outputFile, g.pprofPath)

		if _, err := cmd.CombinedOutput(); err != nil {
			log.Printf("生成%s格式失败: %v", format, err)
			continue
		}

		results[format] = outputFile
	}

	return results, nil
}

// validateFile 验证文件
func (g *PProfSVGGenerator) validateFile() error {
	if _, err := os.Stat(g.pprofPath); os.IsNotExist(err) {
		return fmt.Errorf("pprof文件不存在: %s", g.pprofPath)
	}
	return nil
}

// generateOutputFilename 生成输出文件名
func (g *PProfSVGGenerator) generateOutputFilename() string {
	baseName := strings.TrimSuffix(filepath.Base(g.pprofPath), filepath.Ext(g.pprofPath))
	if g.profileType != "" {
		baseName = g.profileType + "_" + baseName
	}
	return filepath.Join(g.outputDir, baseName+".svg")
}

// generateOutputFilenameWithExt 生成带扩展名的输出文件名
func (g *PProfSVGGenerator) generateOutputFilenameWithExt(ext string) string {
	baseName := strings.TrimSuffix(filepath.Base(g.pprofPath), filepath.Ext(g.pprofPath))
	if g.profileType != "" {
		baseName = g.profileType + "_" + baseName
	}
	return filepath.Join(g.outputDir, baseName+ext)
}

// CleanOutputDir 清理输出目录
func (g *PProfSVGGenerator) CleanOutputDir() error {
	return os.RemoveAll(g.outputDir)
}

// 使用示例
func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run pprof_svg.go <pprof文件路径> [输出目录]")
		fmt.Println("示例: go run pprof_svg.go cpu.pprof ./output")
		return
	}

	pprofFile := os.Args[1]
	outputDir := "./pprof_svg"
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	// 创建生成器实例
	generator := NewPProfSVGGenerator(pprofFile).WithOutputDir(outputDir)

	// 生成SVG
	svgPath, err := generator.GenerateSVG()
	if err != nil {
		log.Fatalf("生成SVG失败: %v", err)
	}

	fmt.Printf("SVG图片已生成: %s\n", svgPath)

	// 可选：生成多种格式
	formats, err := generator.GenerateMultipleFormats()
	if err != nil {
		log.Printf("生成多种格式时出错: %v", err)
	}

	fmt.Println("生成的文件:")
	for format, path := range formats {
		fmt.Printf("  %s: %s\n", format, path)
	}
}
