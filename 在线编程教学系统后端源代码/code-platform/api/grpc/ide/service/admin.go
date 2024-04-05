package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"code-platform/api/grpc/ide/pb"
	"code-platform/pkg/errorx"
	"code-platform/pkg/osx"
	"code-platform/pkg/strconvx"
	"code-platform/pkg/timex"
	"code-platform/service/ide/define"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	listContainersFormat = `"{{.ID}}\t{{.Names}}\t{{.Ports}}\t{{.CreatedAt}}\t{{.Size}}"`

	timeFormat = "2006-01-02 15:04:05 -0700 MST"
)

var (
	portCompile  = regexp.MustCompile(`^.*?(\d+)->`)
	usageCompile = regexp.MustCompile(`^(.*?)\t(.*?)$`)
)

var sortFuncMap = map[pb.OrderType]func(records interface{}, isReverse bool) func(i, j int) bool{
	pb.OrderType_byTime: func(records interface{}, isReverse bool) func(i, j int) bool {
		tables := records.([]*rowData)
		if isReverse {
			return func(i, j int) bool {
				return tables[i].createdAt.After(tables[j].createdAt)
			}
		}
		return func(i, j int) bool {
			return tables[i].createdAt.Before(tables[j].createdAt)
		}
	},

	pb.OrderType_byDiskSize: func(records interface{}, isReverse bool) func(i, j int) bool {
		tables := records.([]*rowData)
		if isReverse {
			return func(i, j int) bool {
				return tables[i].rawsize > tables[j].rawsize
			}
		}
		return func(i, j int) bool {
			return tables[i].rawsize < tables[j].rawsize
		}
	},

	pb.OrderType_byCPU: func(records interface{}, isReverse bool) func(i, j int) bool {
		tables := records.([]*rowDataForUsage)
		if isReverse {
			return func(i, j int) bool {
				return tables[i].cpuPerc > tables[j].cpuPerc
			}
		}
		return func(i, j int) bool {
			return tables[i].cpuPerc < tables[j].cpuPerc
		}
	},

	pb.OrderType_byMemory: func(records interface{}, isReverse bool) func(i, j int) bool {
		tables := records.([]*rowDataForUsage)
		if isReverse {
			return func(i, j int) bool {
				return tables[i].memPerc > tables[j].memPerc
			}
		}
		return func(i, j int) bool {
			return tables[i].memPerc < tables[j].memPerc
		}
	},
}

type (
	rowData struct {
		createdAt     time.Time
		containerID   string
		containerName string
		rawPort       string
		rawsize       string
	}

	usageInfo struct {
		cpuPerc  string
		memUsage string
	}

	rowDataForUsage struct {
		*usageInfo
		containerID string
		cpuPerc     float64
		memPerc     float64
	}
)

func (i *IDEServer) GetContainers(ctx context.Context, req *pb.GetContainersRequest) (*pb.GetContainersResponse, error) {
	switch req.Order {
	case pb.OrderType_byTime, pb.OrderType_byDiskSize:
		return i.getContainersByLS(ctx, req)
	case pb.OrderType_byCPU, pb.OrderType_byMemory:
		return i.getContainersByStats(ctx, req)
	}
	return nil, status.Errorf(codes.Unimplemented, "unknown orderType %q", req.Order.String())
}

