# nvme setup

My dev lab looks like this:

 * kvm host  (ClearLinux)
 * target host (rocky 9)
 * initiator host (ubuntu)

# nvme Disk

On my kvm host, I created a disk called *nvme.img* as follows:
```
cd /vm-files
qemu-img create -f qcow2 nvme.img 10G
chown qemu:qemu /vm-files/nvme.img
```

# target Setup

Instructions on how to configure QEMU nvme devices with multiple namespaces is found here;
https://qemu-project.gitlab.io/qemu/system/devices/nvme.html

On the target host, I edited the VM XML configuration, adding the following at the end:
```xml
<qemu:commandline>
  <qemu:arg value='-drive'/>
  <qemu:arg value='file=/vm-files/nvme.img,format=raw,if=none,id=D22'/>
  <qemu:arg value='-device'/>
  <qemu:arg value='nvme,drive=D22,serial=1234'/>
</qemu:commandline>
```

```xml
  <qemu:commandline>
    <qemu:arg value="-device"/>
    <qemu:arg value="nvme,id=nvme-ctrl-0,serial=deadbeef"/>
    <qemu:arg value="-drive"/>
    <qemu:arg value="file=/vm-files/nvm-1.img,if=none,id=nvm-1"/>
    <qemu:arg value="-device"/>
    <qemu:arg value="nvme-ns,drive=nvm-1"/>
    <qemu:arg value="-drive"/>
    <qemu:arg value="file=/vm-files/nvm-2.img,if=none,id=nvm-2"/>
    <qemu:arg value="-device"/>
    <qemu:arg value="nvme-ns,drive=nvm-2"/>
  </qemu:commandline>
```

At the top of the XLM file, I added this to the first line:
```xml
<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
```

When I started up the VM, it conflicted with the Video device, so I added a 2nd video device, and removed the first video device to 
get around the conflicting pci address.

Verify the nvme device is working:
```bash
dmesg | grep nvme
dnf -y install nvme-cli smartmontools
nvme list
smartctl -i /dev/nvme0n1
```

For details see :
 * https://blog.frankenmichl.de/2018/02/13/add-nvme-device-to-vm/
 * https://blog.christophersmart.com/2019/12/18/kvm-guests-with-emulated-ssd-and-nvme-drives/
 * https://futurewei-cloud.github.io/ARM-Datacenter/qemu/nvme-of-tcp-vms/

# target configuration

Next, you need to configure the nvme/tcp device so it can be accessed via tcp:
```bash
modprobe nvmet
modprobe nvmet-tcp
lsmod | grep nvme
cd /sys/kernel/config/nvmet/subsystems
mkdir nvme-test-target
cd nvme-test-target/
echo 1 | sudo tee -a attr_allow_any_host > /dev/null
mkdir namespaces/1
cd namespaces/1
ls
nvme list
echo -n /dev/nvme0n1 |sudo tee -a device_path > /dev/null
echo 1|sudo tee -a enable > /dev/null
ip addr
mkdir /sys/kernel/config/nvmet/ports/1
cd /sys/kernel/config/nvmet/ports/1
echo 192.168.0.107 | sudo tee -a addr_traddr > /dev/null
echo tcp|sudo tee -a addr_trtype > /dev/null
echo 4420|sudo tee -a addr_trsvcid > /dev/null
echo ipv4|sudo tee -a addr_adrfam > /dev/null
ln -s /sys/kernel/config/nvmet/subsystems/nvme-test-target/ /sys/kernel/config/nvmet/ports/1/subsystems/nvme-test-target
dmesg | grep nvmet_tcp
```

After all that, you should see a line like this, assuming your target IP address is 192.168.0.107:
```
[ 1952.700233] nvmet_tcp: enabling port 1 (192.168.0.107:4420)
```

# initiator setup

```bash
modprobe nvme
modprobe nvme-tcp
apt install nvme-cli
nvme list
nvme discover -t tcp -a 192.168.0.107 -s 4420
nvme connect -t tcp -n nvme-test-target -a 192.168.0.107 -s 4420
nvme list
```

You should get output like this:
```bash
Node                  SN                   Model                                    Namespace Usage                      Format           FW Rev  
--------------------- -------------------- ---------------------------------------- --------- -------------------------- ---------------- --------
/dev/nvme0n1          646ec652826976d694ed Linux                                    1          10.74  GB /  10.74  GB    512   B +  0 B   5.14.0-3
```

# initiator test

Here is an example of using the remote nvme over tcp:
```bash
fdisk /dev/nvme0n1
mkfs.ext4 /dev/nvme0n1p1 
mkdir /remotenvme
mount /dev/nvme0n1p1 /remotenvme
df -h
ls -l /
echo "hello" > /remotenvme/hello.txt
ls -l /remotenvme/
```

# initiator detach

You detach the initiator from the nvme tcp device with:
```
nvme disconnect /dev/nvme0n1 -n nvme-test-target
```

# nvme namespaces

https://narasimhan-v.github.io/2020/06/12/Managing-NVMe-Namespaces.html
