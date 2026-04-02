package mock

import (
	"fmt"
	"time"

	"github.com/mzaran/w9s/internal/dao"
)

// populateSampleData fills the mock client with realistic sample cluster data.
func populateSampleData(c *WarewulfClient) {
	populateProfiles(c)
	populateImages(c)
	populateOverlays(c)
	populateNodes(c)
	populatePowerStatus(c)
}

func populateProfiles(c *WarewulfClient) {
	c.profiles.data["default"] = &dao.WwProfile{
		Comment:        "Default compute node profile",
		ClusterName:    "mycluster",
		ImageName:      "rocky9",
		Ipxe:           "default",
		RuntimeOverlay: []string{"generic", "runtime"},
		SystemOverlay:  []string{"wwinit"},
		Kernel: &dao.KernelConf{
			Version: "5.14.0-362.el9.x86_64",
			Args:    []string{"quiet", "crashkernel=auto"},
		},
		Init: "/sbin/init",
		Root: "initramfs",
	}

	c.profiles.data["gpu"] = &dao.WwProfile{
		Comment:        "GPU compute node profile",
		ClusterName:    "mycluster",
		Profiles:       []string{"default"},
		ImageName:      "rocky9-cuda",
		RuntimeOverlay: []string{"generic", "runtime"},
		SystemOverlay:  []string{"wwinit"},
		Kernel: &dao.KernelConf{
			Version: "5.14.0-362.el9.x86_64",
			Args:    []string{"quiet", "crashkernel=auto", "nvidia.NVreg_EnableGpuFirmware=0"},
		},
		Init: "/sbin/init",
		Root: "initramfs",
	}
}

func populateImages(c *WarewulfClient) {
	c.images.data["rocky9"] = &dao.WwImage{
		Kernels:   []string{"5.14.0-362.el9.x86_64", "5.14.0-284.el9.x86_64"},
		Size:      2254857830, // ~2.1 GB
		BuildTime: time.Now().Add(-24 * time.Hour).Unix(),
		Writable:  false,
	}

	c.images.data["rocky9-cuda"] = &dao.WwImage{
		Kernels:   []string{"5.14.0-362.el9.x86_64"},
		Size:      3758096384, // ~3.5 GB
		BuildTime: time.Now().Add(-48 * time.Hour).Unix(),
		Writable:  false,
	}
}

func populateOverlays(c *WarewulfClient) {
	c.overlays.data["wwinit"] = &dao.WwOverlay{
		Files: []string{
			"/etc/hostname",
			"/etc/hosts",
			"/etc/resolv.conf",
			"/etc/sysconfig/network",
			"/etc/NetworkManager/system-connections/eth0.nmconnection",
			"/warewulf/init",
		},
		Site: false,
	}

	c.overlays.data["generic"] = &dao.WwOverlay{
		Files: []string{
			"/etc/passwd",
			"/etc/group",
			"/etc/shadow",
			"/root/.ssh/authorized_keys",
		},
		Site: false,
	}

	c.overlays.data["runtime"] = &dao.WwOverlay{
		Files: []string{
			"/etc/slurm/slurm.conf",
			"/etc/munge/munge.key",
			"/etc/exports",
			"/etc/fstab.ww",
		},
		Site: false,
	}

	c.overlays.data["chrony"] = &dao.WwOverlay{
		Files: []string{
			"/etc/chrony.conf",
		},
		Site: false,
	}
}

func populateNodes(c *WarewulfClient) {
	// 8 compute nodes: compute-01 through compute-08
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("compute-%02d", i)
		c.nodes.data[name] = &dao.WwNode{
			Discoverable: dao.WWBoolFalse,
			WwProfile: dao.WwProfile{
				Profiles:       []string{"default"},
				Comment:        fmt.Sprintf("Compute node %d", i),
				ClusterName:    "mycluster",
				ImageName:      "rocky9",
				RuntimeOverlay: []string{"generic", "runtime"},
				SystemOverlay:  []string{"wwinit"},
				Kernel: &dao.KernelConf{
					Version: "5.14.0-362.el9.x86_64",
				},
				PrimaryNetDev: "eth0",
				NetDevs: map[string]*dao.NetDev{
					"eth0": {
						Device:  "eth0",
						Ipaddr:  fmt.Sprintf("10.0.0.%d", i),
						Netmask: "255.255.255.0",
						Gateway: "10.0.0.254",
						Hwaddr:  fmt.Sprintf("00:16:3e:00:00:%02x", i),
						OnBoot:  dao.WWBoolTrue,
					},
				},
				Ipmi: &dao.IpmiConf{
					Ipaddr:   fmt.Sprintf("10.0.1.%d", i),
					Netmask:  "255.255.255.0",
					Gateway:  "10.0.1.254",
					UserName: "admin",
					Password: "password",
				},
			},
		}
		// Power status: all on except compute-07 which is off.
		if i == 7 {
			c.power.status[name] = "off"
		} else {
			c.power.status[name] = "on"
		}
	}

	// 2 GPU nodes: gpu-01 and gpu-02
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("gpu-%02d", i)
		c.nodes.data[name] = &dao.WwNode{
			Discoverable: dao.WWBoolFalse,
			WwProfile: dao.WwProfile{
				Profiles:       []string{"default", "gpu"},
				Comment:        fmt.Sprintf("GPU node %d", i),
				ClusterName:    "mycluster",
				ImageName:      "rocky9-cuda",
				RuntimeOverlay: []string{"generic", "runtime"},
				SystemOverlay:  []string{"wwinit"},
				Kernel: &dao.KernelConf{
					Version: "5.14.0-362.el9.x86_64",
					Args:    []string{"quiet", "nvidia.NVreg_EnableGpuFirmware=0"},
				},
				PrimaryNetDev: "eth0",
				NetDevs: map[string]*dao.NetDev{
					"eth0": {
						Device:  "eth0",
						Ipaddr:  fmt.Sprintf("10.0.0.%d", 8+i),
						Netmask: "255.255.255.0",
						Gateway: "10.0.0.254",
						Hwaddr:  fmt.Sprintf("00:16:3e:00:01:%02x", i),
						OnBoot:  dao.WWBoolTrue,
					},
				},
				Ipmi: &dao.IpmiConf{
					Ipaddr:   fmt.Sprintf("10.0.1.%d", 8+i),
					Netmask:  "255.255.255.0",
					Gateway:  "10.0.1.254",
					UserName: "admin",
					Password: "password",
				},
			},
		}
		c.power.status[name] = "on"
	}
}

func populatePowerStatus(_ *WarewulfClient) {
	// Power status is set inline during populateNodes.
}
