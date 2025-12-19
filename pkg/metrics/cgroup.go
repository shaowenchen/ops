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

	// For cgroup v2, try to find path from /proc/self/cgroup
	if version == CGroupV2 {
		cgroupFile := "/proc/self/cgroup"
		data, err := os.ReadFile(cgroupFile)
		if err != nil {
			// Fallback to root cgroup
			return cgroupV2Path, nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			// For cgroup v2, format is: 0::/path
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
		}
		// Fallback to root cgroup
		return cgroupV2Path, nil
	}

	// For cgroup v1, we use the root paths directly
	// CPU and memory might be in different paths, but we'll use the common parent
	if version == CGroupV1 {
		return "/sys/fs/cgroup", nil
	}

	return "", fmt.Errorf("unable to find cgroup path: unsupported cgroup version")
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
	data, err := os.ReadFile(cgroupV1CPUPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v1 CPU usage: %w", err)
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
	data, err := os.ReadFile(cgroupV1MemoryPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read cgroup v1 memory usage: %w", err)
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