func (i *IDEServer) getContainersByLS(ctx context.Context, req *pb.GetContainersRequest) (*pb.GetContainersResponse, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", listContainersFormat)
	stdout, stderr, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		i.Logger.Errorf(err, "run command %q failed for %s", cmd.String(), string(stderr))
		return nil, status.Errorf(codes.Internal, "err %v for %q", err, string(stderr))
	}

	if len(stdout) == 0 {
		return &pb.GetContainersResponse{}, nil
	}

	containerInfos, total, err := getContainerInfosByLS(stdout, int(req.Offset), int(req.Limit), req.Order, req.IsReverse)
	if err != nil {
		i.Logger.Error(err, "get container infos failed")
		return nil, status.Error(codes.Internal, err.Error())
	}

	containerIDs := make([]string, 0, len(containerInfos))
	for _, containerInfo := range containerInfos {
		containerIDs = append(containerIDs, containerInfo.ContainerId)
	}

	args := make([]string, 0, len(containerIDs)+4)
	args = append(args, "stats", "--no-stream", "--format", `"{{.CPUPerc}}\t{{.MemUsage}}"`)
	args = append(args, containerIDs...)
	usageCmd := exec.CommandContext(ctx, "docker", args...)
	stdout, stderr, err = osx.CommandOutput(ctx, usageCmd)
	if err != nil {
		i.Logger.Errorf(err, "run command %q failed for %s", cmd.String(), string(stderr))
		return nil, status.Errorf(codes.Internal, "err %v for %q", err, string(stderr))
	}

	usages, err := getContainerUsages(stdout)
	if err != nil {
		i.Logger.Error(err, "get container usage failed")
		return nil, status.Error(codes.Internal, err.Error())
	}

	for i := range containerInfos {
		containerInfos[i].CpuPerc = usages[i].cpuPerc
		containerInfos[i].MemoryUsage = usages[i].memUsage
	}

	return &pb.GetContainersResponse{ContainerInfos: containerInfos, Total: total}, nil
}

func (i *IDEServer) getContainersByStats(ctx context.Context, req *pb.GetContainersRequest) (*pb.GetContainersResponse, error) {
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", `"{{.ID}}\t{{.CPUPerc}}\t{{.MemPerc}}\t{{.MemUsage}}"`)
	stdout, stderr, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		i.Logger.Errorf(err, "run command %q failed for %s", cmd.String(), string(stderr))
		return nil, status.Errorf(codes.Internal, "err %v for %q", err, string(stderr))
	}

	if len(stdout) == 0 {
		return &pb.GetContainersResponse{}, nil
	}

	containerInfos, total, err := getContainerInfosByStats(stdout, int(req.Offset), int(req.Limit), req.Order, req.IsReverse)
	if err != nil {
		i.Logger.Error(err, "get container infos by stats failed")
		return nil, status.Error(codes.Internal, err.Error())
	}

	args := make([]string, 0, 2*len(containerInfos)+3)
	args = append(args, "ps", "--format", listContainersFormat)

	for _, containerInfo := range containerInfos {
		args = append(args, "--filter", "id="+containerInfo.ContainerId)
	}

	psCmd := exec.CommandContext(ctx, "docker", args...)
	stdout, stderr, err = osx.CommandOutput(ctx, psCmd)
	if err != nil {
		i.Logger.Errorf(err, "run command %q failed for %s", cmd.String(), string(stderr))
		return nil, status.Errorf(codes.Internal, "err %v for %q", err, string(stderr))
	}

	if len(stdout) == 0 {
		i.Logger.Errorf(err, "stdout shouldn't be empty")
		return nil, status.Errorf(codes.Internal, "stdout shouldn't be empty")
	}

	m, err := getContainersBasicMap(stdout)
	if err != nil {
		i.Logger.Error(err, "get container basic map failed")
		return nil, status.Error(codes.Internal, err.Error())
	}
	for index, containerInfo := range containerInfos {
		bs := m[containerInfo.ContainerId]
		if bs != nil {
			labID, studentID, teacherInfo, err := containerNameToIDs(bs.containerName)
			if err != nil {
				i.Logger.Errorf(err, "containerName %q to ids failed", bs.containerName)
				return nil, status.Error(codes.Internal, err.Error())
			}
			portSlice := portCompile.FindStringSubmatch(bs.rawPort)
			if portSlice == nil || len(portSlice) != 2 {
				err := fmt.Errorf("%q is not a valid port", bs.rawPort)
				i.Logger.Error(err, "")
				return nil, status.Error(codes.Internal, err.Error())
			}

			port, err := strconv.Atoi(portSlice[1])
			if err != nil {
				i.Logger.Errorf(err, "%q is not a valid port", portSlice[1])
				return nil, status.Error(codes.Internal, err.Error())
			}

			sizeSlice := strings.Fields(bs.rawsize)
			if len(sizeSlice) == 0 {
				err := fmt.Errorf("%q is not a valid size", bs.rawPort)
				i.Logger.Error(err, "")
				return nil, status.Error(codes.Internal, err.Error())
			}
			containerInfos[index].CreatedAt = bs.createdAt.Unix()
			containerInfos[index].LabId = labID
			containerInfos[index].StudentId = studentID
			containerInfos[index].TeacherInfo = teacherInfo
			containerInfos[index].Port = uint32(port)
			containerInfos[index].Size = sizeSlice[0]
		}
	}
	return &pb.GetContainersResponse{ContainerInfos: containerInfos, Total: total}, nil
}

