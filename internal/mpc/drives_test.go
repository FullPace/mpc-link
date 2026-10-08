package mpc

import "testing"

// The MPC X of 2026-10-08: an SSD ("MPC SSD", exFAT, mounted by edisksd) with MPC Link's own sshfs mount inside it.
const mpcX = `ID=387 SRC=/dev/sda1 TARGET=/media/MPC\040SSD ROOT=/ OPTS=uhelper=edisksd
===
/dev/root / ext4 ro,relatime 0 0
devtmpfs /dev devtmpfs rw,relatime,size=1025912k,nr_inodes=188348,mode=755 0 0
proc /proc proc rw,relatime 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev 0 0
devpts /dev/pts devpts rw,relatime,gid=5,mode=620,ptmxmode=666 0 0
tmpfs /run tmpfs rw,nosuid,nodev,size=413028k,nr_inodes=819200,mode=755 0 0
cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate,memory_recursiveprot 0 0
tmpfs /etc/machine-id tmpfs ro,size=413028k,nr_inodes=819200,mode=755 0 0
debugfs /sys/kernel/debug debugfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev,nr_inodes=1048576 0 0
fusectl /sys/fs/fuse/connections fusectl rw,nosuid,nodev,noexec,relatime 0 0
configfs /sys/kernel/config configfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /media tmpfs rw 0 0
tmpfs /secure-media tmpfs rw 0 0
/dev/root /media/acvs-content ext4 ro,relatime 0 0
/dev/mmcblk0p7 /data ext4 rw,relatime 0 0
/dev/mmcblk0p7 /media/az01-internal ext4 rw,relatime 0 0
overlay /etc overlay rw,relatime,lowerdir=/etc,upperdir=/data/system/etc/overlay,workdir=/data/system/etc/.work,uuid=null 0 0
overlay /var overlay rw,relatime,lowerdir=/var,upperdir=/data/system/var/overlay,workdir=/data/system/var/.work,uuid=null 0 0
tmpfs /var/volatile tmpfs rw,relatime 0 0
overlay /srv overlay rw,relatime,lowerdir=/srv,upperdir=/var/volatile/srv,workdir=/var/volatile/.srv-work,uuid=on 0 0
/dev/mmcblk1p1 /sdcard ext4 rw,relatime 0 0
/dev/mmcblk1p1 /media/az01-internal-sd ext4 rw,relatime 0 0
/dev/sda1 /media/MPC\040SSD exfat rw,nosuid,nodev,noexec,relatime,nosymfollow,fmask=0022,dmask=0022,iocharset=utf8,errors=remount-ro 0 0
`

func names(d []Drive) []string {
	var out []string
	for _, x := range d {
		out = append(out, x.Name+"="+x.Path)
	}
	return out
}

func TestDrives(t *testing.T) {
	cases := []struct {
		name, in string
		want     []string
	}{
		{"MPC X with SSD", mpcX, []string{"MPC SSD=/media/MPC SSD", "Internal=" + internalBase}},
		{"SD card and USB stick", "ID=1 SRC=/dev/mmcblk2p1 TARGET=/media/MY\\040SD ROOT=/ OPTS=uhelper=edisksd\nID=2 SRC=/dev/sdb1 TARGET=/media/STICK ROOT=/ OPTS=uhelper=edisksd\n===\n" +
			"/dev/mmcblk2p1 /media/MY\\040SD vfat rw 0 0\n/dev/sdb1 /media/STICK exfat rw 0 0\n",
			[]string{"MY SD=/media/MY SD", "STICK=/media/STICK", "Internal=" + internalBase}},
		{"no helper table: other disks under /media", "===\n/dev/root /media/acvs-content ext4 ro 0 0\n/dev/mmcblk1p1 /media/az01-internal-sd ext4 rw 0 0\n" +
			"/dev/sda1 /media/Backup\\040Drive exfat rw 0 0\n:/ /media/Backup\\040Drive/Mac fuse.sshfs rw 0 0\n",
			[]string{"Backup Drive=/media/Backup Drive", "Internal=" + internalBase}},
		{"internal only", "===\n/dev/mmcblk1p1 /media/az01-internal-sd ext4 rw 0 0\n", []string{"Internal=" + internalBase}},
	}
	for _, c := range cases {
		got := names(parseDrives(c.in))
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v, want %v", c.name, got, c.want)
				break
			}
		}
	}
	d := parseDrives(mpcX)
	if p, ok := Pick(d, ""); !ok || p.Name != "MPC SSD" {
		t.Errorf("automatic: %v %v", p, ok)
	}
	if p, ok := Pick(d, InternalDrive); !ok || p.Path != internalBase {
		t.Errorf("internal: %v %v", p, ok)
	}
	if _, ok := Pick(d, "SD Card"); ok {
		t.Errorf("a missing drive must not be picked")
	}
}
