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

var (
	disk_ignore = []string{
		"loop",
		"sr",
	}
)

var FSSPEC_IGNORE = map[string]struct{}{
	"/dev/root":   struct{}{},
	"none":        struct{}{},
	"nodev":       struct{}{},
	"proc":        struct{}{},
	"hugetlbfs":   struct{}{},
	"sysfs":       struct{}{},
	"securityfs":  struct{}{},
	"binfmt_misc": struct{}{},
	"pstore":      struct{}{},
	"gvfsd-fuse":  struct{}{},
}

var FSTYPE_IGNORE = map[string]struct{}{
	"cgroup":     struct{}{},
	"debugfs":    struct{}{},
	"devpts":     struct{}{},
	"devtmpfs":   struct{}{},
	"rpc_pipefs": struct{}{},
	"rootfs":     struct{}{},
	"overlay":    struct{}{},
	"tmpfs":      struct{}{},
	"pstore":     struct{}{},
	"autofs":     struct{}{},
	"mqueue":     struct{}{},
	"configfs":   struct{}{},
	"fusectl":    struct{}{},
	"nfsd":       struct{}{},
	"squashfs":   struct{}{},
}

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

	pre_ReadRequests  uint64 // Total number of reads completed successfully.
	pre_ReadBytes     uint64 // Total number of Bytes read successfully.
	pre_WriteRequests uint64 // total number of writes completed successfully.
	pre_WriteBytes    uint64 // total number of Bytes written successfully.
}

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

type mountpoint struct {
	FsSpec    string
	FsFile    string
	FsVfstype string
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
					FsVfstype: "TODO",
				}
				break
			}
		}
	}
	file.Close()
	return points, nil
}

func listmountedpartition() (map[string]*mountpoint, error) {
	points := make(map[string]*mountpoint, 0)
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

		if _, exist := FSSPEC_IGNORE[fsSpec]; exist {
			continue
		}
		if _, exist := FSTYPE_IGNORE[fsVfstype]; exist {
			continue
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

func (this *Fsinfo) getfsstate() error {
	if this.Disks == nil {
		this.Disks = make(map[string]*DiskStates)
	}
	points, err := listmountedpartition()
	if err != nil {
		return err
	}
	for _, p := range points {
		if strings.Contains(p.FsSpec, "/dev/") == false {
			continue
		}
		device := p.FsSpec[len("/dev/"):]
		if _, ok := this.Disks[device]; !ok {
			this.Disks[device] = &DiskStates{}
			this.Disks[device].Device = device
		}
		disk := this.Disks[device]
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
		disk.ReadRequests = statedisk.ReadRequests - disk.pre_ReadRequests
		disk.ReadBytes = statedisk.ReadBytes - disk.pre_ReadBytes
		disk.WriteRequests = statedisk.WriteRequests - disk.pre_WriteRequests
		disk.WriteBytes = statedisk.WriteBytes - disk.pre_WriteBytes

		disk.pre_ReadRequests = statedisk.ReadRequests
		disk.pre_ReadBytes = statedisk.ReadBytes
		disk.pre_WriteRequests = statedisk.WriteRequests
		disk.pre_WriteBytes = statedisk.WriteBytes
	}
	return nil
}