func getContainerInfosByStats(stdout []byte, offset, limit int, order pb.OrderType, isReverse bool) ([]*pb.GetContainersResponse_ContainerInfo, uint32, error) {
	if len(stdout) == 0 {
		return nil, 0, nil
	}

	rows := bytes.Split(stdout, []byte("\n"))

	rows = rows[:len(rows)-1]
	if offset >= len(rows) {
		return nil, 0, nil
	}

	table := make([]*rowDataForUsage, 0, len(rows))
	for _, row := range rows {
		cols := strings.Split(string(row), "\t")
		if len(cols) != 4 {
			return nil, 0, fmt.Errorf("%q is not standard container format", string(row))
		}

		cpuPerc, err := strconv.ParseFloat(cols[1][:len(cols[1])-1], 10)
		if err != nil {
			return nil, 0, fmt.Errorf("cpuPerc %q is invalid", cols[1])
		}

		memPerc, err := strconv.ParseFloat(cols[2][:len(cols[2])-1], 10)
		if err != nil {
			return nil, 0, fmt.Errorf("memPerc %q is invalid", cols[2])
		}

		table = append(table, &rowDataForUsage{
			containerID: cols[0][1:],
			usageInfo: &usageInfo{
				cpuPerc:  cols[1],
				memUsage: cols[3][:len(cols[3])-1],
			},
			cpuPerc: cpuPerc,
			memPerc: memPerc,
		})
	}

	sort.Slice(table, sortFuncMap[order](table, isReverse))

	table = table[offset:]
	if len(table) > limit {
		table = table[:limit:limit]
	}

	containersInfos := make([]*pb.GetContainersResponse_ContainerInfo, 0, len(rows))
	for _, row := range table {
		containersInfos = append(containersInfos, &pb.GetContainersResponse_ContainerInfo{
			ContainerId: row.containerID,
			CpuPerc:     row.usageInfo.cpuPerc,
			MemoryUsage: row.memUsage,
		})
	}

	return containersInfos, uint32(len(rows)), nil
}

func getContainersBasicMap(stdout []byte) (map[string]*rowData, error) {
	if len(stdout) == 0 {
		return nil, nil
	}

	rows := bytes.Split(stdout, []byte("\n"))
	rows = rows[:len(rows)-1]

	table := make([]*rowData, 0, len(rows))
	for _, row := range rows {
		cols := strings.Split(string(row), "\t")
		if len(cols) != 5 {
			return nil, fmt.Errorf("%q is not standard container format", string(row))
		}

		createdAt, err := time.ParseInLocation(timeFormat, cols[3], timex.ShanghaiLocation)
		if err != nil {
			return nil, fmt.Errorf("time parse %q failed for err %v", cols[3], err)
		}

		table = append(table, &rowData{
			// 去掉之前引号
			containerID:   cols[0][1:],
			containerName: cols[1],
			rawPort:       cols[2],
			rawsize:       cols[4],
			createdAt:     createdAt,
		})
	}

	m := make(map[string]*rowData)
	for _, row := range table {
		m[row.containerID] = row
	}

	return m, nil
}

func (i *IDEServer) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.Empty, error) {
	cmd := exec.CommandContext(ctx, "docker", "stop", "-t", "3", req.ContainerId)
	_, stderr, err := osx.CommandOutput(ctx, cmd)
	switch err {
	case nil:
	case errorx.ErrWrongCode:
		return nil, status.Error(codes.Aborted, err.Error())
	default:
		i.Logger.Errorf(err, "run command %q failed for %s", cmd.String(), string(stderr))
		return nil, status.Errorf(codes.Internal, "err %v for %q", err, string(stderr))
	}
	return &pb.Empty{}, nil
}

