/**
 * @Author:         yi
 * @Description:    system
 * @Version:        1.0.0
 * @Date:           2025/4/21 9:02
 */

package device

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
)

type (
	System struct {
	}

	Mem struct {
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
		Free  uint64 `json:"free"`
	}

	Hardware struct {
		Cpu       float64      `json:"cpu"`
		Mem       *Mem         `json:"mem"`
		Network   *NetworkItem `json:"network"`
		Msg       string       `json:"msg"`
		Timestamp int64        `json:"timestamp"`
	}

	NetworkStats struct {
		InterfaceName string
		RecvRateKB    float64
		SendRateKB    float64
		TotalRecvKB   uint64
		TotalSentKB   uint64
	}

	NetworkItem struct {
		DownRate float64 `json:"downRate"`
		UpRate   float64 `json:"upRate"`
	}

	SevConf struct {
		Port int    `json:"port"`
		Name string `json:"name"`
	}

	Sev struct {
		Cpu       float64 `json:"cpu"`
		Mem       uint64  `json:"mem"`
		Msg       string  `json:"msg"`
		Name      string  `json:"name"`
		Timestamp int64   `json:"timestamp"`
	}
)

func NewSystem() *System {
	return new(System)
}

func (s System) GetNetwork(interval time.Duration) (*NetworkItem, error) {
	stats, err := s.GetNetworkStats(interval)
	if err != nil {
		return nil, err
	}

	var (
		downRate,
		upRate float64
	)
	for _, stat := range stats {
		if stat.RecvRateKB < 0.1 && stat.SendRateKB < 0.1 {
			continue
		}

		downRate += stat.RecvRateKB
		upRate += stat.SendRateKB
	}

	return &NetworkItem{
		DownRate: downRate,
		UpRate:   upRate,
	}, nil
}

func (s System) GetNetworkStats(interval time.Duration) (map[string]NetworkStats, error) {
	prevCounters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial network stats: %v", err)
	}

	prevMap := make(map[string]net.IOCountersStat)
	for _, counter := range prevCounters {
		prevMap[counter.Name] = counter
	}

	time.Sleep(interval)

	// 第二次采样
	currentCounters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get current network stats: %v", err)
	}

	results := make(map[string]NetworkStats)

	for _, current := range currentCounters {
		prev, exists := prevMap[current.Name]
		if !exists {
			continue
		}

		var intervalSec = interval.Seconds()
		results[current.Name] = NetworkStats{
			InterfaceName: current.Name,
			RecvRateKB:    float64(current.BytesRecv-prev.BytesRecv) / 1024 / intervalSec,
			SendRateKB:    float64(current.BytesSent-prev.BytesSent) / 1024 / intervalSec,
			TotalRecvKB:   current.BytesRecv / 1024,
			TotalSentKB:   current.BytesSent / 1024,
		}
	}

	return results, nil
}

// Hardware 硬件使用信息
func (s System) Hardware(interval time.Duration) *Hardware {
	var (
		wg   sync.WaitGroup
		resp = new(Hardware)
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			if err := recover(); err != nil {
				functions.LogError("device system health Hardware panic err:", err, string(debug.Stack()))
			}
		}()

		cpuPercent, err := cpu.Percent(interval, false)
		if err != nil {
			resp = &Hardware{
				Msg: err.Error(),
			}
			return
		}

		memory, err := mem.VirtualMemory()
		if err != nil {
			resp = &Hardware{
				Msg: err.Error(),
			}
			return
		}

		network, err := s.GetNetwork(interval)
		if err != nil {
			resp = &Hardware{
				Msg: err.Error(),
			}
			return
		}

		resp = &Hardware{
			Timestamp: time.Now().UnixMilli(),
			Cpu:       functions.RoundFloat(cpuPercent[0], 2),
			Mem: &Mem{
				Total: memory.Total / 1024 / 1024,
				Used:  memory.Used / 1024 / 1024,
				Free:  memory.Free / 1024 / 1024,
			},
			Network: network,
		}
	}()
	wg.Wait()

	return resp
}

// MemTotal 总内存
func (s System) MemTotal() uint64 {
	memory, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	return memory.Total / 1024 / 1024
}

