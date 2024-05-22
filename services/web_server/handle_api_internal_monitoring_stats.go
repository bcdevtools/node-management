package web_server

import (
	"github.com/bcdevtools/node-management/utils"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"math"
)

func HandleApiInternalMonitoringStats(c *gin.Context) {
	w := wrapGin(c)

	cfg := w.Config()

	cpuInfo := make(map[string]any)
	vmInfo := make(map[string]any)
	disksInfo := make([]map[string]any, 0)

	pCore, errPCore := cpu.Counts(false)
	if errPCore == nil {
		cpuInfo["physical_cores"] = pCore
	}

	lCore, errLCore := cpu.Counts(true)
	if errLCore == nil {
		cpuInfo["logical_cores"] = lCore
	}

	cpusPercent, errCpuPercent := cpu.Percent(0, false)
	if errCpuPercent == nil {
		cpuInfo["used_percent"] = normalizePercentage(cpusPercent[0])
	}

	vm, errVm := mem.VirtualMemory()
	if errVm == nil {
		vmInfo["total"] = convertByteToGb(vm.Total)
		vmInfo["used"] = convertByteToGb(vm.Used)
		vmInfo["used_percent"] = normalizePercentage(vm.UsedPercent)
	}

	for _, monitorDisk := range cfg.MonitorDisks {
		du, err := disk.Usage(monitorDisk)
		if err != nil {
			utils.PrintlnStdErr("ERR: failed to get disk spec", "disk", monitorDisk, "error", err.Error())
			continue
		}

		disksInfo = append(disksInfo, map[string]any{
			"mount":        monitorDisk,
			"total":        convertByteToGb(du.Total),
			"used":         convertByteToGb(du.Used),
			"used_percent": normalizePercentage(du.UsedPercent),
		})
	}

	w.PrepareDefaultSuccessResponse(map[string]any{
		"cpu":   cpuInfo,
		"ram":   vmInfo,
		"disks": disksInfo,
	}).SendResponse()
}

func convertByteToGb(byteCount uint64) float64 {
	result := float64(byteCount) / 1024 / 1024 / 1024
	result = math.Round(result*100) / 100
	return result
}

func normalizePercentage(percent float64) float64 {
	return float64(int(percent*100)) / 100
}
