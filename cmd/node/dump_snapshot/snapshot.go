package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"sort"
	"strings"
)

type snapshot struct {
	height int64
	format uint32
	chunks uint
}
type snapshots []snapshot

func (ss snapshots) Sort() snapshots {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].height > ss[j].height
	})
	return ss
}

func (ss snapshot) HeightStr() string {
	return fmt.Sprintf("%d", ss.height)
}

func (ss snapshot) FormatStr() string {
	return fmt.Sprintf("%d", ss.format)
}

func loadSnapshotList(binary, nodeHomeDirectory string) (snapshots, error) {
	output, ec := utils.LaunchAppAndGetOutput(binary, []string{"snapshots", "list", "--home", nodeHomeDirectory})
	if ec != 0 {
		return nil, fmt.Errorf("failed to list snapshots")
	}

	var snapshots []snapshot
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "height:") {
			continue
		}
		if !strings.Contains(line, "format:") {
			continue
		}
		if !strings.Contains(line, "chunks:") {
			continue
		}

		var s snapshot
		_, err := fmt.Sscanf(strings.TrimSpace(line), "height: %d format: %d chunks: %d", &s.height, &s.format, &s.chunks)
		if err != nil {
			return nil, fmt.Errorf("failed to parse snapshot line: %s", line)
		}
		if s.height == 0 || s.format == 0 || s.chunks == 0 {
			return nil, fmt.Errorf("invalid snapshot line: %s, value %v", line, s)
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}
