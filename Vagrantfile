# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

MACHINE_NAME = "gesclient"
MACHINE_MEMORY = 2048
MACHINE_CPU = 2
MACHINE_IP = "192.168.22.10"
MACHINE_TIMEZONE = "UTC"
MACHINE_HOSTNAME = "gesclient"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.ssh.forward_agent = true

  config.vm.define MACHINE_NAME do |kiosk|
    kiosk.vm.hostname = MACHINE_HOSTNAME

    kiosk.vm.provider :virtualbox do |vb|
      vb.name = MACHINE_NAME
      vb.memory = MACHINE_MEMORY
      vb.cpus = MACHINE_CPU
    end

    kiosk.vm.network :private_network, ip: MACHINE_IP

    kiosk.vm.synced_folder "#{ENV['GOPATH']}", "/home/vagrant/code",
      id: "gopath",
      :nfs => false,
      :mount_options => ['dmode=777','fmode=777'],
      :create => true

    ####
    # Update apt-get packages list
    ##########
    kiosk.vm.provision "shell", path: "provisioning/apt_get_update.sh"

    ####
    # Go and tools
    ##########

    # Git
    kiosk.vm.provision "shell", path: "provisioning/install_git.sh", privileged: false

    # Go
    kiosk.vm.provision "shell", path: "provisioning/install_go.sh", privileged: false

    ####
    # EventStore
    ##########

    # Provision EventStore
    kiosk.vm.provision "shell", path: "provisioning/install_eventstore.sh", args: [MACHINE_IP, MACHINE_TIMEZONE]

  end
end