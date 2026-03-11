// @Title        main
// @Description  main
// @Create       yiyiyi 2025/9/5 09:03

package sc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

func SyscallScriptCrossPlatformDetach(scriptPath string) (int, error) {
	if _, err := os.Stat(scriptPath); err != nil {
		return 0, err
	}

	if runtime.GOOS == "windows" {
		var cmd *exec.Cmd
		cmd = exec.Command("cmd.exe", "/C", "start", "/B", scriptPath)
		cmd.SysProcAttr = sysProcAttr()
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Dir = filepath.Dir(scriptPath)
		if err := cmd.Start(); err != nil {
			return 0, fmt.Errorf("启动失败: %v\nCommand: %s", err, cmd.String())
		}

		var pid = cmd.Process.Pid
		_ = cmd.Process.Release()
		return pid, nil

	}

	{
		var (
			cmd            = exec.Command("chmod", "+x", scriptPath)
			stdout, stderr bytes.Buffer
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Dir = filepath.Dir(scriptPath)
		if err := cmd.Run(); err != nil {
			return 0, fmt.Errorf("权限设置失败: %v\nCommand: %s", err, cmd.String())
		}
	}

	var cmd *exec.Cmd
	cmd = exec.Command("nohup", "/bin/sh", scriptPath, "&")
	// cmd.SysProcAttr = sysProcAttr()
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Dir = filepath.Dir(scriptPath)
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("启动失败: %v\nCommand: %s", err, cmd.String())
	}

	var pid = cmd.Process.Pid
	return pid, nil
}

func SyscallBinCrossPlatformDetach(binPath string, args ...string) (int, error) {
	if _, err := os.Stat(binPath); err != nil {
		return 0, err
	}

	if len(args) <= 0 {
		return 0, errors.New("参数不能为空")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(binPath, args...)
		// cmd.SysProcAttr = sysProcAttr()
	} else {
		var data = append([]string{binPath}, args...)
		data = append(data, "&")
		cmd = exec.Command("nohup", data...)
	}

	println("exec command: ", cmd.String())
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// var stdout, stderr bytes.Buffer
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// println("stdout", stdout.String(), "\n", stderr.String())
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("启动失败: %v", err)
	}

	var pid = cmd.Process.Pid
	_ = cmd.Process.Release()
	return pid, nil
}

func ExecCommand(cmdString string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", cmdString)
	} else {
		return exec.Command("sh", "-c", cmdString)
	}
}

func CheckCommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func GetPid(port int) int {
	if runtime.GOOS == "windows" {
		return getPidWindows(port)
	} else {
		return getPidUnix(port)
	}
}

func getPidWindows(port int) int {
	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return -1
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, err := strconv.Atoi(fields[len(fields)-1])
				if err == nil {
					return pid
				}
			}
		}
	}
	return -1
}

func getPidUnix(port int) int {
	// 使用lsof直接获取PID
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			return pid
		}
	}

	//  如果lsof失败，使用netstat
	cmd = exec.Command("netstat", "-tlnp")
	output, err = cmd.Output()
	if err != nil {
		return -1
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
			// 查找PID部分，通常是最后一个字段的/前面
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				lastField := fields[len(fields)-1]
				if idx := strings.Index(lastField, "/"); idx != -1 {
					pidStr := lastField[:idx]
					if pid, err := strconv.Atoi(pidStr); err == nil {
						return pid
					}
				}
			}
		}
	}
	return -1
}

// 开启进程
func StartProcess(pathExe string, arg ...string) (int, error) {
	var cmd = exec.Command(pathExe, arg...)
	cmd.Dir = filepath.Dir(pathExe)
	cmd.Stdout = os.Stdout
	cmdReader, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}

	var scanner = bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println("server output", scanner.Text())
		}
	}()

	if err = cmd.Start(); err != nil {
		return 0, err
	}

	if err = cmd.Wait(); err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

// 停止进程
func StopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := proc.Kill(); err != nil {
		return err
	}
	return nil
}

func KillProcess(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("找不到进程: %+v; err: %v", pid, err)
	}

	if runtime.GOOS != "windows" {
		var signal = syscall.SIGKILL
		err = p.Signal(signal)
		if err != nil {
			return fmt.Errorf("无法杀死进程: %+v; err: %v", pid, err)
		}
		return nil
	}

	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
}

func GetPidByName(name string) (int, error) {
	processes, err := process.Processes()
	if err != nil {
		return -1, err
	}

	for _, p := range processes {
		pName, err := p.Name()
		if err != nil {
			continue
		}

		if pName == name {
			return int(p.Pid), nil
		}
	}

	return -1, nil
}