// Services 服务占用资源信息
func (s System) Services(interval time.Duration, records []*SevConf) []*Sev {
	var (
		wg   sync.WaitGroup
		resp []*Sev
	)
	for _, item := range records {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func(item *SevConf) {
				if err := recover(); err != nil {
					functions.LogError("device system health Services"+item.Name+" panic err:", err, string(debug.Stack()))
				}
			}(item)

			var timestamp = time.Now().UnixMilli()
			if item.Port <= 0 {
				resp = append(resp, &Sev{
					Msg:       "未设置端口",
					Name:      item.Name,
					Timestamp: timestamp,
				})
				return
			}

			var pid = sc.GetPid(item.Port)
			if pid <= 0 {
				resp = append(resp, &Sev{
					Msg:       "服务未启用或已停止",
					Name:      item.Name,
					Timestamp: timestamp,
				})
				return
			}

			// 获取pid
			p, err := process.NewProcess(int32(pid))
			if err != nil {
				resp = append(resp, &Sev{
					Msg:       err.Error(),
					Name:      item.Name,
					Timestamp: timestamp,
				})
				return
			}

			cpuPercent, err := p.Percent(interval)
			if err != nil {
				resp = append(resp, &Sev{
					Msg:       err.Error(),
					Name:      item.Name,
					Timestamp: timestamp,
				})
				return
			}

			memInfo, err := p.MemoryInfo()
			if err != nil {
				resp = append(resp, &Sev{
					Msg:       err.Error(),
					Name:      item.Name,
					Timestamp: timestamp,
				})
				return
			}

			resp = append(resp, &Sev{
				Name:      item.Name,
				Cpu:       cpuPercent,
				Mem:       memInfo.RSS / 1024 / 1024,
				Timestamp: timestamp,
			})
		}()
	}

	wg.Wait()
	return resp
}

type DiskUsageInfo struct {
	Device      string  `json:"device"`       // 设备名称
	Mountpoint  string  `json:"mountpoint"`   // 挂载点
	Fstype      string  `json:"fstype"`       // 文件系统类型
	TotalBytes  uint64  `json:"total_bytes"`  // 总字节数
	UsedBytes   uint64  `json:"used_bytes"`   // 已使用字节数
	FreeBytes   uint64  `json:"free_bytes"`   // 未使用字节数
	UsedPercent float64 `json:"used_percent"` // 使用百分比
}

func (s System) GetDiskUsage() ([]DiskUsageInfo, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %v", err)
	}

	var diskUsageList []DiskUsageInfo

	for _, partition := range partitions {
		if s.shouldSkipPartition(partition) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("Warning: Failed to get usage for %s: %v", partition.Mountpoint, err)
			continue
		}

		diskUsageList = append(diskUsageList, DiskUsageInfo{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			TotalBytes:  usage.Total,
			UsedBytes:   usage.Used,
			FreeBytes:   usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	return functions.ArrUniqueWithCall(
		diskUsageList,
		func(item DiskUsageInfo) string {
			return item.Device
		},
	), nil
}

// shouldSkipPartition 判断是否应该跳过该分区
func (s System) shouldSkipPartition(partition disk.PartitionStat) bool {
	skipFSTypes := []string{
		"devtmpfs", "tmpfs", "squashfs", "overlay", "cgroup", "proc", "sysfs",
		"devpts", "securityfs", "pstore", "efivarfs", "mqueue", "hugetlbfs",
		"debugfs", "configfs", "fusectl", "autofs", "fuse.gvfsd-fuse",
	}

	for _, fsType := range skipFSTypes {
		if partition.Fstype == fsType {
			return true
		}
	}

	skipMountpoints := []string{
		"/dev", "/proc", "/run", "/sys", "/snap", "/var/lib/docker",
	}

	for _, mountpoint := range skipMountpoints {
		if strings.HasPrefix(partition.Mountpoint, mountpoint) {
			return true
		}
	}

	// 跳过只读文件系统（可选）
	if functions.Contains("ro", partition.Opts) {
		return true
	}

	return false
}

// GetDiskUsageByMountpoint 根据挂载点获取特定磁盘的使用情况
func (s System) GetDiskUsageByMountpoint(mountpoint string) (*DiskUsageInfo, error) {
	usage, err := disk.Usage(mountpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage for %s: %v", mountpoint, err)
	}

	// 获取分区信息以补充设备名称
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions: %v", err)
	}

	var device string
	for _, partition := range partitions {
		if partition.Mountpoint == mountpoint {
			device = partition.Device
			break
		}
	}

	return &DiskUsageInfo{
		Device:      device,
		Mountpoint:  mountpoint,
		TotalBytes:  usage.Total,
		UsedBytes:   usage.Used,
		FreeBytes:   usage.Free,
		UsedPercent: usage.UsedPercent,
	}, nil
}

// getDeviceShortName 获取设备短名称
func (s System) getDeviceShortName(device string) string {
	parts := strings.Split(device, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return device
}

// GetTotalDiskUsage 获取总的磁盘使用情况
func (s System) GetTotalDiskUsage() (*DiskUsageInfo, error) {
	allDisks, err := s.GetDiskUsage()
	if err != nil {
		return nil, err
	}

	var totalUsed, totalFree, totalSize uint64

	for _, disk := range allDisks {
		totalUsed += disk.UsedBytes
		totalFree += disk.FreeBytes
		totalSize += disk.TotalBytes
	}

	totalPercent := 0.0
	if totalSize > 0 {
		totalPercent = float64(totalUsed) / float64(totalSize) * 100
	}

	return &DiskUsageInfo{
		Device:      "总计",
		Mountpoint:  "所有磁盘",
		TotalBytes:  totalSize,
		UsedBytes:   totalUsed,
		FreeBytes:   totalFree,
		UsedPercent: totalPercent,
	}, nil
}
