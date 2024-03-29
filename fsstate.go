package systemmonitor

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

/*
var (
	diskIgnore = []string{
		"loop",
		"sr",
	}
)
*/

var fsSpecIgnore = map[string]struct{}{
	"/dev/root":   {},
	"none":        {},
	"nodev":       {},
	"proc":        {},
	"hugetlbfs":   {},
	"sysfs":       {},
	"securityfs":  {},
	"binfmt_misc": {},
	"pstore":      {},
	"gvfsd-fuse":  {},
}

var fsTypeIgnore = map[string]struct{}{
	"cgroup":     {},
	"debugfs":    {},
	"devpts":     {},
	"devtmpfs":   {},
	"rpc_pipefs": {},
	"rootfs":     {},
	"overlay":    {},
	"tmpfs":      {},
	"pstore":     {},
	"autofs":     {},
	"mqueue":     {},
	"configfs":   {},
	"fusectl":    {},
	"nfsd":       {},
	"squashfs":   {},
}

// DiskStates 磁盘状态
type DiskStates struct {
	Device    string
	FsSpec    string
	FsFile    string
	FsVfstype string

	BytesAll             uint64
	BytesUsed            uint64
	BytesUsedPermillage  int
	InodesAll            uint64
	InodesUsed           uint64
	InodesUsedPermillage int

	ReadRequests  uint64 // Total number of reads completed successfully.
	ReadBytes     uint64 // Total number of Bytes read successfully.
	WriteRequests uint64 // total number of writes completed successfully.
	WriteBytes    uint64 // total number of Bytes written successfully.

	preReadRequests  uint64 // Total number of reads completed successfully.
	preReadBytes     uint64 // Total number of Bytes read successfully.
	preWriteRequests uint64 // total number of writes completed successfully.
	preWriteBytes    uint64 // total number of Bytes written successfully.
}

// Fsinfo 文件系统信息
type Fsinfo struct {
	Disks map[string]*DiskStates
}

func getDevice(dir string, uuidorlabel string) (string, error) {
	var device string
	if len(uuidorlabel) < 1 {
		return "", nil
	}
	errorcode := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		_, err1 := os.Readlink(path)
		devpath, err2 := filepath.EvalSymlinks(path)
		if f.IsDir() || err1 != nil || err2 != nil {
			return nil
		}
		if uuidorlabel == f.Name() {
			device = devpath
		}
		return nil
	})
	return device, errorcode
}

func getDeviceByUUID(uuid string) (string, error) {
	device, errorcode := getDevice("/dev/disk/by-uuid", uuid)
	return device, errorcode
}

func getDeviceByLABEL(label string) (string, error) {
	device, errorcode := getDevice("/dev/disk/by-label", label)
	return device, errorcode
}

/*
func listdisks() (disks []string, err error) {
	disks = make([]string, 0, 10)

	dir, err := ioutil.ReadDir("/sys/block/")
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		var isignored bool = false
		for _, ignore := range diskIgnore {
			if strings.Index(fi.Name(), ignore) == 0 {
				isignored = true
			}
		}
		if !isignored {
			disks = append(disks, fi.Name())
		}
	}

	return disks, nil
}
*/

type mountpoint struct {
	FsSpec    string // 分区设备节点，如/dev/sda1
	FsFile    string // 分区挂载目录
	FsVfstype string // 分区格式
}

func listrootfspartition() (*mountpoint, error) {
	var points *mountpoint
	var dev string
	file, err := os.Open("/proc/cmdline")
	if err != nil {
		panic(err)
	}
	rd := bufio.NewReader(file)
	for {
		line, err2 := rd.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fields := strings.Fields(string(line))
		if len(fields) < 1 {
			continue
		}
		for _, p := range fields {
			index1 := strings.Index(p, "root=")
			index2 := strings.Index(p, "root=UUID=")
			index3 := strings.Index(p, "root=LABEL=")
			if index1 < 0 {
				continue
			}
			if index2 == 0 {
				dev, err2 = getDeviceByUUID(p[len("root=UUID="):])
			} else if index3 == 0 {
				dev, err2 = getDeviceByLABEL(p[len("root=LABEL="):])
			} else {
				dev = p[len("root="):]
			}
			if err2 == nil && len(dev) > 0 {
				points = &mountpoint{
					FsSpec:    dev,
					FsFile:    "/",
					FsVfstype: "unknown",
				}
				break
			}
		}
	}
	file.Close()
	return points, nil
}

