//go:build darwin && cgo

package stats

/*
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <unistd.h>

static int get_vm_info64(vm_statistics64_data_t *vmstat, uint64_t *page_size) {
    mach_msg_type_number_t count = HOST_VM_INFO64_COUNT;
    kern_return_t kr = host_statistics64(mach_host_self(), HOST_VM_INFO64, (host_info64_t)vmstat, &count);
    if (kr != KERN_SUCCESS) {
        return kr;
    }
    *page_size = (uint64_t)sysconf(_SC_PAGESIZE);
    return 0;
}

static int get_cpu_load(natural_t ticks[CPU_STATE_MAX]) {
    mach_msg_type_number_t count = HOST_CPU_LOAD_INFO_COUNT;
    host_cpu_load_info_data_t info;
    kern_return_t kr = host_statistics(mach_host_self(), HOST_CPU_LOAD_INFO, (host_info_t)&info, &count);
    if (kr != KERN_SUCCESS) {
        return kr;
    }
    ticks[CPU_STATE_USER] = info.cpu_ticks[CPU_STATE_USER];
    ticks[CPU_STATE_SYSTEM] = info.cpu_ticks[CPU_STATE_SYSTEM];
    ticks[CPU_STATE_IDLE] = info.cpu_ticks[CPU_STATE_IDLE];
    ticks[CPU_STATE_NICE] = info.cpu_ticks[CPU_STATE_NICE];
    return 0;
}

static int get_processor_loads(natural_t **ticks, natural_t *num_cpus, mach_msg_type_number_t *num_ticks) {
    processor_info_array_t info_array;
    mach_msg_type_number_t info_count;
    natural_t processor_count;
    kern_return_t kr = host_processor_info(mach_host_self(), PROCESSOR_CPU_LOAD_INFO, &processor_count, &info_array, &info_count);
    if (kr != KERN_SUCCESS) {
        return kr;
    }
    *ticks = (natural_t*)info_array;
    *num_cpus = processor_count;
    *num_ticks = info_count;
    return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/sysmon/system-monitor-cli/internal/models"
	"golang.org/x/sys/unix"
)

func getDarwinMemoryStats() (*models.MemoryStats, error) {
	memSize, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory size: %w", err)
	}

	var vmstat C.vm_statistics64_data_t
	var pageSize C.uint64_t
	if res := C.get_vm_info64(&vmstat, &pageSize); res != 0 {
		return nil, fmt.Errorf("host_statistics64 failed with code %d", res)
	}

	ps := uint64(pageSize)
	if ps == 0 {
		ps = uint64(unix.Getpagesize())
	}

	// Used memory includes active, wired down, and compressed memory pages
	active := uint64(vmstat.active_count) * ps
	wired := uint64(vmstat.wire_count) * ps
	compressed := uint64(vmstat.compressor_page_count) * ps
	used := active + wired + compressed

	if used > memSize {
		used = memSize
	}
	available := memSize - used

	return &models.MemoryStats{
		Total:     memSize,
		Used:      used,
		Available: available,
		Percent:   models.CalculatePercentage(used, memSize),
	}, nil
}

func getDarwinCPUStats(prevOverall *cpuTime, prevPerCore []cpuTime) (*models.CPUStats, *cpuTime, []cpuTime, error) {
	var ticks [C.CPU_STATE_MAX]C.natural_t
	if res := C.get_cpu_load(&ticks[0]); res != 0 {
		return nil, nil, nil, fmt.Errorf("host_statistics CPU load failed with code %d", res)
	}

	currOverall := &cpuTime{
		user:   uint64(ticks[C.CPU_STATE_USER]),
		system: uint64(ticks[C.CPU_STATE_SYSTEM]),
		idle:   uint64(ticks[C.CPU_STATE_IDLE]),
		nice:   uint64(ticks[C.CPU_STATE_NICE]),
	}

	var stats models.CPUStats
	if prevOverall != nil {
		stats.Overall = calculateCPUPercent(*prevOverall, *currOverall)
	} else {
		stats.Overall = 0.0
	}

	// Per-core stats
	var cpuTicks *C.natural_t
	var numCPUs C.natural_t
	var numTicks C.mach_msg_type_number_t
	var currPerCore []cpuTime

	if res := C.get_processor_loads(&cpuTicks, &numCPUs, &numTicks); res == 0 {
		defer C.vm_deallocate(C.mach_task_self_, C.vm_address_t(uintptr(unsafe.Pointer(cpuTicks))), C.vm_size_t(numTicks*4))

		slice := (*[1 << 28]C.natural_t)(unsafe.Pointer(cpuTicks))[:numTicks:numTicks]
		for i := 0; i < int(numCPUs); i++ {
			offset := i * int(C.CPU_STATE_MAX)
			if offset+int(C.CPU_STATE_NICE) < len(slice) {
				currPerCore = append(currPerCore, cpuTime{
					user:   uint64(slice[offset+int(C.CPU_STATE_USER)]),
					system: uint64(slice[offset+int(C.CPU_STATE_SYSTEM)]),
					idle:   uint64(slice[offset+int(C.CPU_STATE_IDLE)]),
					nice:   uint64(slice[offset+int(C.CPU_STATE_NICE)]),
				})
			}
		}

		if prevPerCore != nil && len(prevPerCore) == len(currPerCore) {
			for i := range currPerCore {
				stats.PerCore = append(stats.PerCore, calculateCPUPercent(prevPerCore[i], currPerCore[i]))
			}
		} else {
			for range currPerCore {
				stats.PerCore = append(stats.PerCore, stats.Overall)
			}
		}
	} else {
		// Fallback to ncpu sysctl
		ncpu, err := unix.SysctlUint32("hw.ncpu")
		if err == nil {
			for i := uint32(0); i < ncpu; i++ {
				stats.PerCore = append(stats.PerCore, stats.Overall)
			}
		}
	}

	return &stats, currOverall, currPerCore, nil
}
