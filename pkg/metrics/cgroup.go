/*
Copyright 2022 shaowenchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	cgroupV1CPUPath    = "/sys/fs/cgroup/cpu/cpuacct.usage"
	cgroupV1MemoryPath = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	cgroupV2Path       = "/sys/fs/cgroup"
	cgroupV2CPUStat    = "cpu.stat"
	cgroupV2MemoryCur  = "memory.current"
)

// CGroupVersion represents the cgroup version
type CGroupVersion int

const (
	CGroupV1 CGroupVersion = iota
	CGroupV2
	CGroupUnknown
)

// detectCGroupVersion detects the cgroup version
func detectCGroupVersion() CGroupVersion {
	// Check for cgroup v2 (unified hierarchy)
	if _, err := os.Stat(filepath.Join(cgroupV2Path, cgroupV2CPUStat)); err == nil {
		if _, err := os.Stat(filepath.Join(cgroupV2Path, cgroupV2MemoryCur)); err == nil {
			return CGroupV2
		}
	}

	// Check for cgroup v1
	if _, err := os.Stat(cgroupV1CPUPath); err == nil {
		if _, err := os.Stat(cgroupV1MemoryPath); err == nil {
			return CGroupV1
		}
	}

	return CGroupUnknown
}

// findCGroupPath finds the cgroup path for the current process
func findCGroupPath() (string, error) {
	version := detectCGroupVersion()

	cgroupFile := "/proc/self/cgroup"
	data, err := os.ReadFile(cgroupFile)
	if err != nil {
		if version == CGroupV2 {
			return cgroupV2Path, nil
		}
		return "", fmt.Errorf("failed to read cgroup file: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		// For cgroup v2, format is: 0::/path
		if version == CGroupV2 {
			if strings.HasPrefix(line, "0::") {
				parts := strings.Split(line, ":")
				if len(parts) >= 3 {
					path := strings.TrimPrefix(parts[2], "/")
					if path != "" {
						fullPath := filepath.Join(cgroupV2Path, path)
						if _, err := os.Stat(fullPath); err == nil {
							return fullPath, nil
						}
					}
				}
			}
		} else if version == CGroupV1 {
			// For cgroup v1, format is: hierarchyID:controllers:path
			// We need to find the line with cpu or cpuacct controller
			parts := strings.Split(line, ":")
			if len(parts) >= 3 {
				controllers := parts[1]
				path := parts[2]
				// Check if this line has cpu or cpuacct controller
				if strings.Contains(controllers, "cpu") || strings.Contains(controllers, "cpuacct") {
					if path != "" && path != "/" {
						// Try to find the actual cgroup mount point
						// Common mount points: /sys/fs/cgroup/cpu or /sys/fs/cgroup/cpuacct
						for _, mountPoint := range []string{"/sys/fs/cgroup/cpu", "/sys/fs/cgroup/cpuacct"} {
							fullPath := filepath.Join(mountPoint, path)
							if _, err := os.Stat(fullPath); err == nil {
								return fullPath, nil
							}
						}
					}
				}
			}
		}
	}

	// Fallback
	if version == CGroupV2 {
		return cgroupV2Path, nil
	}
	if version == CGroupV1 {
		return "/sys/fs/cgroup/cpu", nil
	}

	return "", fmt.Errorf("unable to find cgroup path: unsupported cgroup version")
}

// findCGroupV1Paths finds CPU and memory cgroup paths for v1
func findCGroupV1Paths() (cpuPath, memoryPath string, err error) {
	cgroupFile := "/proc/self/cgroup"
	data, err := os.ReadFile(cgroupFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cgroup file: %w", err)
	}

	var cpuCgroupPath, memoryCgroupPath string

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			controllers := parts[1]
			path := parts[2]

			if path != "" && path != "/" {
				// Find CPU path
				if (strings.Contains(controllers, "cpu") || strings.Contains(controllers, "cpuacct")) && cpuCgroupPath == "" {
					for _, mountPoint := range []string{"/sys/fs/cgroup/cpu", "/sys/fs/cgroup/cpuacct"} {
						fullPath := filepath.Join(mountPoint, path, "cpuacct.usage")
						if _, err := os.Stat(fullPath); err == nil {
							cpuCgroupPath = filepath.Join(mountPoint, path)
							break
						}
					}
				}

				// Find memory path
				if strings.Contains(controllers, "memory") && memoryCgroupPath == "" {
					for _, mountPoint := range []string{"/sys/fs/cgroup/memory"} {
						fullPath := filepath.Join(mountPoint, path, "memory.usage_in_bytes")
						if _, err := os.Stat(fullPath); err == nil {
							memoryCgroupPath = filepath.Join(mountPoint, path)
							break
						}
					}
				}
			}
		}
	}

	if cpuCgroupPath == "" {
		cpuCgroupPath = "/sys/fs/cgroup/cpu"
	}
	if memoryCgroupPath == "" {
		memoryCgroupPath = "/sys/fs/cgroup/memory"
	}

	return cpuCgroupPath, memoryCgroupPath, nil
}

// GetCPUUsage returns CPU usage in seconds (cumulative)
func GetCPUUsage() (float64, error) {
	version := detectCGroupVersion()

	switch version {
	case CGroupV1:
		return getCPUUsageV1()
	case CGroupV2:
		return getCPUUsageV2()
	default:
		return 0, fmt.Errorf("unsupported cgroup version")
	}
}

// getCPUUsageV1 reads CPU usage from cgroup v1
func getCPUUsageV1() (float64, error) {
	cpuPath, _, err := findCGroupV1Paths()
	if err != nil {
		return 0, err
	}

	cpuUsagePath := filepath.Join(cpuPath, "cpuacct.usage")
	// Fallback to default path if the found path doesn't exist
	if _, err := os.Stat(cpuUsagePath); err != nil {
		cpuUsagePath = cgroupV1CPUPath
	}

	data, err := os.ReadFile(cpuUsagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v1 CPU usage from %s: %w", cpuUsagePath, err)
	}

	usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU usage: %w", err)
	}

	// Convert from nanoseconds to seconds
	return float64(usage) / 1e9, nil
}

// getCPUUsageV2 reads CPU usage from cgroup v2
func getCPUUsageV2() (float64, error) {
	cgroupPath, err := findCGroupPath()
	if err != nil {
		return 0, err
	}

	// Try the found path first
	cpuStatPath := filepath.Join(cgroupPath, cgroupV2CPUStat)
	if _, err := os.Stat(cpuStatPath); err != nil {
		// Try root cgroup
		cpuStatPath = filepath.Join(cgroupV2Path, cgroupV2CPUStat)
		if _, err := os.Stat(cpuStatPath); err != nil {
			return 0, fmt.Errorf("cpu.stat not found: %w", err)
		}
	}

	data, err := os.ReadFile(cpuStatPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v2 CPU stat: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "usage_usec") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				usage, err := strconv.ParseUint(parts[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("failed to parse CPU usage_usec: %w", err)
				}
				// Convert from microseconds to seconds
				return float64(usage) / 1e6, nil
			}
		}
	}

	return 0, fmt.Errorf("usage_usec not found in cpu.stat")
}

// GetMemoryUsage returns memory usage in bytes
func GetMemoryUsage() (uint64, error) {
	version := detectCGroupVersion()

	switch version {
	case CGroupV1:
		return getMemoryUsageV1()
	case CGroupV2:
		return getMemoryUsageV2()
	default:
		return 0, fmt.Errorf("unsupported cgroup version")
	}
}

// getMemoryUsageV1 reads memory usage from cgroup v1
func getMemoryUsageV1() (uint64, error) {
	_, memoryPath, err := findCGroupV1Paths()
	if err != nil {
		return 0, err
	}

	memoryUsagePath := filepath.Join(memoryPath, "memory.usage_in_bytes")
	// Fallback to default path if the found path doesn't exist
	if _, err := os.Stat(memoryUsagePath); err != nil {
		memoryUsagePath = cgroupV1MemoryPath
	}

	data, err := os.ReadFile(memoryUsagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v1 memory usage from %s: %w", memoryUsagePath, err)
	}

	usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory usage: %w", err)
	}

	return usage, nil
}

// getMemoryUsageV2 reads memory usage from cgroup v2
func getMemoryUsageV2() (uint64, error) {
	cgroupPath, err := findCGroupPath()
	if err != nil {
		return 0, err
	}

	// Try the found path first
	memoryCurPath := filepath.Join(cgroupPath, cgroupV2MemoryCur)
	if _, err := os.Stat(memoryCurPath); err != nil {
		// Try root cgroup
		memoryCurPath = filepath.Join(cgroupV2Path, cgroupV2MemoryCur)
		if _, err := os.Stat(memoryCurPath); err != nil {
			return 0, fmt.Errorf("memory.current not found: %w", err)
		}
	}

	data, err := os.ReadFile(memoryCurPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v2 memory usage: %w", err)
	}

	usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory usage: %w", err)
	}

	return usage, nil
}