func getPIDsUsingDirLinux(dirPath string) ([]string, error) {
	// 方法1: 使用 lsof
	cmd := exec.Command("lsof", "+D", dirPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lsof failed: %v", err)
	}

	// 解析输出，提取 PID
	var pIds []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // 跳过标题行
		fields := strings.Fields(line)
		if len(fields) > 1 {
			pIds = append(pIds, fields[1]) // 第二列是 PID
		}
	}
	return pIds, nil
}

func getPIDsUsingDirWindows(dirPath string) ([]string, error) {
	cmd := exec.Command("handle.exe", "-nobanner", dirPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("handle.exe failed: %v", err)
	}

	// 解析输出（示例：explorer.exe pid: 1234）
	re := regexp.MustCompile(`pid: (\d+)`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	var pIds []string
	for _, match := range matches {
		pIds = append(pIds, match[1])
	}
	return pIds, nil
}

func UnlockerDir(dirPath string) error {
	var pIds []string
	var err error

	switch runtime.GOOS {
	case "linux", "darwin":
		pIds, err = getPIDsUsingDirLinux(dirPath)
	case "windows":
		pIds, err = getPIDsUsingDirWindows(dirPath)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	if err != nil {
		return err
	}

	if len(pIds) == 0 {
		fmt.Println("No processes using the directory.")
		return nil
	}

	// 结束进程
	for _, item := range pIds {
		pid, err := strconv.Atoi(item)
		if err != nil {
			return err
		}

		if err = KillProcess(pid); err != nil {
			return err
		}
	}

	return nil
}

func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}

	defer func() {
		_ = listener.Close()
	}()
	return true
}

func IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil

	case "linux", "darwin", "freebsd", "openbsd", "netbsd":
		return os.Geteuid() == 0

	default:
		return false
	}
}

func ExecPermission(binPath string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("无法设置执行权限: %v", err)
	}

	if !HasExecutePermission(binPath) {
		return fmt.Errorf("文件没有执行权限")
	}

	return nil
}

// 检查文件是否有执行权限
func HasExecutePermission(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

func CheckPort(port int, protocol int) (bool, string) {
	var (
		cmd      *exec.Cmd
		outBytes bytes.Buffer
	)
	switch runtime.GOOS {
	case "windows":
		if protocol == 0 {
			cmd = exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano -p %s | findstr :%d", "tcp", port))
		} else {
			cmd = exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano -p %s | findstr :%d", "udp", port))
		}
	case "darwin":
		if protocol == 0 {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("netstat -anp tcp | grep %d", port))
		} else {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("netstat -anp udp | grep %d", port))
		}
	default:
		if protocol == 0 {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("netstat -tanp | grep :%d", port))
		} else {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("netstat -uanp | grep :%d", port))
		}
	}
	cmd.Stdout = &outBytes
	if err := cmd.Run(); err != nil {
		return false, err.Error()
	}

	return parseNetstatOutput(outBytes.String(), port, protocol)
}

func parseNetstatOutput(output string, port int, protocol int) (bool, string) {
	var (
		re        *regexp.Regexp
		strOutput string
	)
	if runtime.GOOS == "windows" {
		if protocol == 0 {
			strOutput = fmt.Sprintf(`TCP\s+.+?:%d\s+.+?:(\d+\|\*)s+\b(ESTABLISHED|LISTENING)\b\s.*?(\d+)\s`, port)
		} else {
			strOutput = fmt.Sprintf(`UDP\s+.+?:(%d)\s+.+?:(\d+|\*)\s+.*?(\d+)\s`, port)
		}

	} else if runtime.GOOS == "darwin" {
		if protocol == 0 {
			strOutput = fmt.Sprintf(`.*\.%d\s.*?\.(\d+|\*)\s+\b(ESTABLISHED|LISTEN)\b\s.*?`, port)
		} else {
			strOutput = fmt.Sprintf(`.*\.%d\s.*?\.(\d+|\*)\s+\b\b\s.*?`, port)
		}
	} else {
		if protocol == 0 {
			strOutput = fmt.Sprintf(`:%d\s.*?:(\d+|\*)\s+\b(ESTABLISHED|LISTEN)\b\s.*?(\d+)/\S+`, port)
			// ESTABLISHED
		} else {
			strOutput = fmt.Sprintf(`:(%d)\s.*?:(\d+|\*)\s.*?(\d+)/\S+`, port)
		}
	}

	re = regexp.MustCompile(strOutput)
	var matches = re.FindStringSubmatch(output)
	if runtime.GOOS == "darwin" {
		if len(matches) < 3 {
			return false, "-1"
		}

		return true, "unknown"
	}
	if len(matches) < 4 {
		return false, "-1"
	}

	return true, matches[3]
}