func containerNameToIDs(name string) (labID, studentID uint64, teacherInfo *pb.TeacherInfo, err error) {
	var (
		isTeacher bool
		teacherID uint64
	)

	ids := strings.Split(strings.TrimPrefix(name, define.ContainerNamePrefix), "-")
	switch len(ids) {
	case 2:
		isTeacher = false
	case 3:
		isTeacher = true
	default:
		return 0, 0, nil, fmt.Errorf("%q is not valid container name", name)
	}
	labID, err = strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("%q is not as labID", ids[0])
	}
	studentID, err = strconv.ParseUint(ids[1], 10, 64)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("%q is not as studentID", ids[1])
	}
	if isTeacher {
		teacherID, err = strconv.ParseUint(ids[2], 10, 64)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("%q is not as studentID", ids[0])
		}

		teacherInfo = &pb.TeacherInfo{
			TeacherId: teacherID,
		}
	}
	return labID, studentID, teacherInfo, nil
}

func getContainerUsages(stdout []byte) ([]*usageInfo, error) {
	if len(stdout) == 0 {
		return nil, nil
	}

	rows := bytes.Split(stdout, []byte("\n"))
	// 最后一行为空行，真实行数应为 len(rows) - 1
	rows = rows[:len(rows)-1]
	usageInfos := make([]*usageInfo, 0, len(rows))
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		usageSlice := usageCompile.FindSubmatch(row)
		if usageSlice == nil || len(usageSlice) != 3 {
			return nil, fmt.Errorf("%q is not a valid port", string(row))
		}
		usageInfos = append(usageInfos, &usageInfo{
			cpuPerc:  strconvx.BytesToString(usageSlice[1][1:]),
			memUsage: strconvx.BytesToString(usageSlice[2][:len(usageSlice[2])-1]),
		})
	}

	return usageInfos, nil
}

func getContainerInfosByLS(stdout []byte, offset, limit int, order pb.OrderType, isReverse bool) ([]*pb.GetContainersResponse_ContainerInfo, uint32, error) {
	if len(stdout) == 0 {
		return nil, 0, nil
	}

	rows := bytes.Split(stdout, []byte("\n"))

	// 最后一行为空行，真实行数应为 len(rows) - 1
	rows = rows[:len(rows)-1]
	if offset >= len(rows) {
		return nil, 0, nil
	}

	containersInfos := make([]*pb.GetContainersResponse_ContainerInfo, 0, len(rows))

	table := make([]*rowData, 0, len(rows))
	for _, row := range rows {
		cols := strings.Split(string(row), "\t")
		if len(cols) != 5 {
			return nil, 0, fmt.Errorf("%q is not standard container format", string(row))
		}

		createdAt, err := time.ParseInLocation(timeFormat, cols[3], timex.ShanghaiLocation)
		if err != nil {
			return nil, 0, fmt.Errorf("time parse %q failed for err %v", cols[3], err)
		}

		table = append(table, &rowData{
			// 去掉之前引号
			containerID:   cols[0][1:],
			containerName: cols[1],
			rawPort:       cols[2],
			rawsize:       cols[4],
			createdAt:     createdAt,
		})
	}

	table = table[:len(table):len(table)]

	sort.Slice(table, sortFuncMap[order](table, isReverse))

	table = table[offset:]
	if len(table) > limit {
		table = table[:limit:limit]
	}

	for _, row := range table {
		labID, studentID, teacherInfo, err := containerNameToIDs(row.containerName)
		if err != nil {
			return nil, 0, err
		}
		portSlice := portCompile.FindStringSubmatch(row.rawPort)
		if portSlice == nil || len(portSlice) != 2 {
			return nil, 0, fmt.Errorf("%q is not a valid port", row.rawPort)
		}

		port, err := strconv.Atoi(portSlice[1])
		if err != nil {
			return nil, 0, fmt.Errorf("%q is not a valid port", portSlice[1])
		}

		sizeSlice := strings.Fields(row.rawsize)
		if len(sizeSlice) == 0 {
			return nil, 0, fmt.Errorf("%q is not a valid size", row.rawPort)
		}

		containersInfos = append(containersInfos, &pb.GetContainersResponse_ContainerInfo{
			// 去掉之前引号
			ContainerId: row.containerID,
			LabId:       labID,
			StudentId:   studentID,
			CreatedAt:   row.createdAt.Unix(),
			Size:        sizeSlice[0],
			TeacherInfo: teacherInfo,
			Port:        uint32(port),
		})
	}

	return containersInfos[:len(containersInfos):len(containersInfos)], uint32(len(rows)), nil
}
