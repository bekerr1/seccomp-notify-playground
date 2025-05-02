# -*- mode: ruby -*-
# vi: set ft=ruby :
#

NUM_WORKERS = 3 # Set number of worker nodes
VAGRANT_BOX = "ubuntu/jammy64"
APISERVER_ADVERTISE_ADDRESS = "172.16.56.10"

Vagrant.configure("2") do |config|
  config.vm.box = VAGRANT_BOX
  
  # Enable shared folder for all VMs
  config.vm.synced_folder "./shared", "/shared", create: true, mount_options: ["dmode=777", "fmode=666"]

  config.vm.define "control-plane" do |cp|
    cp.vm.hostname = "k8s-control-plane"
    cp.vm.network "private_network", ip: APISERVER_ADVERTISE_ADDRESS
    cp.vm.provider "virtualbox" do |vb|
      vb.memory = 4096
      vb.cpus = 2
    end
    cp.vm.provision "ansible" do |ansible|
      ansible.playbook = "playbooks/setup.yml"
      ansible.extra_vars = {
        node_role: "control-plane",
        apiserver_advertise_address: APISERVER_ADVERTISE_ADDRESS
      }
    end
  end

  (1..NUM_WORKERS).each do |i|
    config.vm.define "worker-#{i}" do |worker|
      worker.vm.hostname = "k8s-worker-#{i}"
      worker.vm.network "private_network", ip: "172.16.56.#{10 + i}"
      worker.vm.provider "virtualbox" do |vb|
        vb.memory = 2048
        vb.cpus = 2
      end
      worker.vm.provision "ansible" do |ansible|
        ansible.playbook = "playbooks/setup.yml"
        ansible.extra_vars = { node_role: "worker" }
      end
    end
  end
end