func listmountedpartition() (map[string]*mountpoint, error) {
	points := make(map[string]*mountpoint)
	rootpoint, err1 := listrootfspartition()
	if rootpoint != nil && err1 == nil {
		points[rootpoint.FsSpec] = rootpoint
	}
	file, err := os.Open("/proc/mounts")
	if err != nil {
		panic(err)
	}
	rd := bufio.NewReader(file)
	for {
		line, err2 := rd.ReadString('\n')
		if err != nil || io.EOF == err2 {
			break
		}
		fields := strings.Fields(string(line))
		if len(fields) != 6 || fields[0] == "" {
			continue
		}

		fsSpec := fields[0]
		fsFile := fields[1]
		fsVfstype := fields[2]

		if _, exist := fsSpecIgnore[fsSpec]; exist {
			continue
		}
		if _, exist := fsTypeIgnore[fsVfstype]; exist {
			continue
		}
		if fsFile == "/" {
			if rootpoint != nil {
				points[rootpoint.FsSpec].FsVfstype = fsVfstype
			}
		}

		if _, ok := points[fsSpec]; !ok {
			points[fsSpec] = &mountpoint{
				FsSpec:    fsSpec,
				FsFile:    fsFile,
				FsVfstype: fsVfstype,
			}
		}
	}
	file.Close()
	return points, nil
}

type diskstate struct {
	ReadRequests  uint64 // Total number of reads completed successfully.
	ReadBytes     uint64 // Total number of Bytes read successfully.
	WriteRequests uint64 // total number of writes completed successfully.
	WriteBytes    uint64 // total number of Bytes written successfully.
}

func listdiskstate(device string) (*diskstate, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		panic(err)
	}
	statedisk := &diskstate{}
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		fields := strings.Fields(string(line))
		if len(fields) != 14 || fields[2] == "" || fields[2] != device {
			continue
		}

		if statedisk.ReadRequests, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
			return nil, err
		}
		if statedisk.ReadBytes, err = strconv.ParseUint(fields[5], 10, 64); err != nil {
			return nil, err
		}
		statedisk.ReadBytes = statedisk.ReadBytes * 512
		if statedisk.WriteRequests, err = strconv.ParseUint(fields[7], 10, 64); err != nil {
			return nil, err
		}
		if statedisk.WriteBytes, err = strconv.ParseUint(fields[9], 10, 64); err != nil {
			return nil, err
		}
		statedisk.WriteBytes = statedisk.WriteBytes * 512
	}
	file.Close()
	return statedisk, nil
}

func (fsinfo *Fsinfo) getfsstate() error {
	if fsinfo.Disks == nil {
		fsinfo.Disks = make(map[string]*DiskStates)
	}
	points, err := listmountedpartition()
	if err != nil {
		return err
	}
	for _, p := range points {
		if !strings.Contains(p.FsSpec, "/dev/") {
			continue
		}
		device := p.FsSpec[len("/dev/"):]
		if _, ok := fsinfo.Disks[device]; !ok {
			fsinfo.Disks[device] = &DiskStates{}
			fsinfo.Disks[device].Device = device
		}
		disk := fsinfo.Disks[device]
		disk.FsSpec = p.FsSpec
		disk.FsFile = p.FsFile
		disk.FsVfstype = p.FsVfstype

		fs := syscall.Statfs_t{}
		err := syscall.Statfs(p.FsFile, &fs)
		if err != nil {
			return err
		}

		disk.BytesAll = uint64(fs.Frsize) * fs.Blocks
		disk.BytesUsed = uint64(fs.Frsize) * (fs.Blocks - fs.Bfree)
		disk.BytesUsedPermillage = int(float64(disk.BytesUsed) / float64(disk.BytesAll) * 1000)
		disk.InodesAll = fs.Files
		disk.InodesUsed = fs.Files - fs.Ffree
		disk.InodesUsedPermillage = int(float64(disk.InodesUsed) / float64(disk.InodesAll) * 1000)

		statedisk, diskerr := listdiskstate(device)
		if diskerr != nil {
			continue
		}
		disk.ReadRequests = statedisk.ReadRequests - disk.preReadRequests
		disk.ReadBytes = statedisk.ReadBytes - disk.preReadBytes
		disk.WriteRequests = statedisk.WriteRequests - disk.preWriteRequests
		disk.WriteBytes = statedisk.WriteBytes - disk.preWriteBytes

		disk.preReadRequests = statedisk.ReadRequests
		disk.preReadBytes = statedisk.ReadBytes
		disk.preWriteRequests = statedisk.WriteRequests
		disk.preWriteBytes = statedisk.WriteBytes
	}
	return nil
}
