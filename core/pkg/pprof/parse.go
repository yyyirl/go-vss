// @Title        parse
// @Description  main
// @Create       yiyiyi 2025/9/11 10:44

package pprof

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PProfAnalyzer PProf文件分析器
type PProfAnalyzer struct {
	filePath    string
	profileType string
	topResults  []TopResult
	summary     map[string]interface{}
	rawOutput   string
}

// TopResult 分析结果结构
type TopResult struct {
	Flat        float64
	FlatPercent float64
	Cum         float64
	CumPercent  float64
	Function    string
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	FileName    string                 `json:"file_name"`
	ProfileType string                 `json:"profile_type"`
	TopResults  []TopResult            `json:"top_results"`
	Summary     map[string]interface{} `json:"summary"`
	RawData     string                 `json:"raw_data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewPProfAnalyzer 创建新的PProf分析器实例
func NewPProfAnalyzer(filePath string) *PProfAnalyzer {
	return &PProfAnalyzer{
		filePath:    filePath,
		profileType: "auto",
		summary:     make(map[string]interface{}),
	}
}

// WithProfileType 设置profile类型
func (a *PProfAnalyzer) WithProfileType(profileType string) *PProfAnalyzer {
	a.profileType = profileType
	return a
}

// Analyze 执行分析
func (a *PProfAnalyzer) Analyze() (*AnalysisResult, error) {
	if err := a.validateFile(); err != nil {
		return nil, err
	}

	if err := a.detectProfileType(); err != nil {
		return nil, err
	}

	if err := a.runGoToolAnalysis(); err != nil {
		return nil, err
	}

	if err := a.parseTopOutput(); err != nil {
		return nil, err
	}

	if err := a.collectSummary(); err != nil {
		return nil, err
	}

	return a.buildResult(), nil
}

// validateFile 验证文件是否存在
func (a *PProfAnalyzer) validateFile() error {
	if _, err := os.Stat(a.filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", a.filePath)
	}
	return nil
}

// detectProfileType 检测profile类型
func (a *PProfAnalyzer) detectProfileType() error {
	if a.profileType != "auto" {
		return nil
	}

	filename := strings.ToLower(filepath.Base(a.filePath))
	switch {
	case strings.Contains(filename, "cpu"):
		a.profileType = "cpu"
	case strings.Contains(filename, "heap"):
		a.profileType = "heap"
	case strings.Contains(filename, "goroutine"):
		a.profileType = "goroutine"
	case strings.Contains(filename, "block"):
		a.profileType = "block"
	case strings.Contains(filename, "mutex"):
		a.profileType = "mutex"
	default:
		a.profileType = "cpu"
	}

	return nil
}

// runGoToolAnalysis 运行go tool pprof分析
func (a *PProfAnalyzer) runGoToolAnalysis() error {
	cmd := exec.Command("go", "tool", "pprof", "-top", a.filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go tool pprof执行失败: %v\n输出: %s", err, output)
	}

	a.rawOutput = string(output)
	return nil
}

// parseTopOutput 解析top输出
func (a *PProfAnalyzer) parseTopOutput() error {
	var results []TopResult
	lines := strings.Split(a.rawOutput, "\n")

	re := regexp.MustCompile(`^\s*(\d+\.\d+)([a-zA-Z]*)\s+(\d+\.\d+)%\s+(\d+\.\d+)%\s+(.+)`)

	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); matches != nil {
			flat, _ := strconv.ParseFloat(matches[1], 64)
			flatPercent, _ := strconv.ParseFloat(matches[3], 64)
			cumPercent, _ := strconv.ParseFloat(matches[4], 64)

			results = append(results, TopResult{
				Flat:        flat,
				FlatPercent: flatPercent,
				Cum:         flat,
				CumPercent:  cumPercent,
				Function:    strings.TrimSpace(matches[5]),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FlatPercent > results[j].FlatPercent
	})

	a.topResults = results
	return nil
}

// collectSummary 收集摘要信息
func (a *PProfAnalyzer) collectSummary() error {
	fileInfo, err := os.Stat(a.filePath)
	if err != nil {
		return err
	}

	a.summary["file_size"] = fileInfo.Size()
	a.summary["file_mod_time"] = fileInfo.ModTime()
	a.summary["profile_type"] = a.profileType

	// 根据不同类型收集特定信息
	switch a.profileType {
	case "heap":
		cmd := exec.Command("go", "tool", "pprof", "-alloc_space", "-text", a.filePath)
		output, _ := cmd.CombinedOutput()
		a.summary["heap_analysis"] = string(output)

	case "goroutine":
		goroutineCount := a.countGoroutines()
		a.summary["goroutine_count"] = goroutineCount

	case "cpu":
		sampleCount := a.countSamples()
		a.summary["sample_count"] = sampleCount
	}

	return nil
}

// countGoroutines 统计goroutine数量
func (a *PProfAnalyzer) countGoroutines() int {
	re := regexp.MustCompile(`goroutine profile: total (\d+)`)
	matches := re.FindStringSubmatch(a.rawOutput)
	if len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		return count
	}
	return -1
}

// countSamples 统计样本数量
func (a *PProfAnalyzer) countSamples() int {
	re := regexp.MustCompile(`Total: (\d+) samples`)
	matches := re.FindStringSubmatch(a.rawOutput)
	if len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		return count
	}
	return -1
}

// buildResult 构建分析结果
func (a *PProfAnalyzer) buildResult() *AnalysisResult {
	return &AnalysisResult{
		FileName:    a.filePath,
		ProfileType: a.profileType,
		TopResults:  a.topResults,
		Summary:     a.summary,
		RawData:     a.rawOutput,
		Timestamp:   time.Now(),
	}
}

// PrintReport 打印分析报告
func (a *PProfAnalyzer) PrintReport(result *AnalysisResult) {
	fmt.Printf("=== PProf 分析报告 ===\n")
	fmt.Printf("文件: %s\n", result.FileName)
	fmt.Printf("类型: %s\n", result.ProfileType)
	fmt.Printf("时间: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("\n")

	fmt.Printf("Top 10 结果:\n")
	fmt.Printf("%-8s %-8s %-8s %s\n", "Flat%", "Cum%", "Flat", "Function")
	fmt.Printf("-------- -------- -------- ----------\n")

	for i, result := range result.TopResults {
		if i >= 10 {
			break
		}
		fmt.Printf("%-8.1f %-8.1f %-8.1f %s\n",
			result.FlatPercent, result.CumPercent, result.Flat, result.Function)
	}

	fmt.Printf("\n摘要信息:\n")
	for key, value := range result.Summary {
		fmt.Printf("%-15s: %v\n", key, value)
	}

	a.printSuggestions(result)
}

// printSuggestions 打印优化建议
func (a *PProfAnalyzer) printSuggestions(result *AnalysisResult) {
	fmt.Printf("\n优化建议:\n")

	switch result.ProfileType {
	case "cpu":
		if len(result.TopResults) > 0 && result.TopResults[0].FlatPercent > 30 {
			fmt.Printf("- 🔥 函数 '%s' 占用了 %.1f%% 的CPU时间，建议优化\n",
				result.TopResults[0].Function, result.TopResults[0].FlatPercent)
		}

	case "heap":
		fmt.Printf("- 💾 使用 'go tool pprof -alloc_objects %s' 查看对象分配详情\n", result.FileName)
		fmt.Printf("- 💾 使用 'go tool pprof -inuse_objects %s' 查看存活对象\n", result.FileName)

	case "goroutine":
		if count, ok := result.Summary["goroutine_count"].(int); ok && count > 1000 {
			fmt.Printf("- 🚨 检测到大量goroutine (%d)，可能存在goroutine泄漏\n", count)
		}
	}

	fmt.Printf("- 📊 详细分析: go tool pprof %s\n", result.FileName)
	fmt.Printf("- 📈 生成SVG图: go tool pprof -web %s\n", result.FileName)
}

// SaveReport 保存分析报告到文件
func (a *PProfAnalyzer) SaveReport(result *AnalysisResult, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	fmt.Fprintf(writer, "=== PProf Analysis Report ===\n")
	fmt.Fprintf(writer, "File: %s\n", result.FileName)
	fmt.Fprintf(writer, "Type: %s\n", result.ProfileType)
	fmt.Fprintf(writer, "Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "Top 10 Results:\n")
	fmt.Fprintf(writer, "%-8s %-8s %-8s %s\n", "Flat%", "Cum%", "Flat", "Function")
	fmt.Fprintf(writer, "-------- -------- -------- ----------\n")

	for i, res := range result.TopResults {
		if i >= 10 {
			break
		}
		fmt.Fprintf(writer, "%-8.1f %-8.1f %-8.1f %s\n",
			res.FlatPercent, res.CumPercent, res.Flat, res.Function)
	}

	return writer.Flush()
}

// func main() {
// 	// 创建生成器实例
// 	generator := pprof.NewPProfSVGGenerator("/Users/yiyiyi/Downloads/BACKEND-API-goroutine-20250911-103518.pprof").WithOutputDir("/Users/yiyiyi/Downloads")
//
// 	// 生成SVG
// 	svgPath, err := generator.GenerateSVG()
// 	if err != nil {
// 		log.Fatalf("生成SVG失败: %v", err)
// 	}
//
// 	fmt.Printf("SVG图片已生成: %s\n", svgPath)
//
// 	// 可选：生成多种格式
// 	formats, err := generator.GenerateMultipleFormats()
// 	if err != nil {
// 		log.Printf("生成多种格式时出错: %v", err)
// 	}
//
// 	fmt.Println("生成的文件:")
// 	for format, path := range formats {
// 		fmt.Printf("  %s: %s\n", format, path)
// 	}
// }
//
// func mainp() {
// 	var filePath = "/Users/yiyiyi/Downloads/BACKEND-API-goroutine-20250911-103518.pprof"
// 	analyzer := pprof.NewPProfAnalyzer(filePath)
//
// 	// 执行分析
// 	result, err := analyzer.Analyze()
// 	if err != nil {
// 		log.Fatalf("分析失败: %v", err)
// 	}
//
// 	// 打印报告
// 	analyzer.PrintReport(result)
//
// 	// 保存报告到文件
// 	reportFile := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "_report.txt"
// 	if err := analyzer.SaveReport(result, reportFile); err != nil {
// 		log.Printf("保存报告失败: %v", err)
// 	} else {
// 		fmt.Printf("\n报告已保存到: %s\n", reportFile)
// 	}
// }
